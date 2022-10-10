package nbrb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/currency"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/log"
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

type Rates struct {
	m          map[*model.Currency]*rates.Rate
	lastUpdate time.Time
	loaded     bool
	mutex      *sync.RWMutex
	syncChan   chan struct{}
	db         *sqlx.DB
	reposCurr  currency.Client
}

func NewRates(db *sqlx.DB, reposCurrencies currency.Client) *Rates {
	return &Rates{
		m:         make(map[*model.Currency]*rates.Rate),
		loaded:    false,
		db:        db,
		reposCurr: reposCurrencies,
		mutex:     &sync.RWMutex{},
	}
}

func (rs *Rates) IsLoaded(ctx context.Context) bool {
	_ = ctx

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	return rs.loaded
}

func (rs *Rates) GetRate(ctx context.Context, curr *model.Currency) (*rates.Rate, bool) {
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
	if rs.lastUpdate.Add(updateTime).After(time.Now()) {
		return nil
	}

	logger := log.LoggerFromContext(ctx)

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
	rateByn := float64(0)
	currencyDef := rs.reposCurr.GetDefault(ctx)
	for _, nbrbRate := range nbrbRates {
		if nbrbRate.CurAbbreviation == currencyDef.Abbr {
			rateByn = nbrbRate.CurOfficialRate / float64(nbrbRate.CurScale)
			break
		}
	}

	tx, err := rs.db.Begin()
	if err != nil {
		rs.mutex.Unlock()
		return errors.Wrap(err, "nbrb tx begin")
	}

	truncateRates := fmt.Sprintf("TRUNCATE TABLE %s", rateTable)
	_, err = tx.Exec(truncateRates)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			rs.mutex.Unlock()
			return errors.Wrap(errRoll, "rollback")
		}
		rs.mutex.Unlock()
		return errors.Wrap(err, "truncate rates")
	}
	rs.m = make(map[*model.Currency]*rates.Rate)

	for _, nbrbRate := range nbrbRates {
		curr, err := rs.reposCurr.GetByAbbr(ctx, nbrbRate.CurAbbreviation)
		if err != nil {
			continue
		}
		rate := nbrbRate.CurOfficialRate / float64(nbrbRate.CurScale)
		r := rate / rateByn

		insertRateQuery := fmt.Sprintf("INSERT INTO %s (currency_id, rate) values ($1, $2)", rateTable)
		_, err = tx.Exec(insertRateQuery, curr.Id, rates.Float64ToRate(r))
		if err != nil {
			errRoll := tx.Rollback()
			if errRoll != nil {
				rs.mutex.Unlock()
				return errors.Wrap(errRoll, "rollback")
			}
			rs.mutex.Unlock()
			return errors.Wrap(err, "insert rate")
		}

		rs.m[curr] = &rates.Rate{
			Currency: curr,
			Rate:     rates.Float64ToRate(r),
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

func (rs *Rates) UpdateRatesSync(ctx context.Context) {
	rs.syncChan = make(chan struct{})
	go func(ctx context.Context, rates *Rates) {
		_ = rates.UpdateRates(ctx)
		rs.syncChan <- struct{}{}
	}(ctx, rs)
}

func (rs *Rates) SyncChan() chan struct{} {
	return rs.syncChan
}
