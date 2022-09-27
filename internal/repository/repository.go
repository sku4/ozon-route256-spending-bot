package repository

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/spending"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Spending interface {
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
