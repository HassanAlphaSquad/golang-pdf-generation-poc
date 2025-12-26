.PHONY: help run swagger build test clean install-deps docker-build docker-run example

# Variables
APP_NAME := pdf-api
VERSION := 1.0.0
PORT := 3000
OUTPUT_DIR := ./output

## help: Display help message
help:
	@echo "PDF Generation API - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo ""

## run: Start the API server with Swagger docs
run: swagger
	@echo "Starting API server on port $(PORT)..."
	@echo "Swagger docs: http://localhost:$(PORT)/swagger/index.html"
	@echo "Health check: http://localhost:$(PORT)/health"
	@PORT=$(PORT) OUTPUT_DIR=$(OUTPUT_DIR) go run cmd/api/main.go

## serve: Alias for 'make run'
serve: run

## swagger: Generate Swagger documentation
swagger:
	@if [ ! -f ~/go/bin/swag ]; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@~/go/bin/swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

## build: Build API server
build: swagger
	@go build -o bin/api -ldflags="-s -w" cmd/api/main.go


## test: Run all tests
test:
	@go test -v ./...

## test-short: Run unit tests only
test-short:
	@go test -short -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

## fmt: Format code
fmt:
	@go fmt ./...

## vet: Run go vet
vet:
	@go vet ./...

## lint: Run linter
lint:
	@golangci-lint run

## install-deps: Install dependencies
install-deps:
	@go mod download
	@go mod tidy

## install-tools: Install tools
install-tools:
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## clean: Clean build artifacts
clean:
	@rm -rf bin/ output/ coverage.out coverage.html *.pdf

## docker-build: Build Docker image
docker-build:
	@docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

## docker-run: Run Docker container
docker-run:
	@docker run -p $(PORT):3000 -v $(PWD)/output:/app/output \
		-e PORT=3000 -e OUTPUT_DIR=/app/output --name $(APP_NAME) --rm $(APP_NAME):latest

## docker-stop: Stop Docker container
docker-stop:
	@docker stop $(APP_NAME) || true

## check: Run fmt, vet, and test
check: fmt vet test

## info: Show project information
info:
	@echo "Project: $(APP_NAME) v$(VERSION)"
	@echo "Port: $(PORT)"
	@echo "Swagger: http://localhost:$(PORT)/swagger/index.html"

# Default target
.DEFAULT_GOAL := help
