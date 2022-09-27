package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func (h *Handler) IncomingMessage(update tgbotapi.Update) (err error) {
	if update.Message != nil {
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				err = h.services.Spending.Start(h.ctx, update)
			case "categories":
				err = h.services.Spending.Categories(h.ctx, update)
			case "categoryadd":
				err = h.services.Spending.CategoryAdd(h.ctx, update)
			case "spendingadd":
				err = h.services.Spending.SpendingAdd(h.ctx, update)
			default:
				err = h.services.Spending.NotFound(h.ctx, update)
			}
		}
	} else if update.CallbackQuery != nil {
		if strings.Index(update.CallbackQuery.Data, "categories") == 0 {
			err = h.services.Spending.CategoriesQuery(h.ctx, update)
		} else if strings.Index(update.CallbackQuery.Data, "spendingadd") == 0 {
			err = h.services.Spending.SpendingAddQuery(h.ctx, update)
		}
	}

	return
}
