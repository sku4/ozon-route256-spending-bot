package service

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service/spending"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go

type Spending interface {
	Start(context.Context, tgbotapi.Update) error
	NotFound(context.Context, tgbotapi.Update) error
	SpendingAdd(context.Context, tgbotapi.Update) error
	SpendingAddQuery(context.Context, tgbotapi.Update) error
	Categories
}

type Categories interface {
	Categories(context.Context, tgbotapi.Update) error
	CategoryAdd(context.Context, tgbotapi.Update) error
	CategoriesQuery(context.Context, tgbotapi.Update) error
}

type Service struct {
	Spending
}

func NewService(repos *repository.Repository, client *client.Client) *Service {
	return &Service{
		Spending: spending.NewService(repos.Spending, client),
	}
}
