package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/log"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	nbrbRatesUrl          = "https://www.nbrb.by/api/exrates/rates?periodicity=0"
	decimalFactor float64 = 10000
	updateTime            = time.Hour
)

type RatesClient interface {
	IsLoaded(context.Context) bool
	GetRate(context.Context, *Currency) (*Rate, bool)
	UpdateRates(context.Context) error
	UpdateRatesSync(context.Context)
	SyncChan() chan struct{}
}

type Rate struct {
	*Currency
	Rate RateFloat64
}

type RateFloat64 int64

func (p RateFloat64) Original() int64 {
	return int64(p)
}

func (p RateFloat64) Float() float64 {
	return float64(p) / decimalFactor
}

type Rates struct {
	m          map[*Currency]*Rate
	lastUpdate time.Time
	loaded     bool
	mutex      sync.RWMutex
	syncChan   chan struct{}
	db         *sqlx.DB
}

func NewRates(db *sqlx.DB) *Rates {
	return &Rates{
		m:      make(map[*Currency]*Rate),
		loaded: false,
		db:     db,
	}
}

func (rs *Rates) IsLoaded(ctx context.Context) bool {
	_ = ctx

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	return rs.loaded
}

func (rs *Rates) GetRate(ctx context.Context, currency *Currency) (*Rate, bool) {
	_ = ctx

	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	r, ok := rs.m[currency]
	if !ok && currency.Id == DefaultCurrency {
		defCur, err := GetById(DefaultCurrency)
		if err != nil {
			return nil, false
		}
		return &Rate{
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
	currencyRub, err := GetById(RUB)
	if err != nil {
		return errors.Wrap(err, "currency rub")
	}
	for _, nbrbRate := range nbrbRates {
		if nbrbRate.CurAbbreviation == currencyRub.Abbr {
			rateByn = nbrbRate.CurOfficialRate / float64(nbrbRate.CurScale)
			break
		}
	}
	for _, nbrbRate := range nbrbRates {
		currency, err := GetByAbbr(nbrbRate.CurAbbreviation)
		if err != nil {
			continue
		}
		rate := nbrbRate.CurOfficialRate / float64(nbrbRate.CurScale)
		r := rate / rateByn
		rs.m[currency] = &Rate{
			Currency: currency,
			Rate:     Float64ToRate(r),
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

func Float64ToRate(f float64) RateFloat64 {
	return RateFloat64(f * decimalFactor)
}
