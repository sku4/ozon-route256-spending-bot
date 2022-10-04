package currency

const (
	USD = iota
	CNY
	EUR
	RUB
	DefaultCurrency = RUB
)

var (
	Abbreviation = map[string]Currency{
		"USD": USD,
		"CNY": CNY,
		"EUR": EUR,
		"RUB": RUB,
	}
)

type Currency int
