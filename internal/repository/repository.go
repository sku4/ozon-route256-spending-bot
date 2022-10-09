package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/spending"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/user"
	"time"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Spending interface {
	Events(context.Context) []spending.Event
	AddEvent(context.Context, int, time.Time, float64) ([]spending.Event, error)
	DeleteEvent(context.Context, int) ([]spending.Event, error)
	Report(context.Context, time.Time, time.Time, rates.Client) (map[int]float64, error)
	Categories
}

type Categories interface {
	Categories(context.Context) []spending.Category
	AddCategory(context.Context, string) ([]spending.Category, error)
	DeleteCategory(context.Context, int) ([]spending.Category, error)
}

type Users interface {
	AddUser(int) (*user.User, error)
	GetUserById(int) (*user.User, error)
}

type Repository struct {
	Spending
	Users
	CurrencyClient currency.Client
	RatesClient    rates.Client
}

func NewRepository(db *sqlx.DB) (*Repository, error) {
	currencies, err := currency.NewCurrencies(db)
	if err != nil {
		return nil, errors.Wrap(err, "create repository currencies")
	}

	return &Repository{
		Spending:       spending.NewSpending(db),
		Users:          user.NewUsers(db, currencies),
		CurrencyClient: currencies,
	}, nil
}
