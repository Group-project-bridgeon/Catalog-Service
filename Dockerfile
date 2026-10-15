# syntax=docker/dockerfile:1

# =========================================================
# Build stage
# =========================================================
FROM golang:1.25.4-alpine AS builder

WORKDIR /src

# Copy dependency files first so Docker can cache this layer.
COPY go.mod go.sum ./

# Download and verify dependencies using BuildKit cache.
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy only the source required to build the service.
COPY cmd ./cmd
COPY internal ./internal

# Build a static Linux binary.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/catalog-service \
    ./cmd/server


# =========================================================
# Runtime stage
# =========================================================
FROM alpine:3.22

# Runtime dependencies + non-root user.
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app -H -s /sbin/nologin app

WORKDIR /app

COPY --from=builder /out/catalog-service /app/catalog-service

ENV PORT=8081 \
    GIN_MODE=release

USER app

EXPOSE 8081

# Liveness check: confirms the Catalog process is alive.
HEALTHCHECK \
    --interval=15s \
    --timeout=3s \
    --start-period=10s \
    --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT}/health" > /dev/null || exit 1

# Exec form keeps the Go process as PID 1,
# so SIGTERM reaches the app for graceful shutdown.
ENTRYPOINT ["/app/catalog-service"]