package server

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sku4/ozon-route256-spending-bot/internal/handler/telegram"
	"github.com/sku4/ozon-route256-spending-bot/pkg/logger"
)

type Server struct {
	client *tgbotapi.BotAPI
	ctx    context.Context
}

func NewServer(ctx context.Context, client *tgbotapi.BotAPI) (*Server, error) {
	return &Server{
		ctx:    ctx,
		client: client,
	}, nil
}

func (s *Server) Run(ctx context.Context, h telegram.IHandler) error {
	h = MetricsMiddleware(h)
	h = TracingMiddleware(h)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := s.client.GetUpdatesChan(u)

	logger.Info("Listening for messages")

	for update := range updates {
		if update.Message != nil {
			logger.Infos(update.Message.From.UserName, update.Message.Text)
		}
		if update.CallbackQuery != nil {
			logger.Infos(update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
		}

		err := h.IncomingMessage(ctx, update)
		if err != nil {
			logger.Infos("error processing message: ", err)
		}
	}
	return nil
}
