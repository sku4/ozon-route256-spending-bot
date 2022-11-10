package kafka

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

var (
	TopicReport   = "report"
	BrokersList   []string
	ConsumerGroup = "report-consumer-group"
	Assignor      = "range"
)

func init() {
	_ = godotenv.Load()
	kafkaPort := os.Getenv("KAFKA_ADVERTISED_PORT")
	BrokersList = append(BrokersList, fmt.Sprintf("localhost:%s", kafkaPort))
}
