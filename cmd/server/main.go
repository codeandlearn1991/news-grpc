package main

import (
	"context"
	"log"
	"net"
	"time"

	newsv1 "github.com/codeandlearn1991/news-grpc/api/news/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"

	ingrpc "github.com/codeandlearn1991/news-grpc/internal/grpc"
	"github.com/codeandlearn1991/news-grpc/internal/memstore"
)

func main() {
	lis, err := net.Listen("tcp", ":50051") //nolint:gosec // Okay for the project
	if err != nil {
		log.Fatalf("failed to list: %v", err)
	}

	// Interceptors essentially wrapp the gRPC handler.
	// Request -> Interceptor -> Interceptor or the gRPC handler.
	//
	// 1. Unary
	// 2. Server Side Stream
	// 3. Client Side Stream
	// 4. Bidirectional Stream
	//
	// 1. Server Side Unary Interceptor -> Server Unary Calls only
	// 2. Client Side Unary Interceptor -> Client Unary Calls only
	// 3. Server Side Stream Interceptor -> Server Streaming Calls only
	// 4. Client Side Stream Interceptor -> Client Streaming Calls only

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				start := time.Now()
				response, err := handler(ctx, req)
				log.Printf("unary call made with: %+v", info)
				log.Printf("time taken: %s", time.Since(start))
				return response, err
			},
			func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				log.Println("second interceptor")
				return handler(ctx, req)
			},
		),
		grpc.StreamInterceptor(func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			log.Println("Server side streaming interceptor")
			return handler(srv, ss)
		},
		),
	)
	newsv1.RegisterNewsServiceServer(srv, ingrpc.NewServer(memstore.New()))
	healthSrv := health.NewServer()
	healthv1.RegisterHealthServer(srv, healthSrv)

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
