ARG GO_VERSION=1.25

FROM golang:${GO_VERSION} AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o vector-store-service .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/vector-store-service /app/vector-store-service
EXPOSE 3015
ENV VECTOR_STORE_PORT=3015
ENTRYPOINT ["/app/vector-store-service"]
