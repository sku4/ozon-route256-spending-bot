package currency

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"sync"
)

type Client interface {
	All() []*model.Currency
	GetDefault() *model.Currency
	GetByAbbr(string) (*model.Currency, error)
	GetById(int) (*model.Currency, error)
}

const (
	currencyTable = "currency"
)

var (
	defaultCurrencyAbbr = "RUB"
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

	query := fmt.Sprintf(`SELECT id, abbreviation FROM %s`, currencyTable)
	if err := cs.db.Select(&cs.currencies, query); err != nil {
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

func (cs *Currencies) All() []*model.Currency {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return cs.currencies
}

func (cs *Currencies) GetDefault() *model.Currency {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return cs.defaultCurrency
}

func (cs *Currencies) GetByAbbr(abbr string) (curr *model.Currency, err error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	for _, currency := range cs.currencies {
		if currency.Abbr == abbr {
			return currency, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("currency '%s' not found", abbr))
}

func (cs *Currencies) GetById(id int) (curr *model.Currency, err error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	for _, currency := range cs.currencies {
		if currency.Id == id {
			return currency, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("currency '%d' not found", id))
}
