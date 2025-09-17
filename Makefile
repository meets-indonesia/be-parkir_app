# Makefile for Parking Digital API

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs migrate-up migrate-down

# Default target
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application locally"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-up      - Start services with Docker Compose"
	@echo "  docker-down    - Stop services with Docker Compose"
	@echo "  docker-logs    - View Docker logs"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/main ./cmd/server

# Run the application locally
run:
	@echo "Running application..."
	go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t be-parkir .

# Start services with Docker Compose (development)
docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Start services with Docker Compose (production)
docker-up-prod:
	@echo "Starting services with Docker Compose (production)..."
	docker-compose -f docker-compose.prod.yml up -d

# Stop services with Docker Compose
docker-down:
	@echo "Stopping services with Docker Compose..."
	docker-compose down

# Stop services with Docker Compose (production)
docker-down-prod:
	@echo "Stopping services with Docker Compose (production)..."
	docker-compose -f docker-compose.prod.yml down

# View Docker logs
docker-logs:
	@echo "Viewing Docker logs..."
	docker-compose logs -f

# View Docker logs (production)
docker-logs-prod:
	@echo "Viewing Docker logs (production)..."
	docker-compose -f docker-compose.prod.yml logs -f

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	# Add migration commands here when using a migration tool

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	# Add rollback commands here when using a migration tool

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Full development setup
dev-setup: deps build docker-up
	@echo "Development setup complete!"

# Full production setup
prod-setup: deps build docker-up-prod
	@echo "Production setup complete!"
