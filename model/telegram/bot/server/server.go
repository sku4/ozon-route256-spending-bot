package server

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
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

func (s *Server) Run(h *handler.Handler) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := s.client.GetUpdatesChan(u)

	logger.Info("Listening for messages")

	for update := range updates {
		if update.Message != nil {
			logger.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
		}
		if update.CallbackQuery != nil {
			logger.Infof("[%s] %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
		}

		err := h.IncomingMessage(update)
		if err != nil {
			logger.SInfo("error processing message: ", err)
		}
	}
	return nil
}
