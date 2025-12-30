ARG GO_VERSION=1.25

FROM registry.menlo.ai/dockerhub/golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-tools .

# Runtime stage
FROM registry.menlo.ai/dockerhub/alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/mcp-tools .

EXPOSE 8091

CMD ["./mcp-tools"]
