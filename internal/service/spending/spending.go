package spending

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"strconv"
)

//go:generate mockgen -source=spending.go -destination=mocks/spending.go

type Service struct {
	repos  repository.Spending
	client *client.Client
}

func NewService(repos repository.Spending, client *client.Client) *Service {
	return &Service{
		repos:  repos,
		client: client,
	}
}

func (s *Service) Categories(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	var inlineKeyboardRows []*client.KeyboardRow
	inlineKeyboardRow := client.NewKeyboardRow()
	inlineKeyboardRow.Add("Add", "categories_add")
	inlineKeyboardRow.Add("List", "categories_list")
	inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)

	err = s.client.SendInlineKeyboard(inlineKeyboardRows,
		"Choose categories command:", update.Message.Chat.ID)
	if err != nil {
		return err
	}

	return
}

func (s *Service) CategoryAdd(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	title := update.Message.CommandArguments()
	_, err = s.repos.AddCategory(ctx, title)
	if err != nil {
		_ = s.client.SendMessage(fmt.Sprintf(
			"Error add category *%s*: %s", title, err.Error()), update.Message.Chat.ID)
		return errors.Wrap(err, "add category")
	}
	err = s.client.SendMessage(fmt.Sprintf(
		"Category *%s* success added\r\n"+
			"Show /categories", title), update.Message.Chat.ID)

	return
}

func (s *Service) CategoriesQuery(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	var inlineKeyboardRows []*client.KeyboardRow
	inlineKeyboardRow := client.NewKeyboardRow()
	msg := "Command not found"

	switch update.CallbackQuery.Data {
	case "categories_home":
		msg = "Choose categories command:"
		inlineKeyboardRow.Add("Add", "categories_add")
		inlineKeyboardRow.Add("List", "categories_list")
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow)
		err = s.client.SendCallbackQuery(inlineKeyboardRows, msg,
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
	case "categories_list":
		msg = "Categories list:"
		categories := s.repos.Categories(ctx)
		for _, category := range categories {
			inlineKeyboardRow.Add(category.Title, "categories_id_"+strconv.Itoa(category.Id))
		}
		inlineKeyboardRow2 := client.NewKeyboardRow()
		inlineKeyboardRow2.Add("<< Back", "categories_home")
		inlineKeyboardRows = append(inlineKeyboardRows, inlineKeyboardRow, inlineKeyboardRow2)
		err = s.client.SendCallbackQuery(inlineKeyboardRows, msg,
			update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID)
	case "categories_add":
		msg = "Write `/categoryadd Food` to added category"
		err = s.client.SendMessage(msg, update.CallbackQuery.Message.Chat.ID)
	}

	return
}

func (s *Service) Start(ctx context.Context, update tgbotapi.Update) (err error) {
	_ = ctx

	msg := "Available commands:\n" +
		"/categories\n" +
		"/categoryadd Food"
	err = s.client.SendMessage(msg, update.Message.Chat.ID)
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
