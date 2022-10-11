package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"strings"
)

func (h *Handler) IncomingMessage(update tgbotapi.Update) (err error) {
	run := h.services.Middleware.UpdateRatesSync(h.ctx)
	ctx, err := h.services.Middleware.DefineUser(h.ctx, update)
	if err != nil {
		return errors.Wrap(err, "incoming message define user")
	}

	if run {
		req := requiredRates(update)
		if req {
			// wait until rates will update
			err = <-h.services.Middleware.RatesSyncChan(ctx)
			if err != nil {
				return errors.Wrap(err, "incoming message rates sync")
			}
		} else {
			go func() {
				// read channel error
				<-h.services.Middleware.RatesSyncChan(ctx)
			}()
		}
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
			case "limit":
				err = h.services.Spending.LimitAdd(ctx, update)
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
		} else if strings.Index(update.CallbackQuery.Data, "limit") == 0 {
			err = h.services.Spending.LimitQuery(ctx, update)
		}
	}

	return
}

func requiredRates(update tgbotapi.Update) (req bool) {
	if update.Message != nil {
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "spendingadd",
				"report7",
				"report31",
				"report365",
				"currency",
				"limit":
				req = true
			}
		}
	} else if update.CallbackQuery != nil {
		if strings.Index(update.CallbackQuery.Data, "spendingadd") == 0 ||
			strings.Index(update.CallbackQuery.Data, "currency") == 0 ||
			strings.Index(update.CallbackQuery.Data, "limit") == 0 {
			req = true
		}
	}

	return
}
