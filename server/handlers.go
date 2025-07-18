package main

import (
	"context"
	"fmt"
	pb "image-proc/proto"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// Upload handles client-streaming image upload
func (s *server) Upload(stream pb.ImageProcessor_UploadServer) error {
	s.logger.Info("Upload Started")

	// genearte image ID and file path
	imgID := uuid.New().String()

	// write into project-level "uploads" directory
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return status.Errorf(codes.Internal, "failed to create upload dir: %v", err)
	}

	tmpPath := fmt.Sprintf("%s/%s.jpg", uploadDir, imgID)
	file, err := os.Create(tmpPath)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create file: %v", err)
	}
	defer file.Close()

	// receive chunks
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			s.logger.Infof("Upload completed: %s", tmpPath)
			return stream.SendAndClose(&pb.UploadResponse{ImageId: imgID})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "upload recv error: %v", err)
		}
		if _, err := file.Write(req.GetChunk()); err != nil {
			return status.Errorf(codes.Internal, "file write error: %v", err)
		}
	}
}

// Process simulates image processing and streams progress
func (s *server) Process(req *pb.ProcessingRequest, stream pb.ImageProcessor_ProcessServer) error {
	s.logger.Info("Starting processing %s with filters %v", req.ImageId, req.Filters)
	steps := 10
	for i := 0; i <= steps; i++ {
		time.Sleep(200 * time.Millisecond)
		pct := int32(i * 100 / steps)
		upd := &pb.ProgressUpdate{Percent: pct, Status: fmt.Sprintf("%d%% complete", pct)}
		if err := stream.Send(upd); err != nil {
			return status.Errorf(codes.Internal, "send error: %v", err)
		}
	}
	s.logger.Info("Proccessing Completed")
	return nil
}

// Tune handles bidirectional parameter tuning
func (s *server) Tune(stream pb.ImageProcessor_TuneServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.logger.Infof("Tune request: %s = %f on image %s", req.Parameter, req.Value, req.ImageId)
		// simulate the preview generation
		preview := []byte(fmt.Sprintf("Preview for %s: %s=%.2f", req.ImageId, req.Parameter, req.Value))
		if err := stream.Send(&pb.TuneResponse{PreviewChunk: preview}); err != nil {
			return err
		}
	}
}
