FROM golang:1.23 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/llm-api ./cmd/server

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/llm-api /app/llm-api
COPY config /app/config
COPY docs /app/docs
WORKDIR /app
USER nonroot:nonroot
ENTRYPOINT ["/app/llm-api"]
