# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o streaming-api cmd/server/main.go

# Stage 2: Final runtime image
FROM alpine:latest

# Install ffmpeg for transcoding
RUN apk add --no-cache ffmpeg

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/streaming-api .

# Create directory for storage
RUN mkdir -p /app/storage

# Expose HTTP port
EXPOSE 8080

# Run the binary
CMD ["./streaming-api"]
