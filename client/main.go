package main

import (
	"context"
	"flag"
	"fmt"
	pb "image-proc/proto"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/protobuf/types/known/emptypb"
)

// getVersion invokes the unary GetVersion RPC
func getVersion(client pb.ImageProcessorClient, sugar *zap.SugaredLogger) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetVersion(ctx, &emptypb.Empty{})
	if err != nil {
		sugar.Fatalf("GetVersion failed: %v", err)
	}
	sugar.Infof("Service version: %s", resp.GetVersion())
}

// uploadFile streams the file contents via the Upload RPC
func uploadFile(client pb.ImageProcessorClient, filePath string, sugar *zap.SugaredLogger) string {
	sugar.Infof("Starting upload for %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		sugar.Fatalf("file open error: %v", err)
	}
	defer file.Close()

	stream, err := client.Upload(context.Background())
	if err != nil {
		sugar.Fatalf("Upload init error: %v", err)
	}

	buf := make([]byte, 64*1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			sugar.Fatalf("file read error: %v", err)
		}
		if err := stream.Send(&pb.UploadRequest{Chunk: buf[:n]}); err != nil {
			sugar.Fatalf("chunk send error: %v", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		sugar.Fatalf("Upload failed: %v", err)
	}
	fmt.Printf("Uploaded image ID: %s\n", resp.GetImageId())
	return resp.GetImageId()
}

func processImage(client pb.ImageProcessorClient, imageID string, filters []string, sugar *zap.SugaredLogger) {
	sugar.Info("Processing with %s with %v", imageID, filters)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.Process(ctx, &pb.ProcessingRequest{ImageId: imageID, Filters: filters})
	if err != nil {
		sugar.Fatalf("process init: %v", err)
	}

	for {
		upd, err := stream.Recv()
		if err == io.EOF {
			sugar.Info("Proccessing Done")
			break
		}
		if err != nil {
			sugar.Fatalf("Process recv: %v", err)
		}
		sugar.Infof("Progress %d%% - %s", upd.GetPercent(), upd.GetStatus())
	}
}

// tuneImage opens a bidirectional Tune stream
func tuneImage(client pb.ImageProcessorClient, imageID string, params []string, sugar *zap.SugaredLogger) {
	stream, err := client.Tune(context.Background())
	if err != nil {
		sugar.Fatalf("Tune init error: %v", err)
	}

	// recieve loop
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				sugar.Fatalf("Tune recv error: %v", err)
			}
			sugar.Infof("Received preview chunk: %s", string(resp.PreviewChunk))
		}
	}()

	// send loop
	for _, p := range params {
		parts := strings.SplitN(p, ":", 2)
		val, _ := strconv.ParseFloat(parts[1], 64)
		req := &pb.TuneRequest{ImageId: imageID, Parameter: parts[0], Value: val}
		if err := stream.Send(req); err != nil {
			sugar.Fatalf("Tune send error: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	stream.CloseSend()
}

func main() {
	// command-line flags
	addr := "localhost:50051"
	filePath := "./test.jpg"
	doProcess := true
	processFilters := []string{"blur", "edge"}
	doTune := true
	tuneParams := []string{"brightness:1.2", "contrast:0.8"}
	flag.Parse()

	// initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// set up connection
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		dialCtx,
		addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.Config{
			BaseDelay:  time.Second,
			Multiplier: 1.6,
			Jitter:     0.2,
			MaxDelay:   5 * time.Second,
		}}),
	)
	if err != nil {
		sugar.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewImageProcessorClient(conn)

	// Phase 1: GetVersion
	getVersion(client, sugar)

	// Phase 2: Upload if file flag provided
	imgID := uploadFile(client, filePath, sugar)

	// Phase 3
	if doProcess {
		processImage(client, imgID, processFilters, sugar)
	}

	// Phase 4
	if doTune {
		tuneImage(client, imgID, tuneParams, sugar)
	}
}
