package repository

import (
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/spending"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

type Spending interface {
	//Start(context.Context) error
}

type Repository struct {
	Spending
}

func NewRepository() *Repository {
	return &Repository{
		Spending: spending.NewSpending(),
	}
}
