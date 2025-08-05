.PHONY: swagger-gen swagger-install build run test clean help

# Variables
BINARY_NAME=caviar-api
MAIN_PATH=cmd/api/main.go
DOCS_PATH=docs

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Swagger targets
swagger-install: ## Install swag CLI tool
	@echo "Installing swag CLI tool..."
	go install github.com/swaggo/swag/cmd/swag@latest

swagger-gen: ## Generate swagger documentation
	@echo "Generating swagger documentation..."
	swag init -g $(MAIN_PATH) -o $(DOCS_PATH) --parseDependency --parseInternal

swagger-fmt: ## Format swagger comments
	@echo "Formatting swagger comments..."
	swag fmt -g $(MAIN_PATH)

# Build targets
build: swagger-gen ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run: swagger-gen ## Run the application
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

# Development targets
dev: swagger-gen ## Run in development mode with hot reload
	@echo "Running in development mode..."
	air -c .air.toml

# Test targets
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Database targets
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations $(NAME)

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf $(DOCS_PATH)/
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME)

# Complete setup
setup: deps swagger-install ## Complete project setup
	@echo "Project setup complete!"

# Quick development workflow
quick: swagger-gen build ## Quick build with swagger generation
	@echo "Quick build complete!"