package server

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/skubach/workshop-1-bot/internal/handler"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/api"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type Grpc struct {
	ctx         context.Context
	grpcService *grpc.Server
	handler     *handler.Handler
}

// NewGrpc created new grpc server
func NewGrpc(ctx context.Context, handler *handler.Handler) *Grpc {
	return &Grpc{
		ctx: ctx,
		// grpc middleware
		grpcService: grpc.NewServer(
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_auth.UnaryServerInterceptor(handler.Auth),
			)),
		),
		handler: handler,
	}
}

// Run grpc on port with handler
func (g *Grpc) Run(port int) (err error) {
	api.RegisterMnemosyneServer(g.grpcService, g.handler)
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
