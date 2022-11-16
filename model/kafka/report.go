package kafka

import (
	"github.com/sku4/ozon-route256-spending-bot/model"
	"time"
)

type Report struct {
	F1, F2   time.Time
	ChatId   int64
	UserCurr model.Currency
}
