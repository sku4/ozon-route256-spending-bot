package service

import (
	"context"
	"github.com/Shopify/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service/middleware"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service/spending"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"time"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go

type Spending interface {
	Start(context.Context, tgbotapi.Update) error
	ErrorMessage(context.Context, tgbotapi.Update, string) error
	NotFound(context.Context, tgbotapi.Update) error
	SpendingAdd(context.Context, tgbotapi.Update) error
	SpendingAddQuery(context.Context, tgbotapi.Update) error
	Categories
	Report
	Currency
	CategoryLimit
}

type Categories interface {
	Categories(context.Context, tgbotapi.Update) error
	CategoryAdd(context.Context, tgbotapi.Update) error
	CategoriesQuery(context.Context, tgbotapi.Update) error
}

type CategoryLimit interface {
	LimitAdd(context.Context, tgbotapi.Update) error
	LimitQuery(context.Context, tgbotapi.Update) error
}

type Report interface {
	Report7(context.Context, tgbotapi.Update) error
	Report31(context.Context, tgbotapi.Update) error
	Report365(context.Context, tgbotapi.Update) error
	SendReport(context.Context, string, time.Time, time.Time, int64) error
}

type Middleware interface {
	DefineUser(context.Context, tgbotapi.Update) (context.Context, error)
	UpdateRatesSync(context.Context) bool
	RatesSyncChan(context.Context) <-chan error
}

type Currency interface {
	Currency(context.Context, tgbotapi.Update) error
	CurrencyQuery(context.Context, tgbotapi.Update) error
}

type Service struct {
	Spending
	Middleware
}

func NewService(repos *repository.Repository, client client.BotClient, rates rates.Client, kafkaProducer sarama.AsyncProducer) *Service {
	return &Service{
		Spending:   spending.NewService(repos.Spending, repos.Categories, repos.CurrencyClient, client, rates, kafkaProducer),
		Middleware: middleware.NewMiddleware(repos.Users, client, rates),
	}
}
