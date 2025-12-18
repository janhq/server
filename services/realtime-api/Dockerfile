ARG GO_VERSION=1.25

FROM golang:${GO_VERSION} as build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/realtime-api ./cmd/server

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/realtime-api /realtime-api
EXPOSE 8186
ENTRYPOINT ["/realtime-api"]
