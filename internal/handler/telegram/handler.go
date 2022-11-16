package telegram

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sku4/ozon-route256-spending-bot/internal/service"
)

type IHandler interface {
	IncomingMessage(context.Context, tgbotapi.Update) error
}

type Func func(context.Context, tgbotapi.Update) error

func (f Func) IncomingMessage(ctx context.Context, update tgbotapi.Update) error {
	return f(ctx, update)
}

type Handler struct {
	services service.Service
}

func NewHandler(services *service.Service) IHandler {
	return &Handler{
		services: *services,
	}
}
