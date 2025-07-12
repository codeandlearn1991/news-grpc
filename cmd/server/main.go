package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	newsv1 "github.com/codeandlearn1991/news-grpc/api/news/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"

	ingrpc "github.com/codeandlearn1991/news-grpc/internal/grpc"
	"github.com/codeandlearn1991/news-grpc/internal/memstore"

	"buf.build/go/protovalidate"
	protovalidate_interceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"golang.org/x/sync/errgroup"
)

func main() {
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

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("validator initialization: %v", err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(protovalidate_interceptor.UnaryServerInterceptor(validator)),
		grpc.ChainStreamInterceptor(protovalidate_interceptor.StreamServerInterceptor(validator)),
	)
	newsv1.RegisterNewsServiceServer(srv, ingrpc.NewServer(memstore.New()))
	healthSrv := health.NewServer()
	healthv1.RegisterHealthServer(srv, healthSrv)

	grp, grpCtx := errgroup.WithContext(context.Background())

	grp.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %w", err)
			}
		}()

		lis, err := net.Listen("tcp", ":50051") //nolint:gosec // Okay for the project
		if err != nil {
			err = fmt.Errorf("failed to list: %w", err)
		}

		if listErr := srv.Serve(lis); listErr != nil {
			err = fmt.Errorf("failed to serve: %w", listErr)
		}

		return err
	})

	grp.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %w", err)
			}
		}()
		interceptSignals(grpCtx)
		healthSrv.Shutdown()
		return shutdown(grpCtx, srv)
	})

	if err := grp.Wait(); err != nil {
		log.Fatal("server shutdown", err)
	}
}

func interceptSignals(ctx context.Context) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-ctx.Done():
		return
	case sig := <-sigc:
		log.Println("intercepted signal: ", sig.String())
		return
	}
}

func shutdown(ctx context.Context, srv *grpc.Server) (err error) {
	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		err = fmt.Errorf("grpc server forcibly shutdown: %w", ctx.Err())
		srv.Stop()
	}

	return err
}
