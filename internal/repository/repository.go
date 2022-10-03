package repository

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/spending"
	"time"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Spending interface {
	Events(context.Context) []spending.Event
	AddEvent(context.Context, int, time.Time, float64) ([]spending.Event, error)
	DeleteEvent(context.Context, int) ([]spending.Event, error)
	Report(context.Context, time.Time, time.Time) map[int]float64
	Categories
}

type Categories interface {
	Categories(context.Context) []spending.Category
	AddCategory(context.Context, string) ([]spending.Category, error)
	DeleteCategory(context.Context, int) ([]spending.Category, error)
}

type Repository struct {
	Spending
}

func NewRepository() *Repository {
	return &Repository{
		Spending: spending.NewSpending(),
	}
}
