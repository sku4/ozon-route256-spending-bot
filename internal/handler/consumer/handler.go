package consumer

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/sku4/ozon-route256-spending-bot/internal/service"
)

type IHandler interface {
	ReportMessage(context.Context, *sarama.ConsumerMessage) error
}

type Func func(context.Context, *sarama.ConsumerMessage) error

func (f Func) ReportMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	return f(ctx, msg)
}

type Handler struct {
	services service.ReportService
}

func NewHandler(services *service.ReportService) IHandler {
	return &Handler{
		services: *services,
	}
}
