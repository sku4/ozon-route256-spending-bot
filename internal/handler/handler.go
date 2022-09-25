package handler

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/service"
)

type Handler struct {
	ctx      context.Context
	services service.Service
}

func NewHandler(ctx context.Context, services *service.Service) *Handler {
	return &Handler{
		ctx:      ctx,
		services: *services,
	}
}
