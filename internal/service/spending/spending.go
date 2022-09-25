package spending

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
)

//go:generate mockgen -source=spending.go -destination=mocks/spending.go

type Service struct {
	spending repository.Spending
}

func NewService(spending repository.Spending) *Service {
	return &Service{
		spending: spending,
	}
}

func (s *Service) Start(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx
	_ = update

	return
}
