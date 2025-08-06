# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Caviar is a Go-based REST API for managing premium caviar products and orders. The application uses:
- **Gin** for HTTP routing and middleware
- **GORM** with PostgreSQL for data persistence  
- **MinIO** for S3-compatible object storage
- **Swagger** for API documentation
- **Zap** for structured logging
- **Docker Compose** for local development environment

## Development Commands

### Build & Run
```bash
make build          # Build the application with swagger docs
make run            # Run the application with swagger generation
make dev            # Run in development mode with hot reload (requires air)
```

### Testing & Quality
```bash
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make lint           # Run golangci-lint
```

### Documentation
```bash
make swagger-gen    # Generate swagger documentation
make swagger-fmt    # Format swagger comments
make swagger-install # Install swag CLI tool
```

### Database Operations
```bash
make migrate-up     # Run database migrations up
make migrate-down   # Run database migrations down
make migrate-create NAME=migration_name # Create new migration
```

### Docker & Setup
```bash
make setup          # Complete project setup (deps + swagger)
make docker-build   # Build Docker image
make docker-run     # Run Docker container
docker-compose up   # Start all services (postgres, dragonfly, minio, api)
```

## Architecture

### Project Structure
```
cmd/api/            # Application entrypoint with swagger annotations
internal/
  app/              # Application bootstrap and dependency injection
  config/           # Configuration structs with environment binding
  controller/rest/  # HTTP handlers, middleware, and routing
  dto/              # Data transfer objects and converters
  models/           # Domain models with GORM annotations
  service/          # Business logic layer
  storage/          # Data access layer
  types/            # Shared type definitions
pkg/
  apperror/         # Structured error handling
  db/               # Database client implementations (PostgreSQL, MinIO)
migrations/         # SQL migration files
```

### Key Patterns
- **Clean Architecture**: Separated layers (controller → service → storage)
- **Dependency Injection**: All dependencies wired in `internal/app/app.go`
- **Error Handling**: Custom `AppError` type with HTTP status codes
- **Configuration**: Environment-based config with struct tags
- **Database Models**: GORM models with UUID primary keys and JSONB fields

### Service Dependencies
The application requires these external services:
- **PostgreSQL**: Primary database (port configurable via POSTGRES_PORT)
- **Dragonfly**: Redis-compatible cache (port configurable via DRAGONFLY_PORT) 
- **MinIO**: S3-compatible object storage (ports configurable via MINIO_PORT/MINIO_CONSOLE_PORT)

### Configuration
All configuration is environment-based using struct tags. Key environment variables:
- Database: `POSTGRES_*` prefix
- Object Storage: `MINIO_*` prefix  
- Authentication: `AUTH_SECRET`
- Server: `SERVER_PORT` (default: 8080)
- Production mode: `IS_PROD` (affects swagger availability)

### API Structure
- Base path: `/api/v1`
- Health check: `/api/v1/health`
- Swagger docs: `/swagger/*` (development only)
- Product management endpoints with authentication middleware

### Product Domain
Complex product model with:
- Multiple variants per product (different masses/prices)
- JSONB storage for caviar-specific details (fish age, grain size, etc.)
- Multi-currency pricing support
- Temperature-controlled shelf life specifications