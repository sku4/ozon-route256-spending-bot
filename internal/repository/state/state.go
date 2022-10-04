package state

import (
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/currency"
)

type State struct {
	Currency *currency.Currency
}

func NewState() (s *State, err error) {
	c, err := currency.GetById(currency.DefaultCurrency)
	if err != nil {
		return nil, errors.Wrap(err, "new state")
	}

	return &State{
		Currency: c,
	}, nil
}
