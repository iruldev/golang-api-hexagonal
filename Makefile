# golang-api-hexagonal Makefile
# Run `make help` to see available targets

.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=api

# Docker parameters
DOCKER_COMPOSE=docker compose

## help: Show this help message
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

# =============================================================================
# Development
# =============================================================================

## setup: Install development tools and dependencies
.PHONY: setup
setup:
	@echo "📦 Installing development tools..."
	@which golangci-lint > /dev/null || go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
	@which goose > /dev/null || go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "📦 Downloading Go modules..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Setup complete!"

## build: Build the application
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/api

## run: Run the application
.PHONY: run
run:
	$(GOCMD) run ./cmd/api

## test: Run all tests
.PHONY: test
test:
	$(GOTEST) -v -race ./...

## lint: Run linter
.PHONY: lint
lint:
	golangci-lint run ./...

## clean: Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	$(GOCMD) clean

# =============================================================================
# Infrastructure
# =============================================================================

## infra-up: Start infrastructure (PostgreSQL)
.PHONY: infra-up
infra-up:
	@echo "🐘 Starting PostgreSQL..."
	$(DOCKER_COMPOSE) up -d
	@echo "⏳ Waiting for PostgreSQL to be healthy (timeout: 60s)..."
	@timeout=60; \
	while ! $(DOCKER_COMPOSE) ps | grep -q "healthy"; do \
		timeout=$$((timeout - 2)); \
		if [ $$timeout -le 0 ]; then \
			echo "❌ Timeout: PostgreSQL did not become healthy in 60s"; \
			echo "   Run 'make infra-logs' to check container logs"; \
			exit 1; \
		fi; \
		echo "  Waiting... ($$timeout s remaining)"; \
		sleep 2; \
	done
	@echo "✅ Infrastructure is ready!"
	@echo ""
	@echo "PostgreSQL connection:"
	@echo "  Host: localhost:5432"
	@echo "  User: \$${POSTGRES_USER:-postgres}"
	@echo "  Pass: \$${POSTGRES_PASSWORD:-postgres}"
	@echo "  DB:   \$${POSTGRES_DB:-golang_api_hexagonal}"

## infra-down: Stop infrastructure (preserve data)
.PHONY: infra-down
infra-down:
	@echo "🛑 Stopping infrastructure..."
	$(DOCKER_COMPOSE) down
	@echo "✅ Infrastructure stopped (data preserved)"

## infra-reset: Stop infrastructure and remove volumes (DESTRUCTIVE)
.PHONY: infra-reset
infra-reset:
	@echo "⚠️  WARNING: This will delete all database data!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	$(DOCKER_COMPOSE) down -v
	@echo "✅ Infrastructure reset complete"

## infra-logs: View infrastructure logs
.PHONY: infra-logs
infra-logs:
	$(DOCKER_COMPOSE) logs -f

## infra-status: Show infrastructure status
.PHONY: infra-status
infra-status:
	$(DOCKER_COMPOSE) ps

# =============================================================================
# Database Migrations
# =============================================================================

## migrate-up: Run all pending migrations
.PHONY: migrate-up
migrate-up:
	@echo "🔄 Running migrations..."
	goose -dir migrations postgres "$(DATABASE_URL)" up
	@echo "✅ Migrations complete"

## migrate-down: Rollback the last migration
.PHONY: migrate-down
migrate-down:
	@echo "⏪ Rolling back last migration..."
	goose -dir migrations postgres "$(DATABASE_URL)" down
	@echo "✅ Rollback complete"

## migrate-status: Show migration status
.PHONY: migrate-status
migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

## migrate-create: Create a new migration (usage: make migrate-create name=description)
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=description"; exit 1; fi
	goose -dir migrations create $(name) sql
