package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
)

func (s *Server) Run(h *handler.Handler) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := s.client.GetUpdatesChan(u)

	sugar := s.logger.Sugar()
	sugar.Info("listening for messages")

	for update := range updates {
		if update.Message == nil {
			continue
		}
		sugar.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)

		err := h.IncomingMessage(update)
		if err != nil {
			sugar.Info("error processing message:", err)
		}
	}
	return nil
}
