package model

type State struct {
	Id       int
	Currency *Currency
}

type StateDB struct {
	Id         int `db:"id"`
	CurrencyId int `db:"currency_id"`
}
