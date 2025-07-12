.PHONY: all test unit-test integration-test lint build clean coverage install-tools

# Default target
all: lint test build

# Run all tests
test: unit-test

# Run unit tests with coverage
unit-test:
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out ./pkg/... ./internal/...

# Run integration tests (if they exist)
integration-test:
	@echo "Running integration tests..."
	@if [ -d "./tests/integration" ]; then \
		go test -v -race ./tests/integration/...; \
	else \
		echo "No integration tests found (./tests/integration directory doesn't exist)"; \
	fi

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Build all packages to verify compilation
build:
	@echo "Building..."
	@go build ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out

# Generate and open coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run tests with verbose output
test-verbose:
	@go test -v ./...

# Quick test without race detector
test-quick:
	@go test ./...