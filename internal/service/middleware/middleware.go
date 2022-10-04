package middleware

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/user"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
)

type iUpdate interface {
	tgbotapi.Update | tgbotapi.CallbackQuery
}

type Update[T iUpdate] struct {
	elem T
}

func NewUpdate[T iUpdate](elem T) *Update[T] {
	return &Update[T]{
		elem: elem,
	}
}

func (u Update[T]) GetUserId() (id int, err error) {
	if upd, ok := any(u.elem).(tgbotapi.Update); ok {
		return upd.Message.From.ID, nil
	} else if upd, ok := any(u.elem).(tgbotapi.CallbackQuery); ok {
		return upd.Message.From.ID, nil
	}

	return 0, errors.New("user id not found")
}

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
	upd := NewUpdate(update)
	userId, err := upd.GetUserId()
	if err != nil {
		return ctx, err
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
