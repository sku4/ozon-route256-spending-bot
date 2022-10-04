package middleware

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/user"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
)

type Middleware struct {
	users  repository.Users
	rates  repository.Rates
	client client.BotClient
}

func NewMiddleware(users repository.Users, rates repository.Rates, client client.BotClient) *Middleware {
	return &Middleware{
		users:  users,
		rates:  rates,
		client: client,
	}
}

func (m Middleware) DefineUser(ctx context.Context, update tgbotapi.Update) (context.Context, error) {
	userId := 0
	if update.Message != nil {
		userId = update.Message.From.ID
	} else if update.CallbackQuery != nil {
		userId = update.CallbackQuery.From.ID
	}

	u := m.users.AddUser(userId)
	ctx = UserToContext(ctx, u)

	return ctx, nil
}

func (m Middleware) UpdateRates(ctx context.Context) {
	m.rates.UpdateRatesSync(ctx)

	return
}

type ctxUser struct{}

func UserToContext(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, ctxUser{}, u)
}

func UserFromContext(ctx context.Context) (u *user.User, err error) {
	if u, ok := ctx.Value(ctxUser{}).(*user.User); ok {
		return u, nil
	}
	return nil, errors.New("user not found in context")
}
