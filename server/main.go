package main

import (
	pb "image-proc/proto"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server implements the ImageProcessor service
type server struct {
	version string
	logger  *zap.SugaredLogger
	pb.UnimplementedImageProcessorServer
}

func main() {
	// logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	suggar := logger.Sugar()

	// listen
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		suggar.Fatalf("failed to listen: %v", err)
	}
	suggar.Infof("gRPC server listening on %s", lis.Addr())

	// server
	grpcServer := grpc.NewServer()
	pb.RegisterImageProcessorServer(grpcServer, &server{
		version: "v0.1.0",
		logger:  suggar,
	})
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		suggar.Fatalf("Failed to serve: %v", err)
	}
}
