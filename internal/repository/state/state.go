package state

import "gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/currency"

type State struct {
	Currency currency.Currency
}

func NewState() *State {
	return &State{
		Currency: currency.DefaultCurrency,
	}
}
