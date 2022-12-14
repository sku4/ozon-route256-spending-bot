package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/category"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/category_limit"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/currency"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/rates"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/spending"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/state"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/user"
	"github.com/sku4/ozon-route256-spending-bot/model"
	"github.com/sku4/ozon-route256-spending-bot/pkg/decimal"
	"time"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Spending interface {
	AddEvent(context.Context, int, time.Time, decimal.Decimal) (int, error)
	DeleteEvent(context.Context, int) error
	Report(context.Context, time.Time, time.Time, rates.Client, model.Currency) (map[int]decimal.Decimal, error)
}

type Categories interface {
	Categories(context.Context) ([]model.Category, error)
	AddCategory(context.Context, string) (int, error)
	DeleteCategory(context.Context, int) error
	category.Search
}

type Users interface {
	AddUser(context.Context, int) (*user.User, error)
	GetByTgId(context.Context, int) (*user.User, error)
}

type Repository struct {
	Spending
	Categories
	Users
	CurrencyClient   currency.Client
	RatesClient      rates.Client
	StateClient      state.Client
	CategorySearch   category.Search
	CategoryLimitSet category_limit.CategoryLimitSet
}

func NewRepository(db *sqlx.DB) (*Repository, error) {
	currencyClient, err := currency.NewCurrencies(db)
	if err != nil {
		return nil, errors.Wrap(err, "create repository currencies")
	}
	categoryClient := category.NewCategory(db)
	spendingClient := spending.NewSpending(db, categoryClient)
	categoryLimitSet := category_limit.NewCategoryLimit(db, categoryClient)
	stateClient := state.NewStates(db, currencyClient, categoryLimitSet)
	usersClient := user.NewUsers(db, currencyClient, stateClient)

	return &Repository{
		Spending:         spendingClient,
		Categories:       categoryClient,
		Users:            usersClient,
		CurrencyClient:   currencyClient,
		StateClient:      stateClient,
		CategoryLimitSet: categoryLimitSet,
	}, nil
}
