package currency

const (
	USD = iota
	CNY
	EUR
	RUB
	DefaultCurrency = RUB
)

var (
	AbbrCurr = map[string]Currency{
		"USD": USD,
		"CNY": CNY,
		"EUR": EUR,
		"RUB": RUB,
	}
	CurrAbbr = map[Currency]string{
		USD: "USD",
		CNY: "CNY",
		EUR: "EUR",
		RUB: "RUB",
	}
)

type Currency int
