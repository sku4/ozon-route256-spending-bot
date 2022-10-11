package spending

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/category"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
	"sync"
	"time"
)

const (
	eventTable = "event"
)

type Spending struct {
	mutex          *sync.RWMutex
	db             *sqlx.DB
	categorySearch category.Search
}

type Event struct {
	model.Event
	Price decimal.Decimal
}

func NewSpending(db *sqlx.DB, categorySearch category.Search) *Spending {
	return &Spending{
		mutex:          &sync.RWMutex{},
		db:             db,
		categorySearch: categorySearch,
	}
}

func (s *Spending) AddEvent(ctx context.Context, categoryId int, date time.Time, price decimal.Decimal) (eventId int, err error) {
	s.mutex.Lock()

	cat, err := s.categorySearch.CategoryGetById(ctx, categoryId)
	if errors.Is(err, category.NotFoundError) {
		s.mutex.Unlock()
		return 0, errors.Wrap(err, "category not found")
	}

	createCategoryQuery := fmt.Sprintf("INSERT INTO %s (category_id, event_at, price) values ($1, $2, $3) RETURNING id",
		eventTable)
	row := s.db.QueryRowContext(ctx, createCategoryQuery, cat.Id, date.Format("2006-01-02"), price.Original())
	err = row.Scan(&eventId)
	if err != nil {
		s.mutex.Unlock()
		return 0, errors.Wrap(err, "insert event")
	}

	s.mutex.Unlock()

	return
}

func (s *Spending) DeleteEvent(ctx context.Context, id int) (err error) {
	s.mutex.Lock()

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, eventTable)
	_, err = s.db.ExecContext(ctx, query, id)
	if err != nil {
		s.mutex.Unlock()
		return errors.Wrap(err, "delete event")
	}

	s.mutex.Unlock()

	return
}

func (s Spending) Report(ctx context.Context, f1, f2 time.Time, rates rates.Client) (m map[int]decimal.Decimal, err error) {
	stat := make(map[int]decimal.Decimal)

	var events []model.EventDB
	query := fmt.Sprintf(`SELECT id, category_id, event_at, price FROM %s WHERE event_at BETWEEN '%s' AND '%s'`,
		eventTable, f1.Format("2006-01-02"), f2.Format("2006-01-02"))
	if err = s.db.SelectContext(ctx, &events, query); err != nil {
		return nil, errors.Wrap(err, "select report")
	}

	for _, event := range events {
		stat[event.CategoryId] += decimal.ToDecimal(event.Price)
	}

	userCtx, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "user not found")
	}
	userState, err := userCtx.GetState(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "user state")
	}
	userCurr, err := userState.GetCurrency(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "user currency")
	}
	rateUserCurr, ok := rates.GetRate(ctx, userCurr)
	if !ok {
		return nil, errors.Wrap(err, "user currency not found")
	}
	rateUserDecimal := rateUserCurr.Rate

	m = make(map[int]decimal.Decimal)
	for catId, sum := range stat {
		m[catId] = sum.Divide(rateUserDecimal)
	}

	return m, nil
}
