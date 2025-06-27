package main

import (
	"context"
	"flag"
	pb "image-proc/proto"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	grpcEndpoint := flag.String("grpc-endpoint", "localhost:50051", "gRPC server address")
	httpPort := flag.String("http-port", ":8080", "HTTP listen port")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		// NOTE: in Phase 6 you’d swap this for TLS creds
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if err := pb.RegisterImageProcessorHandlerFromEndpoint(ctx, mux, *grpcEndpoint, opts); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	log.Printf("REST gateway listening on %s", *httpPort)
	if err := http.ListenAndServe(*httpPort, mux); err != nil {
		log.Fatalf("gateway ListenAndServe: %v", err)
	}
}
