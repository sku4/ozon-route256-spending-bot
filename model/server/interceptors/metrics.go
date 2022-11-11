package interceptors

import (
	"context"
	"google.golang.org/grpc"
)

type MetricsFunc func(ctx context.Context, fullMethodName string) (context.Context, error)

// UnaryServerInterceptor returns a new unary server interceptors that performs per-request metrics
func UnaryServerInterceptor(metricsFunc MetricsFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var newCtx context.Context
		var err error
		newCtx, err = metricsFunc(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}
