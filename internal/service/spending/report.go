package spending

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/kafka"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
	apiReport "gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api/report"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

//go:generate mockgen -source=report.go -destination=mocks/report.go

type Report struct {
	reposSpend repository.Spending
	reposCat   repository.Categories
	rates      rates.Client
	grpcClient api.SpendingClient
}

func NewReport(reposSpending repository.Spending, reposCategories repository.Categories, rates rates.Client,
	grpcClient api.SpendingClient) *Report {
	return &Report{
		reposSpend: reposSpending,
		reposCat:   reposCategories,
		rates:      rates,
		grpcClient: grpcClient,
	}
}

func (r *Report) Build(ctx context.Context, f1, f2 time.Time, userCurr model.Currency, chatId int64) (err error) {
	report := ""
	m, err := r.reposSpend.Report(ctx, f1, f2, r.rates, userCurr)
	if err != nil {
		return
	}

	userCurrAbbr := userCurr.Abbr

	categories, err := r.reposCat.Categories(ctx)
	if err != nil {
		return errors.Wrap(err, "report categories")
	}
	for _, category := range categories {
		if sum, ok := m[category.Id]; ok {
			report += fmt.Sprintf("_%s_ - %.2f %s\n", category.Title, sum, userCurrAbbr)
		}
	}

	_, err = r.grpcClient.SendReport(ctx, &apiReport.Report{
		F1:     timestamppb.New(f1),
		F2:     timestamppb.New(f2),
		ChatId: chatId,
		UserCurrency: &apiReport.Currency{
			Id:   int64(userCurr.Id),
			Abbr: userCurr.Abbr,
		},
		Message: report,
	})
	if err != nil {
		return errors.Wrap(err, "could not greet build")
	}

	logger.Infos("report success sent:", report, f1, f2)

	return
}

func (s *Service) Report7(ctx context.Context, update tgbotapi.Update) (err error) {
	now := time.Now()
	f1 := now.UTC()
	if now.Weekday() < 1 {
		f1 = now.AddDate(0, 0, -6)
	} else if now.Weekday() > 1 {
		f1 = now.AddDate(0, 0, int(-now.Weekday())+1)
	}
	f1 = time.Date(f1.Year(), f1.Month(), f1.Day(), 0, 0, 0, 0, f1.Location())
	f2 := f1.AddDate(0, 0, 6)
	f2 = time.Date(f2.Year(), f2.Month(), f2.Day(), 23, 59, 59, 0, f2.Location())

	err = s.buildReport(ctx, update, f1, f2)
	if err != nil {
		return errors.Wrap(err, "build report 7")
	}

	return
}

func (s *Service) Report31(ctx context.Context, update tgbotapi.Update) (err error) {
	now := time.Now()
	f1 := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	f2 := time.Date(f1.Year(), f1.Month()+1, 0, 23, 59, 59, 0, f1.Location())

	err = s.buildReport(ctx, update, f1, f2)
	if err != nil {
		return errors.Wrap(err, "build report 31")
	}

	return
}

func (s *Service) Report365(ctx context.Context, update tgbotapi.Update) (err error) {
	now := time.Now()
	f1 := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	f2 := f1.AddDate(1, 0, -1)
	f2 = time.Date(f2.Year(), f2.Month(), f2.Day(), 23, 59, 59, f2.Nanosecond(), f2.Location())

	err = s.buildReport(ctx, update, f1, f2)
	if err != nil {
		return errors.Wrap(err, "build report 365")
	}

	return
}

func (s *Service) buildReport(ctx context.Context, update tgbotapi.Update, f1, f2 time.Time) error {
	userCtx, err := user.FromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "buildReport")
	}
	userState, err := userCtx.GetState(ctx)
	if err != nil {
		return errors.Wrap(err, "buildReport")
	}
	userCurrency, err := userState.GetCurrency(ctx)
	if err != nil {
		return errors.Wrap(err, "buildReport")
	}

	reportJson, err := json.Marshal(kafka.Report{
		F1:       f1,
		F2:       f2,
		ChatId:   update.Message.Chat.ID,
		UserCurr: userCurrency,
	})
	if err != nil {
		return err
	}

	msgReport := sarama.ProducerMessage{
		Topic: kafka.TopicReport,
		Key:   sarama.StringEncoder("report"),
		Value: sarama.StringEncoder(reportJson),
	}

	s.kafkaProducer.Input() <- &msgReport

	return nil
}

func (s *Service) SendReport(ctx context.Context, report string, f1, f2 time.Time, chatId int64) (err error) {
	_ = ctx

	r := ""
	f := "2 Jan 06"
	if report == "" {
		r = fmt.Sprintf("Report by week (*%s - %s*): spending not found", f1.Format(f), f2.Format(f))
	} else {
		r = fmt.Sprintf("Report by week (*%s - %s*):\n", f1.Format(f), f2.Format(f)) + report
	}

	err = s.client.SendMessage(r, chatId)
	if err != nil {
		return err
	}

	logger.Infos("report sent to telegram:", r, chatId)

	return nil
}
