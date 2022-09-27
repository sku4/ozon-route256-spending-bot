package spending

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/spending"
	"time"
)

//go:generate mockgen -source=report.go -destination=mocks/report.go

func (s *Service) Report7(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

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

	report := buildReport(ctx, s.repos, f1, f2)
	r := ""
	f := "2 Jan 06"
	if report == "" {
		r = fmt.Sprintf("Report by week (*%s - %s*): spending not found", f1.Format(f), f2.Format(f))
	} else {
		r = fmt.Sprintf("Report by week (*%s - %s*):\n", f1.Format(f), f2.Format(f)) + report
	}

	err = s.client.SendMessage(r, update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) Report31(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	now := time.Now()
	f1 := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	f2 := time.Date(f1.Year(), f1.Month()+1, 0, 23, 59, 59, 0, f1.Location())

	report := buildReport(ctx, s.repos, f1, f2)
	r := ""
	f := "2 Jan 06"
	if report == "" {
		r = fmt.Sprintf("Report by month (*%s - %s*): spending not found", f1.Format(f), f2.Format(f))
	} else {
		r = fmt.Sprintf("Report by month (*%s - %s*):\n", f1.Format(f), f2.Format(f)) + report
	}

	err = s.client.SendMessage(r, update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) Report365(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	now := time.Now()
	f1 := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	f2 := f1.AddDate(1, 0, -1)
	f2 = time.Date(f2.Year(), f2.Month(), f2.Day(), 23, 59, 59, f2.Nanosecond(), f2.Location())

	report := buildReport(ctx, s.repos, f1, f2)
	r := ""
	f := "2 Jan 06"
	if report == "" {
		r = fmt.Sprintf("Report by year (*%s - %s*): spending not found", f1.Format(f), f2.Format(f))
	} else {
		r = fmt.Sprintf("Report by year (*%s - %s*):\n", f1.Format(f), f2.Format(f)) + report
	}

	err = s.client.SendMessage(r, update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func buildReport(ctx context.Context, repos repository.Spending, f1, f2 time.Time) string {
	report := ""
	m := make(map[int]float64)
	events := repos.Events(ctx)
	for _, event := range events {
		if (event.Date.After(f1) || event.Date.Equal(f1)) && (event.Date.Before(f2) || event.Date.Equal(f2)) {
			m[event.Category.Id] += event.Price
		}
	}

	categories := repos.Categories(ctx)
	c := make(map[int]spending.Category)
	for _, category := range categories {
		c[category.Id] = category
	}

	for categoryId, sum := range m {
		if category, ok := c[categoryId]; ok {
			report += fmt.Sprintf("_%s_ - %.1f\n", category.Title, sum)
		}
	}

	return report
}
