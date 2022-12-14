package main

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sku4/ozon-route256-spending-bot/configs"
	"github.com/sku4/ozon-route256-spending-bot/internal/handler/grpc"
	"github.com/sku4/ozon-route256-spending-bot/internal/handler/telegram"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/rates"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/rates/nbrb"
	"github.com/sku4/ozon-route256-spending-bot/internal/service"
	"github.com/sku4/ozon-route256-spending-bot/model/kafka"
	"github.com/sku4/ozon-route256-spending-bot/model/server"
	"github.com/sku4/ozon-route256-spending-bot/model/telegram/bot/client"
	tg "github.com/sku4/ozon-route256-spending-bot/model/telegram/bot/server"
	"github.com/sku4/ozon-route256-spending-bot/pkg/cache"
	"github.com/sku4/ozon-route256-spending-bot/pkg/logger"
	jaeger "github.com/uber/jaeger-client-go/config"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	cfg, err := configs.Init()
	if err != nil {
		logger.Fatalf("error init config: %s", err.Error())
	}

	err = initTracing(cfg)
	if err != nil {
		logger.Fatalf("error init tracing: %s", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		stop()
		logger.Info("Context is stopped")
	}()

	tgClient, tgServer, err := InitTelegramBot(ctx, cfg)
	if err != nil {
		logger.Fatalf("error init telegram bot: %s", err.Error())
	}

	if err = godotenv.Load(); err != nil {
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

	logger.Info(fmt.Sprintf("Kafka brokers: %s", strings.Join(kafka.BrokersList, ", ")))
	kafkaProducer, err := initKafkaProducer(ctx, kafka.BrokersList)
	if err != nil {
		logger.Fatalf("failed init kafka: %s", err.Error())
	}
	defer func() {
		err = kafkaProducer.Close()
		if err != nil {
			logger.Info(fmt.Sprintf("failed to close producer: %s", err.Error()))
		}
	}()

	cache.Run(ctx)

	repos, err := repository.NewRepository(db)
	if err != nil {
		logger.Fatalf("failed init repository: %s", err.Error())
	}
	ratesClient := InitRates(ctx, db, repos)
	services := service.NewService(repos, tgClient, ratesClient, kafkaProducer)
	handlers := telegram.NewHandler(services)
	grpcHandlers := grpc.NewHandler(ctx, services)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// run telegram server
	go func() {
		if err = tgServer.Run(ctx, handlers); err != nil {
			logger.Fatalf("error occured while running: %s", err.Error())
			quit <- nil
		}
	}()

	// run grpc server
	grpcServer := server.NewGrpc(ctx, grpcHandlers)
	go func() {
		if err = grpcServer.Run(os.Getenv("GRPC_URL")); err != nil {
			logger.Info(err.Error())
			quit <- nil
		}
	}()

	// run rest server
	restServer := server.NewRest(ctx)
	go func() {
		if err = restServer.Run(os.Getenv("GRPC_URL"), cfg.BotRestPort); err != nil {
			logger.Info(err.Error())
			quit <- nil
		}
	}()

	logger.Info("App Started")

	// graceful shutdown
	logger.Info(fmt.Sprintf("Got signal %v, attempting graceful shutdown", <-quit))
	stop()
	logger.Info("Context is stopped")
	grpcServer.GracefulStop()
	logger.Info("gRPC graceful stopped")
	err = restServer.Shutdown()
	if err != nil {
		logger.Info(fmt.Sprintf("error rest server shutdown: %s", err.Error()))
	} else {
		logger.Info("Rest server stopped")
	}

	logger.Info("App Shutting Down")
}

func InitTelegramBot(ctx context.Context, cfg *configs.Config) (tgClient *client.Client, tgServer *tg.Server, err error) {
	tgBot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error init telegram bot")
	}

	tgClient, err = client.NewClient(ctx, tgBot)
	if err != nil {
		return nil, nil, errors.Wrap(err, "telegram client init failed")
	}

	tgServer, err = tg.NewServer(ctx, tgBot)
	if err != nil {
		return nil, nil, errors.Wrap(err, "telegram server init failed")
	}

	return
}

func InitRates(ctx context.Context, db *sqlx.DB, repos *repository.Repository) rates.Client {
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

func initTracing(cfg *configs.Config) (err error) {
	c := jaeger.Configuration{
		Sampler: &jaeger.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}
	_, err = c.InitGlobalTracer(cfg.ServiceName)

	return
}

func initKafkaProducer(ctx context.Context, brokerList []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_2_3_0
	// So we can know the partition and offset of messages.
	config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		return nil, fmt.Errorf("starting Sarama producer: %w", err)
	}

	// We will log to STDOUT if we're not able to produce messages.
	go func() {
		for {
			select {
			case err := <-producer.Errors():
				logger.Infos("Failed to write message:", err)
			case successMsg := <-producer.Successes():
				logger.Infos("Successful to write message, offset:", successMsg.Offset)
			case <-ctx.Done():
				return
			}
		}
	}()

	return producer, nil
}
