package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (h *Handler) IncomingMessage(update tgbotapi.Update) error {
	err := h.services.Spending.Start(h.ctx, update)
	if err != nil {
		return err
	}

	return nil
}
