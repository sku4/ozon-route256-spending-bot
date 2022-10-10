package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/category_limit"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/spending"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/state"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/user"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"time"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Spending interface {
	AddEvent(context.Context, int, time.Time, float64) (int, error)
	DeleteEvent(context.Context, int) error
	Report(context.Context, time.Time, time.Time, rates.Client) (map[int]float64, error)
	Categories
}

type Categories interface {
	Categories(context.Context) ([]model.Category, error)
	AddCategory(context.Context, string) ([]model.Category, error)
	DeleteCategory(context.Context, int) ([]model.Category, error)
	spending.CategorySearch
}

type Users interface {
	AddUser(int) (*user.User, error)
	GetById(int) (*user.User, error)
}

type Repository struct {
	Spending
	Users
	CurrencyClient   currency.Client
	RatesClient      rates.Client
	StateClient      state.Client
	CategorySearch   spending.CategorySearch
	CategoryLimitSet category_limit.CategoryLimitSet
}

func NewRepository(db *sqlx.DB) (*Repository, error) {
	currencyClient, err := currency.NewCurrencies(db)
	if err != nil {
		return nil, errors.Wrap(err, "create repository currencies")
	}
	spendingClient := spending.NewSpending(db)
	categoryLimitSet := category_limit.NewCategoryLimit(db, spendingClient)
	stateClient := state.NewStates(db, currencyClient, categoryLimitSet)
	usersClient := user.NewUsers(db, currencyClient, stateClient, categoryLimitSet)

	return &Repository{
		Spending:         spendingClient,
		Users:            usersClient,
		CurrencyClient:   currencyClient,
		StateClient:      stateClient,
		CategoryLimitSet: categoryLimitSet,
	}, nil
}
