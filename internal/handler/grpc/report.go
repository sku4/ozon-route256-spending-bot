package grpc

import (
	"context"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api/report"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) SendReport(ctx context.Context, in *report.Report) (*api.Empty, error) {
	var empty = &api.Empty{}
	if !in.F1.IsValid() || !in.F2.IsValid() {
		return empty, status.Error(codes.InvalidArgument, "invalid date period")
	}

	err := h.services.SendReport(ctx, in.Message, in.F1.AsTime(), in.F2.AsTime(), in.ChatId)
	if err != nil {
		return empty, err
	}

	return empty, nil
}
