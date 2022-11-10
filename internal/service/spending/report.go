package spending

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/kafka"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
	"time"
)

//go:generate mockgen -source=report.go -destination=mocks/report.go

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
	successMsg := <-s.kafkaProducer.Successes()
	logger.Infos("Successful to write message, offset:", successMsg.Offset)

	return nil
}
