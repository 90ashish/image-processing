package main

import (
	"context"
	pb "image-proc/proto"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// loggingUnaryInterceptor logs each unary RPC’s method, duration, and error.
func loggingUnaryInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logger.Infof("[UNARY] %s took %s; err=%v", info.FullMethod, time.Since(start), err)
		return resp, err
	}
}

// loggingStreamInterceptor logs each streaming RPC’s method, duration, and error.
func loggingStreamInterceptor(logger *zap.SugaredLogger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		logger.Infof("[STREAM] %s took %s; err=%v", info.FullMethod, time.Since(start), err)
		return err
	}
}

func main() {
	// logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Health service reports SERVING
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// listen
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		sugar.Fatalf("failed to listen: %v", err)
	}
	sugar.Infof("gRPC server listening on %s", lis.Addr())

	// Build gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor(sugar)),
		grpc.StreamInterceptor(loggingStreamInterceptor(sugar)),
	)

	// Register our ImageProcessor service
	pb.RegisterImageProcessorServer(grpcServer, &server{
		version: "v0.1.0",
		logger:  sugar,
	})

	// Register health and reflection for introspection
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	// Serve
	if err := grpcServer.Serve(lis); err != nil {
		sugar.Fatalf("serve error: %v", err)
	}
}
