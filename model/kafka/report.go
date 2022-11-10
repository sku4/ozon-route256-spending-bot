package kafka

import (
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"time"
)

type Report struct {
	F1, F2   time.Time
	ChatId   int64
	UserCurr model.Currency
}
