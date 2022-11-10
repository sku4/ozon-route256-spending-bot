package grpc

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api/report"
)

func (h *Handler) SendReport(ctx context.Context, in *report.Report) (*api.Empty, error) {
	var empty = &api.Empty{}
	err := h.services.SendReport(ctx, in.Message, in.F1.AsTime(), in.F2.AsTime(), in.ChatId)
	if err != nil {
		return empty, err
	}

	return empty, nil
}
