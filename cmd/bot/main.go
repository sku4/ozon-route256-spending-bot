package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/configs"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	tg "gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/server"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"os"
	"os/signal"
	"syscall"
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

	tgClient, tgServer, err := InitTelegramBot(ctx, cfg)
	if err != nil {
		logger.Fatalf("error init telegram bot: %s", err.Error())
	}

	if err = godotenv.Load(); err != nil {
		logger.Fatalf("error loading env variables: %s", err.Error())
	}

	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     cfg.Port,
		Username: cfg.Username,
		DBName:   cfg.DBName,
		SslMode:  cfg.SslMode,
		Password: os.Getenv("POSTGRES_PASSWORD"),
	})
	if err != nil {
		logger.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos, err := repository.NewRepository(db)
	if err != nil {
		logger.Fatalf("failed init repository: %s", err.Error())
	}
	ratesClient := InitRates(ctx, db, repos)
	services := service.NewService(repos, tgClient, ratesClient)
	handlers := handler.NewHandler(ctx, services)

	quit := make(chan os.Signal, 1)
	go func() {
		if err := tgServer.Run(handlers); err != nil {
			logger.Fatalf("error occured while running: %s", err.Error())
			quit <- nil
		}
	}()

	logger.Info("App Started")
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

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
