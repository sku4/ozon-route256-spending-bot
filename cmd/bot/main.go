package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/configs"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates/nbrb"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/client"
	tg "gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot/server"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/log"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)
	sugar := logger.Sugar()

	cfg, err := configs.Init()
	if err != nil {
		sugar.Fatalf("error init config: %s", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	ctx = log.ContextWithLogger(ctx, logger)
	defer func() {
		stop()
		sugar.Info("Context is stopped")
	}()

	tgClient, tgServer, err := initTelegramBot(ctx, cfg)
	if err != nil {
		sugar.Fatalf("error init telegram bot: %s", err.Error())
	}

	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		DBName:   cfg.DBName,
		SslMode:  cfg.SslMode,
		Password: cfg.Password,
	})
	if err != nil {
		sugar.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos, err := repository.NewRepository(db)
	if err != nil {
		sugar.Fatalf("failed init repository: %s", err.Error())
	}
	ratesClient := nbrb.NewRates(db, repos.CurrencyClient)
	ratesClient.UpdateRatesSync(ctx)

	services := service.NewService(repos, tgClient, ratesClient)
	handlers := handler.NewHandler(ctx, services)

	quit := make(chan os.Signal, 1)
	go func() {
		if err := tgServer.Run(handlers); err != nil {
			sugar.Fatalf("error occured while running: %s", err.Error())
			quit <- nil
		}
	}()

	sugar.Info("App Started")
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	sugar.Info("App Shutting Down")
}

func initTelegramBot(ctx context.Context, cfg *configs.Config) (tgClient *client.Client, tgServer *tg.Server, err error) {
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
