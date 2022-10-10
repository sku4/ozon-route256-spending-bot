package model

type User struct {
	Id    int
	TgId  int
	State *State
}

type UserDB struct {
	Id      int `db:"id"`
	TgId    int `db:"telegram_id"`
	StateId int `db:"state_id"`
}
