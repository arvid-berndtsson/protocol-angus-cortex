# Multi-stage build for Protocol Argus Cortex
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o protocol-argus-cortex ./cmd/protocol-argus-cortex/

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata libpcap

# Create non-root user
RUN addgroup -g 1001 -S argus && \
    adduser -u 1001 -S argus -G argus

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/protocol-argus-cortex .

# Copy configuration example
COPY --from=builder /app/config.yml.example .

# Create models directory
RUN mkdir -p models && chown -R argus:argus /app

# Switch to non-root user
USER argus

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["./protocol-argus-cortex", "--config", "config.yml"] 