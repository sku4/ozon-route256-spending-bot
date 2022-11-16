package grpc

import (
	"context"
	"github.com/sku4/ozon-route256-spending-bot/internal/service"
	"github.com/sku4/ozon-route256-spending-bot/pkg/api"
)

// Handler struct with grpc api server
type Handler struct {
	ctx      context.Context
	services service.Service
	api.SpendingServer
}

// NewHandler creates a new handler
func NewHandler(ctx context.Context, services *service.Service) *Handler {
	return &Handler{
		ctx:      ctx,
		services: *services,
	}
}
