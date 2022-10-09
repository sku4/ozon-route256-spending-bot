package state

import (
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/currency"
	"sync"
)

type State struct {
	currency *currency.Currency
	mutex    sync.RWMutex
}

func (s *State) SetCurrency(c *currency.Currency) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currency = c
}

func (s *State) GetCurrency() *currency.Currency {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.currency
}

func NewState() (s *State, err error) {
	c, err := currency.GetById(currency.DefaultCurrency)
	if err != nil {
		return nil, errors.Wrap(err, "new state")
	}

	return &State{
		currency: c,
	}, nil
}
