// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: api.proto

package api

import (
	context "context"
	report "github.com/sku4/ozon-route256-spending-bot/pkg/api/report"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SpendingClient is the client API for Spending service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SpendingClient interface {
	// Sends a greeting
	SendReport(ctx context.Context, in *report.Report, opts ...grpc.CallOption) (*Empty, error)
}

type spendingClient struct {
	cc grpc.ClientConnInterface
}

func NewSpendingClient(cc grpc.ClientConnInterface) SpendingClient {
	return &spendingClient{cc}
}

func (c *spendingClient) SendReport(ctx context.Context, in *report.Report, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/api.Spending/SendReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SpendingServer is the server API for Spending service.
// All implementations should embed UnimplementedSpendingServer
// for forward compatibility
type SpendingServer interface {
	// Sends a greeting
	SendReport(context.Context, *report.Report) (*Empty, error)
}

// UnimplementedSpendingServer should be embedded to have forward compatible implementations.
type UnimplementedSpendingServer struct {
}

func (UnimplementedSpendingServer) SendReport(context.Context, *report.Report) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendReport not implemented")
}

// UnsafeSpendingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SpendingServer will
// result in compilation errors.
type UnsafeSpendingServer interface {
	mustEmbedUnimplementedSpendingServer()
}

func RegisterSpendingServer(s grpc.ServiceRegistrar, srv SpendingServer) {
	s.RegisterService(&Spending_ServiceDesc, srv)
}

func _Spending_SendReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(report.Report)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SpendingServer).SendReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.Spending/SendReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SpendingServer).SendReport(ctx, req.(*report.Report))
	}
	return interceptor(ctx, in, info, handler)
}

// Spending_ServiceDesc is the grpc.ServiceDesc for Spending service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Spending_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.Spending",
	HandlerType: (*SpendingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendReport",
			Handler:    _Spending_SendReport_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
