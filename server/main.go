package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	pb "image-proc/proto"
	"io/ioutil"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load server cert & key
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		sugar.Fatalf("failed to load server key pair: %v", err)
	}
	// Load CA cert for client verification (mTLS)
	caPem, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		sugar.Fatalf("failed to read CA cert: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caPem)

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		// To enforce mTLS, uncomment:
		// ClientAuth: tls.RequireAndVerifyClientCert,
	}
	creds := credentials.NewTLS(tlsCfg)

	// Start listening
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		sugar.Fatalf("failed to listen: %v", err)
	}
	sugar.Infof("gRPC server listening on %s (TLS)", lis.Addr())

	// Health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Build gRPC server with TLS and interceptors
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(loggingUnaryInterceptor(sugar)),
		grpc.StreamInterceptor(loggingStreamInterceptor(sugar)),
	)

	// Register service, health & reflection
	pb.RegisterImageProcessorServer(grpcServer, &server{
		version: "v0.1.0",
		logger:  sugar,
	})
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	// Serve
	if err := grpcServer.Serve(lis); err != nil {
		sugar.Fatalf("serve error: %v", err)
	}
}
