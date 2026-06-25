.PHONY: run build test lint clean

# Development
run:
	go run ./cmd/server

# Production build — stripped binary for smaller size
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# Run tests with race detection
test:
	go test ./... -race -v

# Lint (requires golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/