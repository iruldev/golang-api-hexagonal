.PHONY: help dev build test lint clean migrate-up migrate-down gen

# Help (default target)
help:
	@echo "Available targets:"
	@echo "  dev         - Start dependencies and run app"
	@echo "  test        - Run tests with coverage"
	@echo "  lint        - Run golangci-lint"
	@echo "  build       - Build binary to bin/"
	@echo "  clean       - Stop containers and clean build"
	@echo "  migrate-up  - Run database migrations (Story 4.x)"
	@echo "  migrate-down - Rollback migrations (Story 4.x)"
	@echo "  gen         - Generate code with sqlc (Story 4.x)"

# Development
dev:
	@test -f .env || cp .env.example .env 2>/dev/null || true
	docker-compose up -d
	go run cmd/server/main.go

# Build
build:
	go build -o bin/server ./cmd/server

# Testing
test:
	go test -v -cover -race ./...

# Linting
lint:
	golangci-lint run ./...

# Database (placeholders for Story 4.x)
migrate-up:
	@echo "TODO: Add migration command in Story 4.x"

migrate-down:
	@echo "TODO: Add migration command in Story 4.x"

# Code generation (placeholder for sqlc)
gen:
	@echo "TODO: Add sqlc generate in Story 4.x"

# Cleanup
clean:
	docker-compose down -v
	rm -rf bin/
