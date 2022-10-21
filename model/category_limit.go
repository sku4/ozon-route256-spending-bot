package model

type CategoryLimit struct {
	Id       int
	Category *Category
	Limit    int64
}

type CategoryLimitDB struct {
	Id            int    `db:"id"`
	StateId       int    `db:"state_id"`
	CategoryId    int    `db:"category_id"`
	CategoryTitle string `db:"category_title"`
	Limit         int64  `db:"category_limit"`
}
