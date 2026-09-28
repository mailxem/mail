# syntax=docker/dockerfile:1
# Compile on the runner's native architecture, including for ARM targets.
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies with caching
RUN go mod download && go mod verify

# Keep runtime assets and generated build metadata out of the compiler layers.
COPY cmd ./cmd
COPY internal ./internal
COPY docs/swagger ./docs/swagger

ARG TARGETOS
ARG TARGETARCH

# Go schedules package compilation in parallel and shares dependencies across binaries.
# GitHub Actions restores/exports this mount separately from the Docker layer cache.
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /app/build/ ./cmd ./cmd/helper ./cmd/sending-admin \
    && mv /app/build/cmd /app/build/posthoot

# Use a minimal runtime image for the target architecture.
FROM gcr.io/distroless/static-debian12:nonroot

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --chmod=755 --from=builder /app/build/posthoot .
COPY --chmod=755 --from=builder /app/build/helper .
COPY --chmod=755 --from=builder /app/build/sending-admin .
COPY --chmod=755 openapi.json .
COPY --chmod=755 public/build-info.txt /app/public/build-info.txt

# Copy template seeder data for Airley templates
# Source: /app/internal/models/seeder/airley/templates.json
# Destination: /app/internal/models/seeder/airley/templates.json
COPY --chmod=755 internal/models/seeder/airley/templates.json /app/internal/models/seeder/airley/

# Copy all initial setup seeder files for database initialization
# Source: /app/internal/models/seeder/initial-setup/*
# Destination: /app/internal/models/seeder/initial-setup/
COPY --chmod=755 internal/models/seeder/initial-setup/* /app/internal/models/seeder/initial-setup/

# Expose ports
EXPOSE 9001 587

# Set the entry point
CMD ["/app/posthoot"]
