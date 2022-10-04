package spending

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service/middleware"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"strconv"
)

//go:generate mockgen -source=currency.go -destination=mocks/currency.go

const currencyPrefix = "currency_"

func (s *Service) Currency(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	if !s.rates.IsLoaded(ctx) {
		_ = s.client.SendMessage("Rates not loaded, please repeat later", update.Message.Chat.ID)
		return errors.New("rates still not loaded")
	}

	var inlineKeyboardRows []*client.KeyboardRow
	inlineKeyboardRow := client.NewKeyboardRow()
	for id := currency.USD; id <= currency.RUB; id++ {
		abbr := currency.CurrAbbr[currency.Currency(id)]
		inlineKeyboardRow.Add(abbr, currencyPrefix+strconv.Itoa(int(id)))
	}
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)

	err = s.client.SendInlineKeyboard(inlineKeyboardRows,
		fmt.Sprint("Change currency:"), update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) CurrencyQuery(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	var inlineKeyboardRows []*client.KeyboardRow

	postfixCur, err := strconv.Atoi(update.CallbackQuery.Data[len(currencyPrefix):])
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Error currency not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "currency user not convert")
	}

	userCurrency := currency.Currency(postfixCur)
	userCurrAbbr := ""
	for a, c := range currency.AbbrCurr {
		if c == userCurrency {
			userCurrAbbr = a
			break
		}
	}

	if userCurrency > -1 {
		// change currency
		u, err := middleware.UserFromContext(ctx)
		if err != nil {
			_ = s.client.SendMessage(fmt.Sprintf(
				"User not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
			return errors.Wrap(err, "user not found")
		}

		u.State.Currency = userCurrency

		err = s.client.SendCallbackQuery(inlineKeyboardRows, fmt.Sprintf(
			"Currency success changed to *%s*\r\n"+
				"Show /report7 /report31 /report365", userCurrAbbr),
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
	}

	return
}
