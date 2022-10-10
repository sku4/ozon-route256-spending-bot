package spending

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
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
	for _, c := range s.reposCurr.All() {
		inlineKeyboardRow.Add(c.Abbr, currencyPrefix+strconv.Itoa(c.Id))
	}
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)

	err = s.client.SendInlineKeyboard(inlineKeyboardRows, "Change currency:", update.Message.Chat.ID)
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

	userCurrency, err := s.reposCurr.GetById(postfixCur)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Currency not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "Currency not found")
	}
	userCurrAbbr := userCurrency.Abbr

	if userCurrency.Id > -1 {
		// change currency
		u, err := user.FromContext(ctx)
		if err != nil {
			_ = s.client.SendMessage(fmt.Sprintf(
				"User not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
			return errors.Wrap(err, "user not found")
		}

		uState, err := u.GetState()
		if err != nil {
			return err
		}
		err = uState.SetCurrency(userCurrency)
		if err != nil {
			return err
		}

		err = s.client.SendCallbackQuery(inlineKeyboardRows, fmt.Sprintf(
			"Currency success changed to *%s*\r\n"+
				"Show /report7 /report31 /report365", userCurrAbbr),
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
		if err != nil {
			return err
		}
	}

	return
}
