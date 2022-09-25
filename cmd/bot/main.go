package main

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/configs"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	tg "gitlab.ozon.dev/skubach/workshop-1-bot/model/telegram/bot"
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

	tgClient, err := tg.New(ctx, cfg.TelegramBotToken)
	if err != nil {
		sugar.Fatalf("spending bot init failed: %s", err.Error())
	}

	repos := repository.NewRepository()
	services := service.NewService(repos)
	handlers := handler.NewHandler(ctx, services)

	quit := make(chan os.Signal, 1)
	go func() {
		if err := tgClient.Run(handlers); err != nil {
			sugar.Fatalf("error occured while running: %s", err.Error())
			quit <- nil
		}
	}()

	sugar.Info("App Started")
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	sugar.Info("App Shutting Down")
}
