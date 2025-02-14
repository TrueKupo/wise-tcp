# Lightweight Go builder image
FROM golang:1.23 as builder

# Set working directory
WORKDIR /app

# Copy source files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the app binary
RUN go build -o client ./cmd/client/client.go

# Production image
FROM debian:bookworm-slim

# Set working directory
WORKDIR /app

# Copy the binary
COPY --from=builder /app/client /app/client

# Copy the configuration
COPY cfg/client.yml /app/cfg/client.yml

# Set environment variables
ENV SERVER_ADDR="host.docker.internal:9001"
ENV TRY_REPLAY=false

# Run the app
ENTRYPOINT ["./client"]
