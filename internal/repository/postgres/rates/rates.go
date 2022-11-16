package rates

import (
	"context"
	"github.com/sku4/ozon-route256-spending-bot/model"
	"github.com/sku4/ozon-route256-spending-bot/pkg/decimal"
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
