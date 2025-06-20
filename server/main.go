package main

import (
	"context"
	pb "image-proc/proto"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

// server implements the ImageProcessor service
type server struct {
	version string
	logger  *zap.SugaredLogger
	pb.UnimplementedImageProcessorServer
}

// GetVersion returns a static version string
func (s *server) GetVersion(ctx context.Context, _ *emptypb.Empty) (*pb.VersionResponse, error) {
	s.logger.Infof("GetVersion called at %s", time.Now().Format(time.RFC3339))
	return &pb.VersionResponse{Version: s.version}, nil
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	suggar := logger.Sugar()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		suggar.Fatalf("failed to listen: %v", err)
	}
	suggar.Infof("gRPC server listening on %s", lis.Addr())

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
