package consumer

import (
	"context"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler/consumer"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HistogramResponseTime = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "report",
			Subsystem: "http",
			Name:      "histogram_response_time_milliseconds",
			Buckets:   []float64{1, 5, 10, 50, 100, 500, 1000, 2000},
		},
	)
)

func MetricsMiddleware(next consumer.IHandler) consumer.IHandler {
	hr := consumer.Func(func(ctx context.Context, msg *sarama.ConsumerMessage) (err error) {
		startTime := time.Now()
		if err = next.ReportMessage(ctx, msg); err != nil {
			return err
		}
		duration := time.Since(startTime)

		HistogramResponseTime.Observe(float64(duration.Milliseconds()))

		return
	})

	return hr
}
