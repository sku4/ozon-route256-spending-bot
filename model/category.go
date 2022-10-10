package model

type Category struct {
	Id    int    `db:"id"`
	Title string `db:"title"`
}
