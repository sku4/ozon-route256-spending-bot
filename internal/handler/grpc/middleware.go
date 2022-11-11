package grpc

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	grpcTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bot",
			Name:      "grpc_total",
		},
		[]string{"method"},
	)
)

func (h *Handler) Metrics(ctx context.Context, fullMethodName string) (context.Context, error) {
	grpcTotal.WithLabelValues(fullMethodName).Inc()

	return ctx, nil
}
