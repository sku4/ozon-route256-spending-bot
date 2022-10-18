package state

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/category_limit"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
	"sync"
)

const (
	stateTable         = "state"
	currencyTable      = "currency"
	categoryLimitTable = "category_limit"
	categoryTable      = "category"
)

var (
	queryUpdate      = fmt.Sprintf(`UPDATE %s SET currency_id=$1 WHERE id=$2`, stateTable)
	querySelect      = fmt.Sprintf(`SELECT id, currency_id FROM %s WHERE id=$1`, stateTable)
	queryInsert      = fmt.Sprintf("INSERT INTO %s (currency_id) values ($1) RETURNING id", stateTable)
	queryGetWithCurr = fmt.Sprintf(`
				SELECT st.id, c.id as currency_id, c.abbreviation as currency_abbr 
						FROM %s as st
						LEFT JOIN %s as c on c.id = st.currency_id
						WHERE st.id=$1`, stateTable, currencyTable)
	queryGetWithLimits = fmt.Sprintf(`
				SELECT cl.id as category_limit_id, cl.category_id, cat.title as category_title, cl.category_limit 
						FROM %s as cl
						LEFT JOIN %s as cat on cat.id = cl.category_id
						WHERE cl.state_id=$1`,
		categoryLimitTable, categoryTable)
)

type Client interface {
	GetById(context.Context, int) (*State, error)
	GetByIdTx(context.Context, *sql.Tx, int) (*State, error)
	AddState(context.Context) (*State, error)
	AddStateTx(context.Context, *sql.Tx) (*State, error)
}

type State struct {
	model.State
	limits           []*category_limit.CategoryLimit
	mutex            *sync.RWMutex
	db               *sqlx.DB
	reposCurr        currency.Client
	reposCatLimitSet category_limit.CategoryLimitSet
}

func (s *State) SetCurrency(ctx context.Context, c model.Currency) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err = s.db.ExecContext(ctx, queryUpdate, c.Id, s.Id)
	if err != nil {
		return errors.Wrap(err, "update currency")
	}
	s.Currency = c

	return
}

func (s *State) GetCurrency(ctx context.Context) (c model.Currency, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.Currency == c {
		var state model.StateDB
		if err = s.db.GetContext(ctx, &state, querySelect, s.Id); err != nil {
			return model.Currency{}, errors.Wrap(err, "state get currency")
		}
		s.Currency, err = s.reposCurr.GetById(ctx, state.CurrencyId)
		if err != nil {
			return model.Currency{}, errors.Wrap(err, "state currency not found")
		}
	}

	return s.Currency, nil
}

func (s *State) AddLimit(ctx context.Context, categoryId int, limit decimal.Decimal) (err error) {
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
	var state model.StateWithLimits
	if err = s.db.GetContext(ctx, &state, queryGetWithCurr, id); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("state '%d' not found", id))
	}

	if state.CurrencyId == 0 {
		return nil, errors.Wrap(err, fmt.Sprintf("currency '%d' not found", id))
	}

	var curr = model.Currency{
		Id:   state.CurrencyId,
		Abbr: state.CurrencyAbbr,
	}

	var stateLimits []model.StateWithLimits
	if err = s.db.SelectContext(ctx, &stateLimits, queryGetWithLimits, id); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("state limits '%d' not found", id))
	}

	var limits []*category_limit.CategoryLimit
	for _, limit := range stateLimits {
		c := &model.Category{
			Id:    limit.CategoryId,
			Title: limit.CategoryTitle,
		}
		limitDB := &model.CategoryLimitDB{
			Id:         limit.CategoryLimitId,
			Limit:      limit.CategoryLimit,
			CategoryId: limit.CategoryId,
			StateId:    id,
		}
		limits = append(limits, s.reposCatLimitSet.Create(ctx, c, limitDB))
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

func (s *States) GetByIdTx(ctx context.Context, tx *sql.Tx, id int) (st *State, err error) {
	var state model.StateWithLimits
	row := tx.QueryRowContext(ctx, queryGetWithCurr, id)
	err = row.Scan(&state.Id, &state.CurrencyId, &state.CurrencyAbbr)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("state '%d' not found", id))
	}

	var curr = model.Currency{
		Id:   state.CurrencyId,
		Abbr: state.CurrencyAbbr,
	}

	var stateLimits []model.StateWithLimits
	rows, err := tx.QueryContext(ctx, queryGetWithLimits, id)
	if err != nil {
		return nil, errors.Wrap(err, "with limits select")
	}

	for rows.Next() {
		var stateLimit model.StateWithLimits
		err = rows.Scan(&stateLimit.CategoryLimitId, &stateLimit.CategoryId,
			&stateLimit.CategoryTitle, &stateLimit.CategoryLimit)
		if err != nil {
			break
		}
		stateLimits = append(stateLimits, stateLimit)
	}

	var limits []*category_limit.CategoryLimit
	for _, limit := range stateLimits {
		c := &model.Category{
			Id:    limit.CategoryId,
			Title: limit.CategoryTitle,
		}
		limitDB := &model.CategoryLimitDB{
			Id:         limit.CategoryLimitId,
			Limit:      limit.CategoryLimit,
			CategoryId: limit.CategoryId,
			StateId:    id,
		}
		limits = append(limits, s.reposCatLimitSet.Create(ctx, c, limitDB))
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
	row := s.db.QueryRowContext(ctx, queryInsert, c.Id)
	err = row.Scan(&stateId)
	if err != nil {
		return nil, errors.Wrap(err, "insert state")
	}

	return &State{
		State: model.State{
			Id:       stateId,
			Currency: c,
		},
		limits:           []*category_limit.CategoryLimit{},
		mutex:            &sync.RWMutex{},
		db:               s.db,
		reposCurr:        s.reposCurr,
		reposCatLimitSet: s.reposCatLimitSet,
	}, nil
}

func (s *States) AddStateTx(ctx context.Context, tx *sql.Tx) (st *State, err error) {
	c := s.reposCurr.GetDefault(ctx)

	var stateId int
	row := tx.QueryRowContext(ctx, queryInsert, c.Id)
	err = row.Scan(&stateId)
	if err != nil {
		return nil, errors.Wrap(err, "insert state")
	}

	return &State{
		State: model.State{
			Id:       stateId,
			Currency: c,
		},
		limits:           []*category_limit.CategoryLimit{},
		mutex:            &sync.RWMutex{},
		db:               s.db,
		reposCurr:        s.reposCurr,
		reposCatLimitSet: s.reposCatLimitSet,
	}, nil
}
