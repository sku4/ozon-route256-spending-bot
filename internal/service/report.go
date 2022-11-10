package service

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/repository/postgres/rates"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service/spending"
	"gitlab.ozon.dev/skubach/workshop-1-bot/model"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
	"time"
)

//go:generate mockgen -source=service.go -destination=mocks/report.go

type BuildReport interface {
	Build(context.Context, time.Time, time.Time, model.Currency, int64) error
}

type ReportService struct {
	BuildReport
}

func NewReportService(repos *repository.Repository, rates rates.Client, grpcClient api.SpendingClient) *ReportService {
	return &ReportService{
		BuildReport: spending.NewReport(repos.Spending, repos.Categories, rates, grpcClient),
	}
}
