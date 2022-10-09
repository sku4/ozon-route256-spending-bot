package rates

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
)

const (
	decimalFactor float64 = 10000
)

type Client interface {
	IsLoaded(context.Context) bool
	GetRate(context.Context, *model.Currency) (*Rate, bool)
	UpdateRates(context.Context) error
	UpdateRatesSync(context.Context)
	SyncChan() chan struct{}
}

type Rate struct {
	*model.Currency
	Rate RateFloat64
}

type RateFloat64 int64

func (p RateFloat64) Original() int64 {
	return int64(p)
}

func (p RateFloat64) Float() float64 {
	return float64(p) / decimalFactor
}

func Float64ToRate(f float64) RateFloat64 {
	return RateFloat64(f * decimalFactor)
}
