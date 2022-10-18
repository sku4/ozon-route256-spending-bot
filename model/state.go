package model

type State struct {
	Id       int
	Currency Currency
}

type StateDB struct {
	Id         int `db:"id"`
	CurrencyId int `db:"currency_id"`
}

type StateWithLimits struct {
	Id              int    `db:"id"`
	CurrencyId      int    `db:"currency_id"`
	CurrencyAbbr    string `db:"currency_abbr"`
	CategoryId      int    `db:"category_id"`
	CategoryTitle   string `db:"category_title"`
	CategoryLimit   int64  `db:"category_limit"`
	CategoryLimitId int    `db:"category_limit_id"`
}
