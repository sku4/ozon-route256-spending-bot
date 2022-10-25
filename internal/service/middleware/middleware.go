package middleware

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
)

type Middleware struct {
	users  repository.Users
	rates  rates.Client
	client client.BotClient
}

func NewMiddleware(users repository.Users, client client.BotClient, rates rates.Client) *Middleware {
	return &Middleware{
		users:  users,
		rates:  rates,
		client: client,
	}
}

func (m Middleware) DefineUser(ctx context.Context, update tgbotapi.Update) (context.Context, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "DefineUser")
	defer span.Finish()

	userId := 0
	if update.Message != nil {
		userId = update.Message.From.ID
	} else if update.CallbackQuery != nil {
		userId = update.CallbackQuery.From.ID
	}

	span.SetTag("tgUserId", userId)

	u, err := m.users.AddUser(ctx, userId)
	if err != nil {
		return nil, errors.Wrap(err, "define user")
	}
	ctx = user.ToContext(ctx, u)

	return ctx, nil
}

func (m Middleware) UpdateRatesSync(ctx context.Context) (run bool) {
	return m.rates.UpdateRatesSync(ctx)
}

func (m Middleware) RatesSyncChan(ctx context.Context) <-chan error {
	return m.rates.SyncChan(ctx)
}
