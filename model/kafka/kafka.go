package kafka

import (
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
	BrokersList = append(BrokersList, os.Getenv("KAFKA_URL"))
}
