package server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	hgrpc "gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler/grpc"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
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
		// grpc middleware
		grpcService: grpc.NewServer(
		/*grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_auth.UnaryServerInterceptor(handler.Auth),
		)),*/
		),
		handler: handler,
	}
}

// Run grpc on port with handler
func (g *Grpc) Run(port int) (err error) {
	api.RegisterSpendingServer(g.grpcService, g.handler)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}
	logger.Info(fmt.Sprintf("gRPC server is listening on: %d", port))
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
