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
)

type Rates struct {
	m          map[*model.Currency]*rates.Rate
	lastUpdate time.Time
	loaded     bool
	mutex      sync.RWMutex
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
	}
}

func (rs *Rates) IsLoaded(ctx context.Context) bool {
	_ = ctx

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	return rs.loaded
}

func (rs *Rates) GetRate(ctx context.Context, curr *model.Currency) (*rates.Rate, bool) {
	_ = ctx

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	r, ok := rs.m[curr]
	if !ok && curr.Id == rs.reposCurr.GetDefault().Id {
		defCur := rs.reposCurr.GetDefault()

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
	currencyRub := rs.reposCurr.GetDefault()
	for _, nbrbRate := range nbrbRates {
		if nbrbRate.CurAbbreviation == currencyRub.Abbr {
			rateByn = nbrbRate.CurOfficialRate / float64(nbrbRate.CurScale)
			break
		}
	}
	for _, nbrbRate := range nbrbRates {
		curr, err := rs.reposCurr.GetByAbbr(nbrbRate.CurAbbreviation)
		if err != nil {
			continue
		}
		rate := nbrbRate.CurOfficialRate / float64(nbrbRate.CurScale)
		r := rate / rateByn
		rs.m[curr] = &rates.Rate{
			Currency: curr,
			Rate:     rates.Float64ToRate(r),
		}
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
