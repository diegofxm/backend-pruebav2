# Build stage
FROM golang:1.23 AS builder

# Install git and ca-certificates
RUN apt-get update && apt-get install -y git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM debian:bullseye-slim

# Install ca-certificates, timezone data, and wget
RUN apt-get update && apt-get install -y ca-certificates tzdata wget && rm -rf /var/lib/apt/lists/*

# Set timezone to Colombia
ENV TZ=America/Bogota

# Create app directory
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy .env file if exists
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./main"]