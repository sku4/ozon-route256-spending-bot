package rates

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/decimal"
)

type Client interface {
	IsLoaded(context.Context) bool
	GetRate(context.Context, model.Currency) (*Rate, bool)
	UpdateRates(context.Context) error
	UpdateRatesSync(context.Context) bool
	SyncChan(context.Context) <-chan error
}

type Rate struct {
	model.Currency
	Rate decimal.Decimal
}
