# Lightweight Go builder image
FROM golang:1.23 as builder

# Set working directory
WORKDIR /app

# Copy source files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the app binary
RUN go build -o server ./cmd/server/server.go

# Production image
FROM debian:bookworm-slim

# Install CA certificates
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy the binary
COPY --from=builder /app/server /app/server

# Copy the configuration
COPY cfg/server.yml /app/cfg/server.yml

# Expose the configured port
EXPOSE 9001

# Set environment variables
ENV PORT=9001
ENV MAX_CONN=2
ENV POW_DIFFICULTY=20

# Run the app
ENTRYPOINT ["./server"]
