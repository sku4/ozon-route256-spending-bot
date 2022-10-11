package spending

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
	"strconv"
	"strings"
	"time"
)

//go:generate mockgen -source=spending.go -destination=mocks/spending.go

type Service struct {
	reposSpend repository.Spending
	reposCurr  currency.Client
	reposCat   repository.Categories
	client     client.BotClient
	rates      rates.Client
}

type Event struct {
	Price         float64 `json:"Price"`
	CategoryId    int     `json:"CategoryId"`
	D             int     `json:"D"`
	M             int     `json:"M"`
	Y             int     `json:"Y"`
	Today         bool    `json:"Today"`
	SelectedToday bool    `json:"SelectedToday"`
}

func NewEvent(price float64) *Event {
	return &Event{
		Price:      price,
		D:          -1,
		M:          -1,
		Y:          -1,
		CategoryId: -1,
	}
}

func NewService(reposSpending repository.Spending, reposCategories repository.Categories, reposCurrencies currency.Client, client client.BotClient,
	rates rates.Client) *Service {
	return &Service{
		reposCat:   reposCategories,
		reposSpend: reposSpending,
		reposCurr:  reposCurrencies,
		client:     client,
		rates:      rates,
	}
}

const spendingAddPrefix = "spendingadd_"

func (s *Service) Start(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	msg := "Available commands:\n" +
		"/categories\n" +
		"`/categoryadd Food` _- where Food is category name_\n" +
		"`/spendingadd 100` _- where 100 is price_\n" +
		"/report7 _- report by current week_\n" +
		"/report31 _- report by current month_\n" +
		"/report365 _- report by current year_\n" +
		"/currency _- change currency_\n" +
		"`/limit 100` _- limit category by sum spending on month_"
	err = s.client.SendMessage(msg, update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) ErrorMessage(ctx context.Context, update tgbotapi.Update, message string) (err error) {
	_ = ctx

	err = s.client.SendMessage(message, update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) NotFound(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	err = s.client.SendMessage("Command not found", update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) SpendingAdd(ctx context.Context, update tgbotapi.Update) (err error) {
	if !s.rates.IsLoaded(ctx) {
		_ = s.client.SendMessage("Rates not loaded, please repeat later", update.Message.Chat.ID)
		return errors.New("rates still not loaded")
	}

	priceArg := update.Message.CommandArguments()
	price, err := strconv.ParseFloat(priceArg, 64)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Error convert price '*%s*'", priceArg), update.Message.Chat.ID)
		return errors.Wrap(err, "convert price")
	}
	if price <= 0 {
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
	event := NewEvent(price)
	categories, err := s.reposCat.Categories(ctx)
	if err != nil {
		return errors.Wrap(err, "event add categories")
	}
	if len(categories) == 0 {
		_ = s.client.SendMessage("Categories list is empty, please add /categories", update.Message.Chat.ID)
		return errors.New("Categories list is empty")
	}
	for _, category := range categories {
		event.CategoryId = category.Id
		eventSer := eventSerialize(event)
		inlineKeyboardRow.Add(category.Title, spendingAddPrefix+string(eventSer))
	}
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)

	err = s.client.SendInlineKeyboard(inlineKeyboardRows,
		fmt.Sprintf("Choose category (*%.2f %s*):", price, userCurrAbbr), update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) SpendingAddQuery(ctx context.Context, update tgbotapi.Update) (err error) {
	if !s.rates.IsLoaded(ctx) {
		_ = s.client.SendMessage("Rates not loaded, please repeat later", update.Message.Chat.ID)
		return errors.New("rates still not loaded")
	}

	var inlineKeyboardRows []*client.KeyboardRow
	inlineKeyboardRow := client.NewKeyboardRow()

	eventSer := update.CallbackQuery.Data[len(spendingAddPrefix):]
	event, err := eventUnserialize(eventSer)
	if err != nil {
		return errors.Wrap(err, "event unserialize")
	}

	var category model.Category
	if event.CategoryId > -1 {
		cs, err := s.reposCat.Categories(ctx)
		if err != nil {
			return errors.Wrap(err, "event add categories")
		}
		for _, c := range cs {
			if c.Id == event.CategoryId {
				category = c
				break
			}
		}
	}

	userCtx, err := user.FromContext(ctx)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"User not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
		return errors.Wrap(err, "user not found")
	}
	uState, err := userCtx.GetState(ctx)
	if err != nil {
		return errors.Wrap(err, "state not found")
	}
	uCurrency, err := uState.GetCurrency(ctx)
	if err != nil {
		return errors.Wrap(err, "currency not found")
	}
	userCurrAbbr := uCurrency.Abbr

	now := time.Now().UTC()
	if event.D > -1 {
		// add event
		userRate, ok := s.rates.GetRate(ctx, uCurrency)
		if !ok {
			_ = s.client.SendMessage(fmt.Sprintf(
				"Rate not found: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
			return errors.Wrap(err, "rate not found")
		}
		userRateDecimal := userRate.Rate
		t := time.Date(event.Y, time.Month(event.M), event.D, 0, 0, 0, 0, now.Location())
		_, err = s.reposSpend.AddEvent(ctx, event.CategoryId, t, decimal.ToDecimal(event.Price).Multiply(userRateDecimal))
		if err != nil {
			_ = s.client.SendMessage(fmt.Sprintf(
				"Error add event: %s", err.Error()), update.CallbackQuery.Message.Chat.ID)
			return errors.Wrap(err, "add event")
		}
		err = s.client.SendCallbackQuery(inlineKeyboardRows, fmt.Sprintf(
			"Event with price *%v %s* on *%s* success added to *%s*\r\n"+
				"Show /report7 /report31 /report365", event.Price, userCurrAbbr, t.Format("2 Jan 06"), category.Title),
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)

		// notification by limit category
		mess, err := s.checkLimitPrice(ctx, category)
		if err != nil {
			return errors.Wrap(err, "add event check limit price")
		}
		_ = s.client.SendMessage(mess, update.CallbackQuery.Message.Chat.ID)
	} else if event.M > -1 {
		// show days
		firstMonth := time.Date(2006, time.Month(event.M), 1, 0, 0, 0, 0, time.Local)
		msg := fmt.Sprintf("Choose days (*%.2f %s* > *%s* > *%d* > *%s*):",
			event.Price, userCurrAbbr, category.Title, event.Y, firstMonth.Format("Jan"))
		t := time.Date(event.Y, time.Month(event.M)+1, 0, 0, 0, 0, 0, time.Local)
		countDays := t.Day()
		for i := 1; i <= countDays; i++ {
			event.D = i
			eventSer = eventSerialize(event)
			inlineKeyboardRow.Add(strconv.Itoa(i), spendingAddPrefix+string(eventSer))
		}

		event.D = -1
		event.M = -1
		eventSer = eventSerialize(event)
		inlineKeyboardRow2 := client.NewKeyboardRow()
		inlineKeyboardRow2.Add("<< Back", spendingAddPrefix+string(eventSer))
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow, inlineKeyboardRow2)
		err = s.client.SendCallbackQuery(inlineKeyboardRows, msg,
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
	} else if event.Y > -1 {
		// show months
		firstMonth := time.Date(2006, 1, 1, 0, 0, 0, 0, time.Local)
		msg := fmt.Sprintf("Choose months (*%.2f %s* > *%s* > *%d*):",
			event.Price, userCurrAbbr, category.Title, event.Y)
		for i := 1; i <= 12; i++ {
			m := firstMonth.Format("Jan")
			firstMonth = firstMonth.AddDate(0, 1, 0)
			event.M = i
			eventSer = eventSerialize(event)
			inlineKeyboardRow.Add(m, spendingAddPrefix+string(eventSer))
			if i == 6 {
				inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
				inlineKeyboardRow = client.NewKeyboardRow()
			}
		}

		event.M = -1
		event.Y = -1
		eventSer = eventSerialize(event)
		inlineKeyboardRow2 := client.NewKeyboardRow()
		inlineKeyboardRow2.Add("<< Back", spendingAddPrefix+string(eventSer))
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow, inlineKeyboardRow2)
		err = s.client.SendCallbackQuery(inlineKeyboardRows, msg,
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
	} else if event.SelectedToday {
		// show years
		msg := fmt.Sprintf("Choose years (*%.2f %s* > *%s*):", event.Price, userCurrAbbr, category.Title)
		years := []int{now.Year() - 1, now.Year(), now.Year() + 1}
		for _, year := range years {
			event.Y = year
			eventSer = eventSerialize(event)
			inlineKeyboardRow.Add(strconv.Itoa(year), spendingAddPrefix+string(eventSer))
		}

		event.Y = -1
		event.SelectedToday = false
		eventSer = eventSerialize(event)
		inlineKeyboardRow2 := client.NewKeyboardRow()
		inlineKeyboardRow2.Add("<< Back", spendingAddPrefix+string(eventSer))
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow, inlineKeyboardRow2)
		err = s.client.SendCallbackQuery(inlineKeyboardRows, msg,
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
	} else if event.CategoryId > -1 {
		// choose date
		msg := fmt.Sprintf("Choose date (*%.2f %s* > *%s*):", event.Price, userCurrAbbr, category.Title)
		event.Today = true
		event.SelectedToday = true
		event.D = now.Day()
		event.M = int(now.Month())
		event.Y = now.Year()
		eventSer = eventSerialize(event)
		inlineKeyboardRow.Add("Today", spendingAddPrefix+string(eventSer))
		event.Today = false
		event.D = -1
		event.M = -1
		event.Y = -1
		eventSer = eventSerialize(event)
		inlineKeyboardRow.Add("Choose date", spendingAddPrefix+string(eventSer))
		event.SelectedToday = false
		event.CategoryId = -1
		eventSer = eventSerialize(event)
		inlineKeyboardRow2 := client.NewKeyboardRow()
		inlineKeyboardRow2.Add("<< Back", spendingAddPrefix+string(eventSer))
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow, inlineKeyboardRow2)
		err = s.client.SendCallbackQuery(inlineKeyboardRows, msg,
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
	} else if event.Price > 0 {
		// show categories
		msg := fmt.Sprintf("Choose category (*%.2f %s*):", event.Price, userCurrAbbr)
		categories, err := s.reposCat.Categories(ctx)
		if err != nil {
			return errors.Wrap(err, "event add categories")
		}
		if len(categories) == 0 {
			_ = s.client.SendMessage(
				"Categories list is empty, please add /categories", update.CallbackQuery.Message.Chat.ID)
			return errors.New("Categories list is empty")
		}
		for _, c := range categories {
			event.CategoryId = c.Id
			eventSer = eventSerialize(event)
			inlineKeyboardRow.Add(c.Title, spendingAddPrefix+string(eventSer))
		}
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
		err = s.client.SendCallbackQuery(inlineKeyboardRows, msg,
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
		if err != nil {
			return errors.Wrap(err, "send callback query")
		}
	}

	return
}

func (s *Service) GetRateUserFloat(ctx context.Context) (r decimal.Decimal, err error) {
	userCtx, err := user.FromContext(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "user not found")
	}
	uState, err := userCtx.GetState(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "state not found")
	}
	uCurrency, err := uState.GetCurrency(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "currency not found")
	}
	userRate, ok := s.rates.GetRate(ctx, uCurrency)
	if !ok {
		return 0, errors.Wrap(err, "rate not found")
	}

	return userRate.Rate, nil
}

func (s *Service) ConvertPrice(ctx context.Context, price decimal.Decimal) (f decimal.Decimal, err error) {
	userRateFloat64, err := s.GetRateUserFloat(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "convert price")
	}

	return userRateFloat64.Multiply(price), nil
}

func (s *Service) checkLimitPrice(ctx context.Context, category model.Category) (mess string, err error) {
	userCtx, err := user.FromContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "user not found")
	}
	uState, err := userCtx.GetState(ctx)
	if err != nil {
		return "", errors.Wrap(err, "state not found")
	}
	uCurrency, err := uState.GetCurrency(ctx)
	if err != nil {
		return "", errors.Wrap(err, "currency not found")
	}

	limits, err := uState.GetLimits(ctx)
	if err != nil {
		return "", errors.Wrap(err, "check limit price get limits")
	}
	categoryLimit := decimal.Decimal(0)
	for _, limit := range limits {
		if limit.Category.Id == category.Id {
			categoryLimit = limit.Limit
			break
		}
	}
	if categoryLimit > 0 {
		now := time.Now()
		f1 := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		f2 := time.Date(f1.Year(), f1.Month()+1, 0, 23, 59, 59, 0, f1.Location())

		m, err := s.reposSpend.Report(ctx, f1, f2, s.rates)
		if err != nil {
			return "", errors.Wrap(err, "check limit price report 31")
		}
		if sum, ok := m[category.Id]; ok {
			userRateFloat64, err := s.GetRateUserFloat(ctx)
			if err != nil {
				return "", errors.Wrap(err, "convert price")
			}
			sum = sum.Multiply(userRateFloat64)
			if sum > categoryLimit {
				mess = fmt.Sprintf("Sum *%.2f %s* by category *%s* over than *%.2f %s*",
					sum.Divide(userRateFloat64), uCurrency.Abbr,
					category.Title, categoryLimit.Divide(userRateFloat64), uCurrency.Abbr)
			}
		}
	}

	return
}

func eventSerialize(event *Event) string {
	return strings.Join([]string{
		strconv.FormatFloat(event.Price, 'f', 2, 64),
		strconv.Itoa(event.CategoryId),
		strconv.FormatBool(event.SelectedToday),
		strconv.FormatBool(event.Today),
		strconv.Itoa(event.D),
		strconv.Itoa(event.M),
		strconv.Itoa(event.Y),
	}, "_")
}

func eventUnserialize(s string) (*Event, error) {
	event := NewEvent(0)
	args := strings.Split(s, "_")

	price, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return nil, err
	}
	event.Price = price

	categoryId, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, err
	}
	event.CategoryId = categoryId

	selectedToday, err := strconv.ParseBool(args[2])
	if err != nil {
		return nil, err
	}
	event.SelectedToday = selectedToday

	today, err := strconv.ParseBool(args[3])
	if err != nil {
		return nil, err
	}
	event.Today = today

	d, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, err
	}
	event.D = d

	m, err := strconv.Atoi(args[5])
	if err != nil {
		return nil, err
	}
	event.M = m

	y, err := strconv.Atoi(args[6])
	if err != nil {
		return nil, err
	}
	event.Y = y

	return event, nil
}
