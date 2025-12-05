# Build stage
ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o memory-tools ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 10001 appuser && \
    adduser -D -u 10001 -G appuser appuser

WORKDIR /app

# Create cache directory with proper permissions
RUN mkdir -p /app/.cache && chown appuser:appuser /app/.cache

# Copy the binary from builder
COPY --from=builder /app/memory-tools .

# Copy config and migrations
COPY configs ./config
COPY migrations ./migrations

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -fsS http://localhost:8090/healthz || exit 1

# Run the binary
CMD ["./memory-tools"]
