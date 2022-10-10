package state

import (
	"context"
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
	GetById(context.Context, int) (*State, error)
	AddState(context.Context) (*State, error)
}

type State struct {
	model.State
	limits           []*category_limit.CategoryLimit
	mutex            *sync.RWMutex
	db               *sqlx.DB
	reposCurr        currency.Client
	reposCatLimitSet category_limit.CategoryLimitSet
}

func (s *State) SetCurrency(ctx context.Context, c *model.Currency) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	query := fmt.Sprintf(`UPDATE %s SET currency_id = %d WHERE id = %d`, stateTable, c.Id, s.Id)
	_, err = s.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "update currency")
	}
	s.Currency = c

	return
}

func (s *State) GetCurrency(ctx context.Context) (c *model.Currency, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.Currency == nil {
		var state model.StateDB
		query := fmt.Sprintf(`SELECT id, currency_id FROM %s WHERE id = %d`, stateTable, s.Id)
		if err = s.db.GetContext(ctx, &state, query); err != nil {
			return nil, errors.Wrap(err, "state get currency")
		}
		s.Currency, err = s.reposCurr.GetById(ctx, state.CurrencyId)
		if err != nil {
			return nil, errors.Wrap(err, "state currency not found")
		}
	}

	return s.Currency, nil
}

func (s *State) AddLimit(ctx context.Context, categoryId int, limit float64) (err error) {
	_, err = s.reposCatLimitSet.Set(ctx, s.Id, categoryId, limit)

	return
}

func (s *State) GetLimits(ctx context.Context) (cls []*category_limit.CategoryLimit, err error) {
	if s.limits == nil {
		s.limits, err = s.reposCatLimitSet.GetByState(ctx, s.Id)
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

func (s *States) GetById(ctx context.Context, id int) (st *State, err error) {
	var state model.StateDB
	query := fmt.Sprintf(`SELECT id, currency_id FROM %s WHERE id = %d`, stateTable, id)
	if err = s.db.GetContext(ctx, &state, query); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("state '%d' not found", id))
	}

	curr, err := s.reposCurr.GetById(ctx, state.CurrencyId)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("currency '%d' not found", id))
	}

	limits, err := s.reposCatLimitSet.GetByState(ctx, state.Id)
	if err != nil {
		return nil, errors.Wrap(err, "new state")
	}

	st = &State{
		State: model.State{
			Id:       state.Id,
			Currency: curr,
		},
		limits:           limits,
		mutex:            s.mutex,
		db:               s.db,
		reposCurr:        s.reposCurr,
		reposCatLimitSet: s.reposCatLimitSet,
	}

	return
}

func (s *States) AddState(ctx context.Context) (st *State, err error) {
	c := s.reposCurr.GetDefault(ctx)

	var stateId int
	createStateQuery := fmt.Sprintf("INSERT INTO %s (currency_id) values ($1) RETURNING id", stateTable)
	row := s.db.QueryRowContext(ctx, createStateQuery, c.Id)
	err = row.Scan(&stateId)
	if err != nil {
		return nil, errors.Wrap(err, "insert state")
	}

	limits, err := s.reposCatLimitSet.GetByState(ctx, stateId)
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
		db:               s.db,
		reposCurr:        s.reposCurr,
		reposCatLimitSet: s.reposCatLimitSet,
	}, nil
}
