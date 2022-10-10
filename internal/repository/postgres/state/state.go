package state

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/category_limit"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"sync"
)

const (
	stateTable = "state"
)

type Client interface {
	GetById(int) (*State, error)
}

type State struct {
	model.State
	limits           []*category_limit.CategoryLimit
	mutex            *sync.RWMutex
	db               *sqlx.DB
	reposCurr        currency.Client
	reposCatLimitSet category_limit.CategoryLimitSet
}

func NewState(db *sqlx.DB, reposCurr currency.Client, reposCatLimitSet category_limit.CategoryLimitSet) (s *State, err error) {
	c := reposCurr.GetDefault()

	var stateId int
	createStateQuery := fmt.Sprintf("INSERT INTO %s (currency_id) values ($1) RETURNING id", stateTable)
	row := s.db.QueryRow(createStateQuery, c.Id)
	err = row.Scan(&stateId)
	if err != nil {
		return nil, errors.Wrap(err, "insert state")
	}

	limits, err := reposCatLimitSet.GetByState(stateId)
	if err != nil {
		return nil, errors.Wrap(err, "new state")
	}

	return &State{
		State: model.State{
			Id:       stateId,
			Currency: c,
		},
		limits:           limits,
		mutex:            &sync.RWMutex{},
		db:               db,
		reposCurr:        reposCurr,
		reposCatLimitSet: reposCatLimitSet,
	}, nil
}

func (s *State) SetCurrency(c *model.Currency) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	query := fmt.Sprintf(`UPDATE %s SET currency_id = %d WHERE id = %d`, stateTable, c.Id, s.Id)
	_, err = s.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "update currency")
	}
	s.Currency = c

	return
}

func (s *State) GetCurrency() (c *model.Currency, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.Currency == nil {
		var state *model.StateDB
		query := fmt.Sprintf(`SELECT id, currency_id FROM %s WHERE id = %d`, stateTable, s.Id)
		if err = s.db.Select(&state, query); err != nil {
			return nil, errors.Wrap(err, "state get currency")
		}
		s.Currency, err = s.reposCurr.GetById(state.CurrencyId)
		if err != nil {
			return nil, errors.Wrap(err, "state currency not found")
		}
	}

	return s.Currency, nil
}

func (s *State) AddLimit(categoryId int, limit float64) (err error) {
	_, err = s.reposCatLimitSet.Set(s.Id, categoryId, limit)

	return
}

func (s *State) GetLimits() (cls []*category_limit.CategoryLimit, err error) {
	if s.limits == nil {
		s.limits, err = s.reposCatLimitSet.GetByState(s.Id)
		if err != nil {
			return nil, errors.Wrap(err, "get limits")
		}
	}

	return s.limits, nil
}

type States struct {
	states           []*State
	mutex            *sync.RWMutex
	db               *sqlx.DB
	reposCurr        currency.Client
	reposCatLimitSet category_limit.CategoryLimitSet
}

func NewStates(db *sqlx.DB, reposCurr currency.Client, reposCatLimitSet category_limit.CategoryLimitSet) *States {
	cs := &States{
		states:           make([]*State, 0),
		db:               db,
		mutex:            &sync.RWMutex{},
		reposCurr:        reposCurr,
		reposCatLimitSet: reposCatLimitSet,
	}

	return cs
}

func (s *States) GetById(id int) (st *State, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var state *model.StateDB
	query := fmt.Sprintf(`SELECT id, currency_id FROM %s WHERE id = %d`, stateTable, id)
	if err = s.db.Select(&state, query); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("state '%d' not found", id))
	}

	curr, err := s.reposCurr.GetById(state.CurrencyId)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("currency '%d' not found", id))
	}

	limits, err := s.reposCatLimitSet.GetByState(state.Id)
	if err != nil {
		return nil, errors.Wrap(err, "new state")
	}

	st = &State{
		State: model.State{
			Id:       state.Id,
			Currency: curr,
		},
		limits:    limits,
		mutex:     s.mutex,
		db:        s.db,
		reposCurr: s.reposCurr,
	}

	return
}
