package spending

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/category"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
	"sync"
	"time"
)

const (
	decimalFactor float64 = 100
	eventTable            = "event"
)

type Spending struct {
	mutex          *sync.RWMutex
	db             *sqlx.DB
	categorySearch category.Search
}

type Event struct {
	model.Event
	Price PriceFloat64
}

type PriceFloat64 int64

func (p PriceFloat64) Original() int64 {
	return int64(p)
}

func (p PriceFloat64) Float() float64 {
	return float64(p) / decimalFactor
}

func (p PriceFloat64) String() string {
	return fmt.Sprintf("%.2f", float64(p)/decimalFactor)
}

func NewSpending(db *sqlx.DB, categorySearch category.Search) *Spending {
	return &Spending{
		mutex:          &sync.RWMutex{},
		db:             db,
		categorySearch: categorySearch,
	}
}

func (s *Spending) AddEvent(ctx context.Context, categoryId int, date time.Time, price float64) (eventId int, err error) {
	_ = ctx

	s.mutex.Lock()

	cat, err := s.categorySearch.CategoryGetById(categoryId)
	if errors.Is(err, category.NotFoundError) {
		s.mutex.Unlock()
		return 0, errors.Wrap(err, "category not found")
	}

	createCategoryQuery := fmt.Sprintf("INSERT INTO %s (category_id, event_at, price) values ($1, $2, $3) RETURNING id",
		eventTable)
	row := s.db.QueryRow(createCategoryQuery, cat.Id, date.Format("2006-01-02"), Float64ToPrice(price))
	err = row.Scan(&eventId)
	if err != nil {
		s.mutex.Unlock()
		return 0, errors.Wrap(err, "insert event")
	}

	s.mutex.Unlock()

	return
}

func (s *Spending) DeleteEvent(ctx context.Context, id int) (err error) {
	_ = ctx

	s.mutex.Lock()

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, eventTable)
	_, err = s.db.Exec(query, id)
	if err != nil {
		s.mutex.Unlock()
		return errors.Wrap(err, "delete event")
	}

	s.mutex.Unlock()

	return
}

func (s Spending) Report(ctx context.Context, f1, f2 time.Time, rates rates.Client) (m map[int]float64, err error) {
	_ = ctx

	s.mutex.RLock()
	stat := make(map[int]PriceFloat64)

	var events []Event
	query := fmt.Sprintf(`SELECT id, category_id, event_at, price FROM %s WHERE event_at BETWEEN '%s' AND '%s'`,
		eventTable, f1.Format("2006-01-02"), f2.Format("2006-01-02"))
	if err = s.db.Select(&events, query); err != nil {
		return nil, errors.Wrap(err, "select report")
	}

	for _, event := range events {
		stat[event.Category.Id] += event.Price
	}
	s.mutex.RUnlock()

	userCtx, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "user not found")
	}
	userState, err := userCtx.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "user state")
	}
	userCurr, err := userState.GetCurrency()
	if err != nil {
		return nil, errors.Wrap(err, "user currency")
	}
	rateUserCurr, ok := rates.GetRate(ctx, userCurr)
	if !ok {
		return nil, errors.Wrap(err, "user currency not found")
	}
	rateUserFloat := rateUserCurr.Rate.Float()

	m = make(map[int]float64)
	for catId, sum := range stat {
		m[catId] = sum.Float() / rateUserFloat
	}

	return m, nil
}

func Float64ToPrice(f float64) PriceFloat64 {
	return PriceFloat64(f * decimalFactor)
}
