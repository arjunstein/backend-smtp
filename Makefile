.PHONY: all build run test tidy swagger clean help

all: build

swagger:
	@echo "==> Generating Swagger documentation..."
	@go run github.com/swaggo/swag/cmd/swag@v1.8.1 init -g cmd/api/main.go -o docs

build:
	@echo "==> Building binary..."
	@go build -o bin/backend-smtp cmd/api/main.go

run:
	@echo "==> Running server..."
	@go run cmd/api/main.go

test:
	@echo "==> Running tests..."
	@go test -v -race ./...

tidy:
	@echo "==> Tidying dependencies..."
	@go mod tidy

clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf bin/

help:
	@echo "Available commands:"
	@echo "  make build    - Compile the binary to bin/backend-smtp"
	@echo "  make run      - Run API server locally"
	@echo "  make swagger  - Regenerate Swagger OpenAPI docs"
	@echo "  make test     - Run unit and race-condition tests"
	@echo "  make tidy     - Format and clean up go.mod and go.sum"
	@echo "  make clean    - Remove built binaries"
