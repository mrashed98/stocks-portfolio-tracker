# Portfolio App Backend Makefile

.PHONY: help build test test-db migrate-up migrate-down seed clean

# Default target
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  test       - Run unit tests"
	@echo "  test-db    - Run database integration tests"
	@echo "  migrate-up - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  seed       - Seed development data"
	@echo "  clean      - Clean build artifacts"

# Build the application
build:
	go build -o bin/portfolio-app ./main.go

# Run unit tests
test:
	go test -v ./...

# Run database integration tests
test-db:
	@echo "Running database tests..."
	@if [ -z "$(TEST_DB_URL)" ]; then \
		echo "TEST_DB_URL not set. Please set it to run database tests."; \
		echo "Example: export TEST_DB_URL=postgres://test_user:test_pass@localhost:5432/portfolio_test?sslmode=disable"; \
		exit 1; \
	fi
	go test -v ./internal/database/...

# Install migrate tool if not present
install-migrate:
	@which migrate > /dev/null || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)

# Run database migrations
migrate-up: install-migrate
	migrate -path ./migrations -database "$(DATABASE_URL)" up

# Rollback database migrations
migrate-down: install-migrate
	migrate -path ./migrations -database "$(DATABASE_URL)" down

# Create a new migration file
migrate-create: install-migrate
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

# Seed development data
seed:
	go run ./cmd/seed/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run the application in development mode
dev:
	go run main.go

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	@echo "1. Make sure PostgreSQL and Redis are running"
	@echo "2. Create database: createdb portfolio_db"
	@echo "3. Create test database: createdb portfolio_test"
	@echo "4. Copy .env.example to .env and update values"
	@echo "5. Run 'make migrate-up' to apply migrations"
	@echo "6. Run 'make dev' to start the application"