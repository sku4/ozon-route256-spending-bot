package state

import (
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"sync"
)

type State struct {
	model.State
	mutex *sync.RWMutex
}

func (s *State) SetCurrency(c *model.Currency) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Currency = c
}

func (s *State) GetCurrency() *model.Currency {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Currency
}

func NewState(reposCurr currency.Client) (s *State, err error) {
	c := reposCurr.GetDefault()

	return &State{
		State: model.State{
			Currency: c,
		},
		mutex: &sync.RWMutex{},
	}, nil
}
