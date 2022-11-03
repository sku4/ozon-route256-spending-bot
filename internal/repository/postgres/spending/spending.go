package spending

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/category"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/cache"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/user"
	"time"
)

const (
	eventTable = "event"
)

var (
	queryInsert = fmt.Sprintf("INSERT INTO %s (category_id, event_at, price) values ($1, $2, $3) RETURNING id",
		eventTable)
	queryDelete = fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, eventTable)
	queryReport = fmt.Sprintf(`SELECT category_id, sum(price) as price FROM `+
		`%s WHERE event_at BETWEEN $1 AND $2 GROUP BY category_id`,
		eventTable)
	histogramEventPrice = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "bot",
			Subsystem: "event",
			Name:      "histogram_summary_event_category_price",
			Buckets:   []float64{1, 10, 100, 1000, 10000, 100000, 1000000},
		},
		[]string{"category"},
	)
)

type Spending struct {
	db             *sqlx.DB
	categorySearch category.Search
}

type Event struct {
	model.Event
	Price decimal.Decimal
}

func NewSpending(db *sqlx.DB, categorySearch category.Search) *Spending {
	return &Spending{
		db:             db,
		categorySearch: categorySearch,
	}
}

func (s *Spending) AddEvent(ctx context.Context, categoryId int, date time.Time, price decimal.Decimal) (eventId int, err error) {
	cat, err := s.categorySearch.CategoryGetById(ctx, categoryId)
	if errors.Is(err, category.NotFoundError) {
		return 0, errors.Wrap(err, "category not found")
	}

	row := s.db.QueryRowContext(ctx, queryInsert, cat.Id, date.Format("2006-01-02"), price.Original())
	err = row.Scan(&eventId)
	if err != nil {
		return 0, errors.Wrap(err, "insert event")
	}

	histogramEventPrice.
		WithLabelValues(cat.Title).
		Observe(price.Float64())

	return
}

func (s *Spending) DeleteEvent(ctx context.Context, id int) (err error) {
	_, err = s.db.ExecContext(ctx, queryDelete, id)
	if err != nil {
		return errors.Wrap(err, "delete event")
	}

	return
}

func (s Spending) Report(ctx context.Context, f1, f2 time.Time, rates rates.Client) (m map[int]decimal.Decimal, err error) {
	var events []model.EventDB
	keyCacheReport := fmt.Sprintf("events_report_%s_%s", f1.Format("2006_01_02"), f2.Format("2006_01_02"))
	err = cache.Once(&cache.Item{
		Key:   keyCacheReport,
		Value: &events,
		TTL:   time.Minute * 10,
		Ctx:   ctx,
	}, func(ci *cache.Item) (interface{}, error) {
		if err = s.db.SelectContext(ci.Ctx, &events, queryReport,
			f1.Format("2006-01-02"), f2.Format("2006-01-02")); err != nil {
			return nil, errors.Wrap(err, "select report")
		}
		return &events, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "report cache")
	}

	stat := make(map[int]decimal.Decimal)

	for _, event := range events {
		stat[event.CategoryId] = decimal.Decimal(event.Price)
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
