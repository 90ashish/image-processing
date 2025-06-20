# Makefile for generating Go code from .proto files
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

.PHONY: proto 


proto: ## generate proto files
	@echo "Generating Go code from .proto files..."
	protoc \
	--go_out=paths=source_relative:. \
	--go-grpc_out=paths=source_relative:. \
	proto/image.proto

run-client:
	@echo "Starting client..."
	go run client/main.go -file=./test.jpg

run-server:
	@echo "Starting server..."
	go run server/*