package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func (h *Handler) IncomingMessage(update tgbotapi.Update) (err error) {
	h.services.Middleware.UpdateRates(h.ctx)
	ctx, err := h.services.Middleware.DefineUser(h.ctx, update)
	if err != nil {
		return
	}

	if update.Message != nil {
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start", "help":
				err = h.services.Spending.Start(ctx, update)
			case "categories":
				err = h.services.Spending.Categories(ctx, update)
			case "categoryadd":
				err = h.services.Spending.CategoryAdd(ctx, update)
			case "spendingadd":
				err = h.services.Spending.SpendingAdd(ctx, update)
			case "report7":
				err = h.services.Spending.Report7(ctx, update)
			case "report31":
				err = h.services.Spending.Report31(ctx, update)
			case "report365":
				err = h.services.Spending.Report365(ctx, update)
			case "currency":
				err = h.services.Spending.Currency(ctx, update)
			default:
				err = h.services.Spending.NotFound(ctx, update)
			}
		}
	} else if update.CallbackQuery != nil {
		if strings.Index(update.CallbackQuery.Data, "categories") == 0 {
			err = h.services.Spending.CategoriesQuery(ctx, update)
		} else if strings.Index(update.CallbackQuery.Data, "spendingadd") == 0 {
			err = h.services.Spending.SpendingAddQuery(ctx, update)
		} else if strings.Index(update.CallbackQuery.Data, "currency") == 0 {
			err = h.services.Spending.CurrencyQuery(ctx, update)
		}
	}

	return
}
