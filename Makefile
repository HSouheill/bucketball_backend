# Makefile for BucketBall Backend

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start all services with Docker Compose"
	@echo "  docker-down   - Stop all services"
	@echo "  docker-logs   - View logs from all services"
	@echo "  dev           - Start development environment"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/main cmd/main.go

# Run the application locally
run:
	@echo "Starting application..."
	go run cmd/main.go

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t bucketball-backend .

# Start all services with Docker Compose
docker-up:
	@echo "Starting all services..."
	docker-compose up -d

# Stop all services
docker-down:
	@echo "Stopping all services..."
	docker-compose down

# View logs from all services
docker-logs:
	@echo "Viewing logs..."
	docker-compose logs -f

# Start development environment
dev: docker-up
	@echo "Development environment started!"
	@echo "Backend: http://localhost:8080"
	@echo "MongoDB: localhost:27017"
	@echo "Redis: localhost:6379"
	@echo "View logs with: make docker-logs"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060
