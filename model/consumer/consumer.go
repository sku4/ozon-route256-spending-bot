package consumer

import (
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler/consumer"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
)

type Consumer struct {
	handler consumer.IHandler
}

func NewConsumer(services *service.ReportService) *Consumer {
	h := consumer.NewHandler(services)
	h = MetricsMiddleware(h)
	return &Consumer{
		handler: h,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	logger.Info("consumer - setup")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Info("consumer - cleanup")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		_ = c.handler.ReportMessage(session.Context(), message)
		session.MarkMessage(message, "")
	}

	return nil
}
