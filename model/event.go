package model

import "time"

type Event struct {
	Id       int
	Category Category
	Date     time.Time
	Price    int64
}

type EventDB struct {
	Id         int       `db:"id"`
	CategoryId int       `db:"category_id"`
	Date       time.Time `db:"event_at"`
	Price      int64     `db:"price"`
	CreatedAt  time.Time `db:"created_at"`
}
