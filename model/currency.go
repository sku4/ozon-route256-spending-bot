package model

type Currency struct {
	Id   int    `db:"id"`
	Abbr string `db:"abbreviation"`
}
