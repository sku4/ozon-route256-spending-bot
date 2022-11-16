package service

import (
	"context"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository"
	"github.com/sku4/ozon-route256-spending-bot/internal/repository/postgres/rates"
	"github.com/sku4/ozon-route256-spending-bot/internal/service/spending"
	"github.com/sku4/ozon-route256-spending-bot/model"
	"github.com/sku4/ozon-route256-spending-bot/pkg/api"
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
