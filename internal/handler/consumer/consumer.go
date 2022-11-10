package consumer

import (
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
)

type Consumer struct {
	services service.ReportService
}

func NewConsumer(services *service.ReportService) *Consumer {
	return &Consumer{
		services: *services,
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
		c.ReportMessage(session.Context(), message)
		session.MarkMessage(message, "")
	}

	return nil
}
