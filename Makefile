.PHONY: help run build test clean dev install

help:
	@echo "Available commands:"
	@echo "  make run       - Run the application"
	@echo "  make build     - Build the application"
	@echo "  make dev       - Run with auto-reload (requires air)"
	@echo "  make test      - Run tests"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make install   - Install dependencies"

run:
	go run cmd/server/main.go

build:
	go build -o server cmd/server/main.go

dev:
	air

test:
	go test -v ./...

clean:
	rm -f server
	go clean

install:
	go mod download
	go mod tidy
