.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: test
test:
	@go test ./...

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: run
run:
	@go run ./cmd/server

.PHONY: check
check: fmt test
