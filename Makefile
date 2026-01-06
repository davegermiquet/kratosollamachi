.PHONY: help run build test clean dev install lint coverage test-verbose test-race test-integration

# Default target
help:
	@echo "Available commands:"
	@echo "  make run              - Run the application"
	@echo "  make build            - Build the application"
	@echo "  make dev              - Run with auto-reload (requires air)"
	@echo "  make test             - Run tests"
	@echo "  make test-verbose     - Run tests with verbose output"
	@echo "  make test-race        - Run tests with race detector"
	@echo "  make test-integration - Run integration tests (requires Kratos)"
	@echo "  make coverage         - Run tests with coverage report"
	@echo "  make lint             - Run linter (requires golangci-lint)"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make install          - Install dependencies"
	@echo "  make check            - Run all checks (lint + test)"

# Run the application
run:
	go run cmd/server/main.go

# Build the application
build:
	go build -o server cmd/server/main.go

# Development mode with hot reload
dev:
	air

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -race ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | tail -1

# Run integration tests (requires Kratos to be running)
test-integration:
	@echo "Running integration tests (requires Ory Kratos to be running)..."
	INTEGRATION_TEST=true go test -v -tags=integration -timeout 30s -run TestUserAuthenticationFlow

# Run linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f server
	rm -f coverage.out coverage.html
	go clean

# Install dependencies
install:
	go mod download
	go mod tidy

# Run all checks
check: lint test

# Generate mocks (if using mockgen)
mocks:
	@echo "Generating mocks..."
	# Add mockgen commands here if needed

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Verify dependencies
verify:
	go mod verify

# Update dependencies
update:
	go get -u ./...
	go mod tidy
