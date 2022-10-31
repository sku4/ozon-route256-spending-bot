package logger

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"log"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func init() {
	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	localLogger, err := cfg.Build()
	if err != nil {
		log.Fatal("logger init", err)
	}
	logger = localLogger
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func SInfo(args ...interface{}) {
	sugar := logger.Sugar()
	sugar.Info(args...)
}

func Infof(template string, args ...interface{}) {
	sugar := logger.Sugar()
	sugar.Infof(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	sugar := logger.Sugar()
	sugar.Fatalf(template, args...)
}

func WithTrace(ctx context.Context) *zap.Logger {
	localLogger := logger
	if spanContext, ok := opentracing.SpanFromContext(ctx).Context().(jaeger.SpanContext); ok {
		if traceID := spanContext.TraceID(); traceID.IsValid() {
			localLogger = localLogger.With(zap.String("trace-id", traceID.String()))
		}
		if spanID := spanContext.SpanID(); spanID > 0 {
			localLogger = localLogger.With(zap.String("span-id", spanID.String()))
		}
	}

	return localLogger
}
