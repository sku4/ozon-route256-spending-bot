package server

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	hgrpc "github.com/sku4/ozon-route256-spending-bot/internal/handler/grpc"
	"github.com/sku4/ozon-route256-spending-bot/model/server/interceptors"
	"github.com/sku4/ozon-route256-spending-bot/pkg/api"
	"github.com/sku4/ozon-route256-spending-bot/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Grpc struct {
	ctx         context.Context
	grpcService *grpc.Server
	handler     *hgrpc.Handler
}

// NewGrpc created new grpc server
func NewGrpc(ctx context.Context, handler *hgrpc.Handler) *Grpc {
	return &Grpc{
		ctx: ctx,
		// grpc middleware metrics
		grpcService: grpc.NewServer(
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				interceptors.UnaryServerInterceptor(handler.Metrics),
			)),
		),
		handler: handler,
	}
}

// Run grpc on port with handler
func (g *Grpc) Run(grpcUrl string) (err error) {
	api.RegisterSpendingServer(g.grpcService, g.handler)
	lis, err := net.Listen("tcp", grpcUrl)
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}
	logger.Info(fmt.Sprintf("gRPC server is listening on %s", grpcUrl))
	reflection.Register(g.grpcService)
	if err = g.grpcService.Serve(lis); err != nil {
		return errors.Wrap(err, "failed to serve")
	}

	return
}

// GracefulStop grpc service
func (g *Grpc) GracefulStop() {
	g.grpcService.GracefulStop()
}
