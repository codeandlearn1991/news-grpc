package main

import (
	"log"
	"net"

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

	srv := grpc.NewServer()
	newsv1.RegisterNewsServiceServer(srv, ingrpc.NewServer(memstore.New()))
	healthSrv := health.NewServer()
	healthv1.RegisterHealthServer(srv, healthSrv)

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
