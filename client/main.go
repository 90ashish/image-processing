package main

import (
	"context"
	"flag"
	pb "image-proc/proto"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	suggar := logger.Sugar()

	addr := flag.String("addr", "localhost:50051", "grpc server address")
	flag.Parse()

	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		dialCtx,
		*addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.Config{
			BaseDelay:  1 * time.Second,
			Multiplier: 1.6,
			Jitter:     0.2,
			MaxDelay:   5 * time.Second,
		}}),
	)
	if err != nil {
		suggar.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewImageProcessorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetVersion(ctx, &emptypb.Empty{})
	if err != nil {
		suggar.Fatalf("GetVersion error: %v", err)
	}
	suggar.Infof("Service version: %s", resp.GetVersion())
}
