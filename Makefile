# Makefile
.PHONY: build run test clean docker-up docker-down swagger help deps fmt lint
	# adding help that goes through all targets
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':'

# Build all services
## build: Build all service 
build:
	@echo "Building all services..."
	@cd user-ms && go build -o bin/user-service cmd/main.go
	@cd product-ms && go build -o bin/product-service cmd/main.go
	@cd order-ms && go build -o bin/order-service cmd/main.go
	@cd payment-ms && go build -o bin/payment-service cmd/main.go
	@cd api-gateway && go build -o bin/api-gateway cmd/main.go

# Run all services locally
run:
	@echo "Starting all services..."
	@cd user-ms && go run cmd/main.go &
	@cd product-ms && go run cmd/main.go &
	@cd order-ms && go run cmd/main.go &
	@cd payment-ms && go run cmd/main.go &
	@cd api-gateway && go run cmd/main.go &

# Generate swagger documentation
swagger:
	@echo "Generating swagger documentation..."
	@cd user-ms && swag init -g cmd/main.go -o docs
	@cd product-ms && swag init -g cmd/main.go -o docs
	@cd order-ms && swag init -g cmd/main.go -o docs
	@cd payment-ms && swag init -g cmd/main.go -o docs
	@cd api-gateway && swag init -g cmd/main.go -o docs

# Start infrastructure with Docker
docker-up:
	@echo "Starting infrastructure..."
	@docker compose up -d mongodb nats

# Stop all Docker containers
docker-down:
	@echo "Stopping all containers..."
	@docker compose down

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@find . -name "bin" -type d -exec rm -rf {} +
	@find . -name "docs" -type d -exec rm -rf {} +

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/swaggo/swag/cmd/swag@latest

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...
