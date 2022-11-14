package main

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/skubach/workshop-1-bot/configs"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/consumer"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/kafka"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/cache"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg, err := configs.Init()
	if err != nil {
		logger.Fatalf("error init config: %s", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		stop()
		logger.Info("Context is stopped")
	}()

	if err := godotenv.Load(); err != nil {
		logger.Fatalf("error loading env variables: %s", err.Error())
	}

	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB_NAME"),
		SslMode:  os.Getenv("POSTGRES_SSL"),
	})
	if err != nil {
		logger.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos, err := repository.NewRepository(db)
	if err != nil {
		logger.Fatalf("failed init repository: %s", err.Error())
	}
	ratesClient := initRates(ctx, db, repos)

	grpcConn, err := initGrpcConn(os.Getenv("GRPC_URL"))
	if err != nil {
		logger.Fatalf("failed init grpc client: %s", err.Error())
	}
	defer func() {
		err = grpcConn.Close()
		if err != nil {
			logger.Info(fmt.Sprintf("Unrecognized consumer group partition assignor: %s", kafka.Assignor))
		}
		logger.Info("grpc client connection closed")
	}()
	grpcClient := api.NewSpendingClient(grpcConn)

	services := service.NewReportService(repos, ratesClient, grpcClient)
	consumerGroupHandler := consumer.NewConsumer(services)

	http.Handle("/metrics", promhttp.Handler())

	cache.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err = startConsumerGroup(ctx, consumerGroupHandler); err != nil {
			logger.Fatalf("error start consumer group: %s", err.Error())
		}
	}()

	go func() {
		logger.Info(fmt.Sprintf("Rest server is running on: %d", cfg.ReportRestPort))
		err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.ReportRestPort), nil)
		if err != nil {
			logger.Fatalf("error starting http server", err.Error())
			quit <- nil
		}
	}()

	logger.Info("App Started")

	// graceful shutdown
	logger.Info(fmt.Sprintf("Got signal %v, attempting graceful shutdown", <-quit))

	logger.Info("App Shutting Down")
}

func startConsumerGroup(ctx context.Context, consumerGroupHandler *consumer.Consumer) error {
	config := sarama.NewConfig()
	config.Version = sarama.V3_2_3_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 3 * time.Second

	switch kafka.Assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	case "round-robin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	default:
		logger.Fatalf("Unrecognized consumer group partition assignor: %s", kafka.Assignor)
	}

	logger.Info(fmt.Sprintf("Kafka brokers: %s", strings.Join(kafka.BrokersList, ", ")))

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(kafka.BrokersList, kafka.ConsumerGroup, config)
	if err != nil {
		return fmt.Errorf("starting consumer group: %w", err)
	}

	err = consumerGroup.Consume(ctx, []string{kafka.TopicReport}, consumerGroupHandler)
	if err != nil {
		return fmt.Errorf("consuming via handler: %w", err)
	}

	return nil
}

func initRates(ctx context.Context, db *sqlx.DB, repos *repository.Repository) rates.Client {
	ratesClient := nbrb.NewRates(db, repos.CurrencyClient)
	run := ratesClient.UpdateRatesSync(ctx)

	if run {
		go func() {
			// read channel error
			<-ratesClient.SyncChan(ctx)
		}()
	}

	return ratesClient
}

func initGrpcConn(grpcUrl string) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "did not connect")
	}

	return
}
