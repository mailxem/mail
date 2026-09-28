FROM golang:1.25-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o payments ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates
RUN apk --no-cache add ca-certificates curl bash

WORKDIR /code

# Copy the binary from builder stage
COPY --from=builder /app/payments .
COPY --from=builder /app/scripts/seed /code/scripts/seed

# Expose port
EXPOSE 8080

CMD ["./payments"]
