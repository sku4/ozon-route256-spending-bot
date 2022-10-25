package server

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	InFlightRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "bot",
		Subsystem: "http",
		Name:      "in_flight_requests_total",
	})
	SummaryResponseTime = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "bot",
		Subsystem: "http",
		Name:      "summary_response_time_seconds",
		Objectives: map[float64]float64{
			0.5:  0.1,
			0.9:  0.01,
			0.99: 0.001,
		},
	})
	HistogramResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "bot",
			Subsystem: "http",
			Name:      "histogram_response_time_seconds",
			Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2},
		},
		[]string{"cmd"},
	)
)

func MetricsMiddleware(next handler.IHandler) handler.IHandler {
	hr := handler.Func(func(ctx context.Context, upd tgbotapi.Update) (err error) {
		startTime := time.Now()
		if err = next.IncomingMessage(ctx, upd); err != nil {
			return err
		}
		duration := time.Since(startTime)

		SummaryResponseTime.Observe(duration.Seconds())

		cmd := ""
		if upd.Message != nil {
			if upd.Message.IsCommand() {
				cmd = upd.Message.Command()
			}
		}
		if cmd != "" {
			HistogramResponseTime.
				WithLabelValues(cmd).
				Observe(duration.Seconds())
		}

		return
	})
	wrappedHandler := InstrumentHandlerInFlight(InFlightRequests, hr)

	return wrappedHandler
}

func InstrumentHandlerInFlight(g prometheus.Gauge, next handler.IHandler) handler.IHandler {
	return handler.Func(func(ctx context.Context, upd tgbotapi.Update) (err error) {
		g.Inc()
		defer g.Dec()
		if err = next.IncomingMessage(ctx, upd); err != nil {
			return err
		}

		return
	})
}
