package currency

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	USD = iota
	CNY
	EUR
	RUB
	DefaultCurrency = RUB
)

var (
	currencies = []*Currency{
		{
			Id:   USD,
			Abbr: "USD",
		},
		{
			Id:   CNY,
			Abbr: "CNY",
		},
		{
			Id:   EUR,
			Abbr: "EUR",
		},
		{
			Id:   RUB,
			Abbr: "RUB",
		},
	}
)

type Currency struct {
	Id   int
	Abbr string
}

func All() []*Currency {
	return currencies
}

func GetByAbbr(abbr string) (curr *Currency, err error) {
	for _, currency := range currencies {
		if currency.Abbr == abbr {
			return currency, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("currency '%s' not found", abbr))
}

func GetById(id int) (curr *Currency, err error) {
	for _, currency := range currencies {
		if currency.Id == id {
			return currency, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("currency '%d' not found", id))
}
