package client

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

type Client struct {
	client *tgbotapi.BotAPI
	ctx    context.Context
}

const KeyboardButtonTypeSwitch = "switch"

type KeyboardRow struct {
	buttons []KeyboardButton
}

func NewKeyboardRow() *KeyboardRow {
	buttons := make([]KeyboardButton, 0)
	return &KeyboardRow{
		buttons: buttons,
	}
}

func (i *KeyboardRow) Add(k, v string) {
	i.buttons = append(i.buttons, KeyboardButton{
		k, v, "",
	})
}

func (i *KeyboardRow) AddSwitch(k, v string) {
	i.buttons = append(i.buttons, KeyboardButton{
		k, v, KeyboardButtonTypeSwitch,
	})
}

type KeyboardButton struct {
	k, v, t string
}

func NewClient(ctx context.Context, client *tgbotapi.BotAPI) (*Client, error) {
	return &Client{
		ctx:    ctx,
		client: client,
	}, nil
}

func (c *Client) SendMessage(message string, chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true

	_, err := c.client.Send(msg)
	if err != nil {
		return errors.Wrap(err, "send message")
	}

	return nil
}

func (c *Client) SendInlineKeyboard(keyboardRows []*KeyboardRow, message string, chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, message)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = getInlineKeyboard(keyboardRows)

	_, err := c.client.Send(msg)
	if err != nil {
		return errors.Wrap(err, "send inline keyboard")
	}

	return nil
}

func (c *Client) SendCallbackQuery(keyboardRows []*KeyboardRow,
	message string, messageId int, chatId int64) error {
	_ = message

	msgEdit := tgbotapi.NewEditMessageText(chatId, messageId, message)
	msgEdit.ParseMode = tgbotapi.ModeMarkdown
	msgEdit.DisableWebPagePreview = true
	msgEdit.ReplyMarkup = getInlineKeyboard(keyboardRows)

	if _, err := c.client.Send(msgEdit); err != nil {
		return errors.Wrap(err, "send markup callback query")
	}

	return nil
}

func getInlineKeyboard(inlineKeyboardRows []*KeyboardRow) *tgbotapi.InlineKeyboardMarkup {
	var keyboardButtons [][]tgbotapi.InlineKeyboardButton
	for _, row := range inlineKeyboardRows {
		maxButtons := 8
		inlineRow := make([]tgbotapi.InlineKeyboardButton, 0, maxButtons)
		for _, b := range row.buttons {
			switch b.t {
			case KeyboardButtonTypeSwitch:
				inlineRow = append(inlineRow, tgbotapi.NewInlineKeyboardButtonSwitch(b.k, b.v))
			default:
				inlineRow = append(inlineRow, tgbotapi.NewInlineKeyboardButtonData(b.k, b.v))
			}
			if len(inlineRow) == maxButtons {
				keyboardButtons = append(keyboardButtons, inlineRow)
				inlineRow = make([]tgbotapi.InlineKeyboardButton, 0, maxButtons)
			}
		}
		if len(inlineRow) > 0 {
			keyboardButtons = append(keyboardButtons, inlineRow)
		}
	}

	if len(keyboardButtons) == 0 {
		inlineRow := make([]tgbotapi.InlineKeyboardButton, 0)
		keyboardButtons = append(keyboardButtons, inlineRow)
	}

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboardButtons,
	}
}
