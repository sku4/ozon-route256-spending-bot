package service

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service/spending"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go

type Spending interface {
	Start(context.Context, tgbotapi.Update) error
}

type Service struct {
	Spending
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Spending: spending.NewService(repos.Spending),
	}
}
