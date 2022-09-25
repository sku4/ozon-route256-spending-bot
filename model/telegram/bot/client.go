package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/log"
	"go.uber.org/zap"
)

type Server struct {
	client *tgbotapi.BotAPI
	ctx    context.Context
	logger *zap.Logger
}

func New(ctx context.Context, token string) (*Server, error) {
	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "error init spending bot")
	}
	logger := log.LoggerFromContext(ctx)

	return &Server{
		ctx:    ctx,
		client: client,
		logger: logger,
	}, nil
}

func (s *Server) SendMessage(text string, userID int64) error {
	_, err := s.client.Send(tgbotapi.NewMessage(userID, text))
	if err != nil {
		return errors.Wrap(err, "bot.Send")
	}
	return nil
}
