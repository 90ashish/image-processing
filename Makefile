PROTO_DIR   := proto
GOOGLEAPIS  := third_party/googleapis
PROTO_FILES := $(PROTO_DIR)/image.proto

.PHONY: proto run-server run-gateway run-client check-server-health

# 1) gRPC stubs → go into proto/
proto:
	# ensure the gateway plugin is on your PATH:
	which protoc-gen-grpc-gateway >/dev/null

	protoc \
	  -I$(PROTO_DIR) \
	  -I$(GOOGLEAPIS) \
	  --go_out=paths=source_relative:$(PROTO_DIR) \
	  --go-grpc_out=paths=source_relative:$(PROTO_DIR) \
	  --grpc-gateway_out=paths=source_relative:$(PROTO_DIR) \
	  --grpc-gateway_opt=logtostderr=true \
	  $(PROTO_FILES)


# 2) gRPC-Gateway stubs → also go into proto/
proto-gateway:
	protoc \
	  -I$(PROTO_DIR) \
	  -I$(GOOGLEAPIS) \
	  --grpc-gateway_out=paths=source_relative:$(PROTO_DIR) \
	  --grpc-gateway_opt=logtostderr=true \
	  $(PROTO_DIR)/image.proto

run-server:
	@echo "Starting gRPC server..."
	go run server/*.go

run-gateway:
	@echo "Starting REST gateway..."
	go run gateway/main.go

run-client:
	@echo "Starting client..."
	go run client/main.go

check-server-health:
	grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check


## REST Curl CMD :=
# curl http://localhost:8080/v1/version

# curl -N -H "Accept: application/json" \
#     -X POST http://localhost:8080/v1/images/7d95f043-.../process \
#     -d '{}'

# curl -X POST http://localhost:8080/v1/images:upload \                                                                                                                           ─╯
#      --header "Content-Type: application/octet-stream" \
#      --data-binary @test.jpg
