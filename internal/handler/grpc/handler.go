package grpc

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
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
