package spending

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"github.com/sku4/ozon-route256-spending-bot/model/telegram/bot/client"
	"github.com/sku4/ozon-route256-spending-bot/pkg/decimal"
	"github.com/sku4/ozon-route256-spending-bot/pkg/user"
	"strconv"
)

//go:generate mockgen -source=limit.go -destination=mocks/limit.go

const limitPrefix = "limit_"

func (s *Service) LimitAdd(ctx context.Context, update tgbotapi.Update) (err error) {
	if !s.rates.IsLoaded(ctx) {
		_ = s.client.SendMessage("Rates not loaded, please repeat later", update.Message.Chat.ID)
		return errors.New("rates still not loaded")
	}

	priceArg := update.Message.CommandArguments()
	priceLimit, err := strconv.ParseFloat(priceArg, 64)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Error convert price '*%s*'", priceArg), update.Message.Chat.ID)
		return errors.Wrap(err, "convert price")
	}
	if priceLimit <= 0 {
		_ = s.client.SendMessage("Please set price over 0", update.Message.Chat.ID)
		return errors.New("Price less than 0")
	}

	userCtx, err := user.FromContext(ctx)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"User not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "user not found")
	}
	uState, err := userCtx.GetState(ctx)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"State not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "state not found")
	}
	uCurrency, err := uState.GetCurrency(ctx)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Currency not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "currency not found")
	}
	userCurrAbbr := uCurrency.Abbr

	var inlineKeyboardRows []*client.KeyboardRow
	inlineKeyboardRow := client.NewKeyboardRow()
	event := NewEvent(priceLimit)
	cats, err := s.reposCat.Categories(ctx)
	if err != nil {
		return errors.New("limit add categories")
	}
	if len(cats) == 0 {
		_ = s.client.SendMessage("Categories list is empty, please add /categories", update.Message.Chat.ID)
		return errors.New("Categories list is empty")
	}
	for _, c := range cats {
		event.CategoryId = c.Id
		eventSer := EventSerialize(event)
		inlineKeyboardRow.Add(c.Title, limitPrefix+string(eventSer))
	}
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)

	err = s.client.SendInlineKeyboard(inlineKeyboardRows,
		fmt.Sprintf("Choose category limit (*%.2f %s*):", priceLimit, userCurrAbbr), update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) LimitQuery(ctx context.Context, update tgbotapi.Update) (err error) {
	if !s.rates.IsLoaded(ctx) {
		_ = s.client.SendMessage("Rates not loaded, please repeat later", update.Message.Chat.ID)
		return errors.New("rates still not loaded")
	}

	var inlineKeyboardRows []*client.KeyboardRow

	eventSer := update.CallbackQuery.Data[len(limitPrefix):]
	event, err := eventUnserialize(eventSer)
	if err != nil {
		return errors.Wrap(err, "event unserialize")
	}

	catSelected, err := s.reposCat.CategoryGetById(ctx, event.CategoryId)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Category not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "Category not found")
	}

	userCtx, err := user.FromContext(ctx)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"User not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "user not found")
	}
	uState, err := userCtx.GetState(ctx)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"State not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "state not found")
	}
	uCurrency, err := uState.GetCurrency(ctx)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Currency not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "currency not found")
	}
	userCurrAbbr := uCurrency.Abbr

	if catSelected.Id > -1 {
		price, err := s.ConvertPrice(ctx, decimal.ToDecimal(event.Price))
		if err != nil {
			return errors.Wrap(err, "limit convert price")
		}
		err = uState.AddLimit(ctx, catSelected.Id, price)
		if err != nil {
			_ = s.client.SendMessage(fmt.Sprintf(
				"Limit not add: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
			return errors.Wrap(err, "limit not add")
		}

		err = s.client.SendCallbackQuery(inlineKeyboardRows, fmt.Sprintf(
			"Limit *%.2f %s* for category *%s* success added", event.Price, userCurrAbbr, catSelected.Title),
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
		if err != nil {
			return err
		}
	}

	return
}
