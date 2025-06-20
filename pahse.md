Project Phases Overview

    Phase 1: Project scaffolding & Unary RPC
        The server log shows the GetVersion handler being invoked.
        The client prints out the service version.

    Phase 2: Client‐streaming File Upload
        Goals:
        Extend proto/image.proto with a Upload RPC that accepts a stream of byte chunks.

        Server:
        In handlers.go, receive the stream, assemble the chunks into a file, write it to disk (e.g. /tmp/<uuid>.jpg), and return an UploadResponse{image_id}.

        Client:
        In client/main.go, read a local image file in, say, 64 KB chunks, send each chunk on the gRPC stream, then CloseAndRecv() the server’s response.

        Verification:
        After the client finishes, verify the file exists on disk under the generated image ID and matches the original.

    Phase 3: Server‐streaming “Process”

    Phase 4: Bidirectional “Tune”

    Phase 5: Interceptors, Health & Reflection

    Phase 6: TLS/mTLS Security

    Phase 7: REST Gateway & gRPC-Web

    Phase 8: Service Discovery & Load-Balancing

    Phase 9: Observability (Prometheus & Tracing)

    Phase 10: Docker Compose & Kubernetes Deployment