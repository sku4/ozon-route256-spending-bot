package currency

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"sync"
)

type Client interface {
	All(context.Context) []*model.Currency
	GetDefault(context.Context) *model.Currency
	GetByAbbr(context.Context, string) (*model.Currency, error)
	GetById(context.Context, int) (*model.Currency, error)
}

const (
	currencyTable = "currency"
)

var (
	defaultCurrencyAbbr = "RUB"
	querySelectAll      = fmt.Sprintf(`SELECT id, abbreviation FROM %s`, currencyTable)
)

type Currencies struct {
	currencies      []*model.Currency
	db              *sqlx.DB
	mutex           *sync.RWMutex
	defaultCurrency *model.Currency
}

func NewCurrencies(db *sqlx.DB) (*Currencies, error) {
	cs := &Currencies{
		currencies: make([]*model.Currency, 0),
		db:         db,
		mutex:      &sync.RWMutex{},
	}

	if err := cs.db.Select(&cs.currencies, querySelectAll); err != nil {
		return nil, errors.Wrap(err, "select all currencies")
	}

	for _, currency := range cs.currencies {
		if currency.Abbr == defaultCurrencyAbbr {
			cs.defaultCurrency = currency
			break
		}
	}

	return cs, nil
}

func (cs *Currencies) All(ctx context.Context) []*model.Currency {
	_ = ctx

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return cs.currencies
}

func (cs *Currencies) GetDefault(ctx context.Context) *model.Currency {
	_ = ctx

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return cs.defaultCurrency
}

func (cs *Currencies) GetByAbbr(ctx context.Context, abbr string) (curr *model.Currency, err error) {
	_ = ctx

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	for _, currency := range cs.currencies {
		if currency.Abbr == abbr {
			return currency, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("currency '%s' not found", abbr))
}

func (cs *Currencies) GetById(ctx context.Context, id int) (curr *model.Currency, err error) {
	_ = ctx

	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	for _, currency := range cs.currencies {
		if currency.Id == id {
			return currency, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("currency '%d' not found", id))
}
