package nbrb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	nbrbRatesUrl = "https://www.nbrb.by/api/exrates/rates?periodicity=0"
	updateTime   = time.Hour
	rateTable    = "rate"
)

var (
	queryTruncate         = fmt.Sprintf("TRUNCATE TABLE %s", rateTable)
	queryInsert           = fmt.Sprintf("INSERT INTO %s (currency_id, rate) values ($1, $2)", rateTable)
	HistogramCurrencyRate = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "bot",
			Subsystem: "rate",
			Name:      "histogram_currency_rate_value",
			Buckets:   []float64{1, 5, 10, 25, 50, 60, 75, 100, 120},
		},
		[]string{"abbr"},
	)
)

type Rates struct {
	m          map[model.Currency]*rates.Rate
	lastUpdate time.Time
	loaded     bool
	mutex      *sync.RWMutex
	syncChan   chan error
	db         *sqlx.DB
	reposCurr  currency.Client
}

func NewRates(db *sqlx.DB, reposCurrencies currency.Client) *Rates {
	return &Rates{
		m:         make(map[model.Currency]*rates.Rate),
		loaded:    false,
		db:        db,
		reposCurr: reposCurrencies,
		mutex:     &sync.RWMutex{},
		syncChan:  make(chan error),
	}
}

func (rs *Rates) IsLoaded(ctx context.Context) bool {
	_ = ctx

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	return rs.loaded
}

func (rs *Rates) GetRate(ctx context.Context, curr model.Currency) (*rates.Rate, bool) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	r, ok := rs.m[curr]
	if !ok && curr.Id == rs.reposCurr.GetDefault(ctx).Id {
		defCur := rs.reposCurr.GetDefault(ctx)

		return &rates.Rate{
			Currency: defCur,
			Rate:     1,
		}, true
	}

	return r, ok
}

func (rs *Rates) UpdateRates(ctx context.Context) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UpdateRates")
	defer span.Finish()

	resp, err := http.Get(nbrbRatesUrl)
	if err != nil {
		return errors.Wrap(err, "nbrb get rates")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "nbrb rates read all")
	}

	err = resp.Body.Close()
	if err != nil {
		return errors.Wrap(err, "nbrb rates body close")
	}

	var nbrbRates []nbrb.Rate
	if err = json.Unmarshal(body, &nbrbRates); err != nil {
		return errors.Wrap(err, "nbrb rates unmarshal")
	}

	if resp.StatusCode != http.StatusOK {
		logger.Info(string(body))
		return errors.New(fmt.Sprintf("Response status: %s", resp.Status))
	}

	rs.mutex.Lock()
	rateByn := decimal.Decimal(0)
	currencyDef := rs.reposCurr.GetDefault(ctx)
	for _, nbrbRate := range nbrbRates {
		if nbrbRate.CurAbbreviation == currencyDef.Abbr {
			rateByn = decimal.ToDecimal(nbrbRate.CurOfficialRate).Divide(
				decimal.ToDecimal(nbrbRate.CurScale))
			break
		}
	}

	tx, err := rs.db.Begin()
	if err != nil {
		rs.mutex.Unlock()
		return errors.Wrap(err, "nbrb tx begin")
	}

	_, err = tx.Exec(queryTruncate)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			rs.mutex.Unlock()
			return errors.Wrap(errRoll, "rollback")
		}
		rs.mutex.Unlock()
		return errors.Wrap(err, "truncate rates")
	}
	rs.m = make(map[model.Currency]*rates.Rate)

	for _, nbrbRate := range nbrbRates {
		curr, err := rs.reposCurr.GetByAbbr(ctx, nbrbRate.CurAbbreviation)
		if err != nil {
			continue
		}
		rate := decimal.ToDecimal(nbrbRate.CurOfficialRate).Divide(
			decimal.ToDecimal(nbrbRate.CurScale))
		r := rate.Divide(rateByn)

		_, err = tx.Exec(queryInsert, curr.Id, r.Original())
		if err != nil {
			errRoll := tx.Rollback()
			if errRoll != nil {
				rs.mutex.Unlock()
				return errors.Wrap(errRoll, "rollback")
			}
			rs.mutex.Unlock()
			return errors.Wrap(err, "insert rate")
		}

		span.SetTag(curr.Abbr, r)
		HistogramCurrencyRate.
			WithLabelValues(curr.Abbr).
			Observe(r.Float64())

		rs.m[curr] = &rates.Rate{
			Currency: curr,
			Rate:     r,
		}
	}

	err = tx.Commit()
	if err != nil {
		rs.mutex.Unlock()
		return errors.Wrap(err, "rates commit")
	}
	rs.lastUpdate = time.Now()
	rs.loaded = true
	logger.Info("Rates success updates")
	rs.mutex.Unlock()

	return
}

func (rs *Rates) UpdateRatesSync(ctx context.Context) (run bool) {
	rs.mutex.RLock()
	if rs.lastUpdate.Add(updateTime).After(time.Now()) {
		rs.mutex.RUnlock()
		return
	}
	rs.mutex.RUnlock()

	go func(ctx context.Context, rates *Rates) {
		err := rates.UpdateRates(ctx)
		rs.syncChan <- err
	}(ctx, rs)

	return true
}

func (rs *Rates) SyncChan(ctx context.Context) <-chan error {
	_ = ctx

	return rs.syncChan
}
