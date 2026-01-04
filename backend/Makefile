.PHONY: build run test lint migrate-up migrate-down docker-up docker-down clean

# Variables
APP_NAME=medical-api
MAIN_PATH=./cmd/api
MIGRATIONS_PATH=./migrations
DATABASE_URL?=postgres://postgres:postgres@localhost:5432/medical_db?sslmode=disable

# Build
build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

# Run locally
run:
	go run $(MAIN_PATH)

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Tests
test:
	go test -v -race -cover ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting (requires golangci-lint)
lint:
	golangci-lint run ./...

# Database migrations
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $$version

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f api

docker-build:
	docker-compose build

# Development tools
docker-dev:
	docker-compose --profile dev-tools up -d

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install tools
tools:
	go install github.com/cosmtrek/air@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate mocks (requires mockery)
mocks:
	mockery --all --dir=internal/repository --output=internal/repository/mocks
	mockery --all --dir=internal/service --output=internal/service/mocks

# Help
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run with hot reload"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linter"
	@echo "  make migrate-up     - Run all migrations"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-create - Create new migration"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make tools          - Install development tools"
