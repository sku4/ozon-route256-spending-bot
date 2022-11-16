package user

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/user"
)

type ctxUser struct{}

func ToContext(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, ctxUser{}, u)
}

func FromContext(ctx context.Context) (u *user.User, err error) {
	if u, ok := ctx.Value(ctxUser{}).(*user.User); ok {
		return u, nil
	}
	return nil, errors.New("user not found in context")
}
