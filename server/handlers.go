package main

import (
	"context"
	"fmt"
	pb "image-proc/proto"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

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
