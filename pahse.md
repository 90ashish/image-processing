# Project Phases Overview

---

## Phase 1: Project Scaffolding & Unary RPC

**Goal**: Establish basic gRPC project structure and implement a simple unary RPC.

### Tasks:
- Scaffold repository layout (`go.mod`, `proto/`, `server/`, `client/`)
- Define `GetVersion` RPC in `image.proto` with `VersionResponse` message
- Generate Go stubs using `protoc`
- Implement `GetVersion` handler in `server/handlers.go`, returning a static version string
- Create `client/main.go` to dial the server and call `GetVersion`, printing the response
- Enable server reflection in `server/main.go` for introspection

### Expected Outcomes & Verification:
- Server logs show: `GetVersion called at <timestamp>`
- Client prints: `Service version: v0.1.0`
- No compilation or runtime errors

---

## Phase 2: Client-Streaming File Upload

**Goal**: Support uploading large files in chunks via a client-streaming RPC.

### Proto Changes:
- Add `rpc Upload(stream UploadRequest) returns (UploadResponse)`
- Define `UploadRequest { bytes chunk }` and `UploadResponse { string image_id }`

### Server:
- In `handlers.go`, implement `Upload`:
  - Create project-local `uploads/` directory
  - Generate a UUID for the image ID
  - Receive streamed `UploadRequest` messages, writing each chunk to disk as `<image_id>.jpg`
  - On EOF, close and return `UploadResponse { image_id }`

### Client:
- In `client/main.go`, add `-file` flag
- Implement `uploadFile(client, filePath)`:
  - Open the specified file
  - Read and send in 64 KB chunks via `stream.Send(UploadRequest{chunk})`
  - Call `CloseAndRecv()` and print returned `image_id`

### Expected Outcomes & Verification:
- Client logs: `Starting upload for ./test.jpg`, then `Uploaded image ID: <uuid>`
- Server logs: `Upload started`, `Upload completed: uploads/<uuid>.jpg`
- Verify file on disk: file exists and matches original (compare byte-to-byte)

---

## Phase 3: Server-Streaming “Process”

**Goal**: Simulate image processing and stream progress updates to the client.

### Proto Changes:
- Add `rpc Process(ProcessingRequest) returns (stream ProgressUpdate)`
- Define `ProcessingRequest { string image_id; repeated string filters }`
- Define `ProgressUpdate { int32 percent; string status }`

### Server:
- Implement `Process` in `handlers.go`:
  - Log start with `image_id` and `filters`
  - Loop over fixed number of steps (e.g., 10), sleeping between each
  - Calculate percent complete and send `ProgressUpdate` on stream
  - Log completion

### Client:
- Add `-process` (bool) and `-filters` (comma-separated) flags
- Implement `processImage(client, imageId, filters)`:
  - Call `Process` and receive a stream
  - Loop over `stream.Recv()`, logging each `percent` and `status` until EOF

### Makefile:
- Extend `run-client` to include `-process -filters=blur,edge`

### Expected Outcomes & Verification:
- Client logs 10 updates from `Progress 10% - 10% complete` to `Progress 100% - 100% complete`
- Server logs matching start and completion messages

---

## Phase 4: Bidirectional “Tune”

**Goal**: Enable real-time parameter tuning with live preview data via bidirectional streaming.

### Proto Changes:
- Add `rpc Tune(stream TuneRequest) returns (stream TuneResponse)`
- Define `TuneRequest { string image_id; string parameter; double value }`
- Define `TuneResponse { bytes preview_chunk }`

### Server:
- Implement `Tune` in `handlers.go`:
  - Loop receiving `TuneRequest` messages until EOF
  - For each request, log the tuning change
  - Generate a dummy preview (byte slice) and send back in `TuneResponse`

### Client:
- Add `-tune` (bool) and `-tune-params` (comma-separated `param:value`) flags
- Implement `tuneImage(client, imageId, params)`:
  - Open bidirectional stream via `client.Tune()`
  - Launch goroutine to receive and handle `TuneResponse` chunks
  - Send `TuneRequest` messages parsed from `-tune-params` at intervals
  - After sending all, call `stream.CloseSend()`

### Expected Outcomes & Verification:
- Client logs confirmation of each preview chunk received
- Server logs each `Tune` request’s parameter change

---

## Phase 5: Interceptors, Health & Reflection

**Goal**: Add middleware, health checks, and reflection for production readiness.

### Tasks:
- Unary & Stream Interceptors: Logging, authentication, metrics (e.g., Zap, Prometheus)
- Health Checking: Register gRPC Health service so K8s can probe status
- Reflection: Ensure reflection remains enabled for debugging

### Expected Outcomes & Verification:
- Health endpoint responds `SERVING`
- Interceptor logs appear for each RPC

---

## Phase 6: TLS/mTLS Security

**Goal**: Secure communications with TLS and optional client certificate authentication.

### Tasks:
- Generate self-signed CA, server, and client certificates
- Configure `grpc.Creds` on server and `grpc.WithTransportCredentials` on client
- (Optional) Enforce mTLS by requiring client certs

### Expected Outcomes & Verification:
- Client-server communication over HTTPS (no `WithInsecure`)
- Verify invalid certs are rejected

---

## Phase 7: REST Gateway & gRPC-Web

**Goal**: Expose gRPC services as REST/JSON and support browser clients.

### Tasks:
- Add `google.api.http` annotations in proto
- Generate and run gRPC-Gateway in a separate binary
- Configure Envoy or grpc-web proxy for browser access

### Expected Outcomes & Verification:
- REST endpoints accessible via HTTP clients
- Browser page can call gRPC-Web API

---

## Phase 8: Service Discovery & Load-Balancing

**Goal**: Enable dynamic endpoint resolution and client-side load balancing.

### Tasks:
- Integrate with Consul/etcd or DNS SRV for service registry
- Use gRPC name resolver and load-balancing policies (`round_robin`)

### Expected Outcomes & Verification:
- Client distributes calls across multiple server instances

---

## Phase 9: Observability (Prometheus & Tracing)

**Goal**: Instrument services for metrics and distributed tracing.

### Tasks:
- Use `grpc-prometheus` interceptor to expose `/metrics`
- Add OpenTelemetry interceptors for tracing

### Expected Outcomes & Verification:
- Metrics visible in Prometheus
- Traces visible in Jaeger or similar systems

---

## Phase 10: Docker Compose & Kubernetes Deployment

**Goal**: Containerize and deploy services in a container orchestration platform.

### Tasks:
- Write multi-stage Dockerfiles for server and gateway
- Compose services with `docker-compose.yaml` (server, gateway, Envoy, Prometheus, Consul)
- Create Kubernetes manifests or Helm charts with health/readiness probes and autoscaling

### Expected Outcomes & Verification:
- Services run in Docker and Kubernetes
- Automatic scaling and health checks in place
