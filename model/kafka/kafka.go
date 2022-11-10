package kafka

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

var (
	TopicReport = "report"
	BrokersList []string
)

func init() {
	_ = godotenv.Load()
	kafkaPort := os.Getenv("KAFKA_ADVERTISED_PORT")

	BrokersList = append(BrokersList, fmt.Sprintf("localhost:%s", kafkaPort))
}
