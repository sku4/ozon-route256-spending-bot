package repository

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/spending"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/user"
	"time"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Spending interface {
	Events(context.Context) []spending.Event
	AddEvent(context.Context, int, time.Time, float64) ([]spending.Event, error)
	DeleteEvent(context.Context, int) ([]spending.Event, error)
	Report(context.Context, time.Time, time.Time, *currency.Rates) (map[int]float64, error)
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

type Rates interface {
	UpdateRates(context.Context) error
	UpdateRatesSync(context.Context)
}

type Repository struct {
	Spending
	Users
	Rates
}

func NewRepository() *Repository {
	return &Repository{
		Spending: spending.NewSpending(),
		Users:    user.NewUsers(),
	}
}
