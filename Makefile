.PHONY: help dev build test test-integration lint clean migrate-up migrate-down migrate-down-all migrate-create gen sqlc-check

# Help (default target)
help:
	@echo "Available targets:"
	@echo "  dev            - Start dependencies and run app"
	@echo "  test           - Run tests with coverage"
	@echo "  test-integration - Run integration tests (requires Docker)"
	@echo "  lint           - Run golangci-lint"
	@echo "  build          - Build binary to bin/"
	@echo "  clean          - Stop containers and clean build"
	@echo "  migrate-up     - Apply all pending migrations"
	@echo "  migrate-down N=1 - Rollback N migrations (default: 1)"
	@echo "  migrate-down-all - Rollback ALL migrations"
	@echo "  migrate-create NAME=x - Create new migration"
	@echo "  gen            - Generate sqlc code"

# Development
dev:
	@test -f .env || cp .env.example .env 2>/dev/null || true
	docker compose up -d
	go run cmd/server/main.go

# Build
build:
	go build -o bin/server ./cmd/server

# Testing (use env -u to unset all vars from .env that could interfere with tests)
test:
	env -u APP_HTTP_PORT -u APP_NAME -u APP_ENV \
	    -u DB_HOST -u DB_PORT -u DB_USER -u DB_NAME -u DB_PASSWORD -u DB_SSL_MODE \
	    -u DB_MAX_OPEN_CONNS -u DB_MAX_IDLE_CONNS -u DB_CONN_MAX_LIFETIME -u DB_CONN_TIMEOUT -u DB_QUERY_TIMEOUT \
	    -u REDIS_HOST -u REDIS_PORT -u REDIS_PASSWORD -u REDIS_DB -u REDIS_POOL_SIZE -u REDIS_MIN_IDLE_CONNS \
	    -u REDIS_DIAL_TIMEOUT -u REDIS_READ_TIMEOUT -u REDIS_WRITE_TIMEOUT \
	    -u OTEL_EXPORTER_OTLP_ENDPOINT -u OTEL_SERVICE_NAME \
	    -u LOG_LEVEL -u LOG_FORMAT \
	    -u ASYNQ_CONCURRENCY -u ASYNQ_RETRY_MAX -u ASYNQ_SHUTDOWN_TIMEOUT \
	    -u APP_CONFIG_FILE \
	go test -v -cover -race -p 1 ./...

# Integration tests (requires Docker)
test-integration:
	go test -v -tags=integration ./...

# Linting
lint:
	golangci-lint run ./...

# Database migrations (Story 4.4)
# Load env vars for database URL
-include .env
export

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

migrate-up: ## Apply all pending database migrations
	migrate -database "$(DATABASE_URL)" -path db/migrations up

migrate-down: ## Rollback N migrations (usage: make migrate-down N=2)
	migrate -database "$(DATABASE_URL)" -path db/migrations down $(or $(N),1)

migrate-down-all: ## Rollback ALL migrations (dangerous!)
	@echo "WARNING: This will rollback ALL migrations!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	migrate -database "$(DATABASE_URL)" -path db/migrations down -all

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_posts)
	@test -n "$(NAME)" || (echo "Usage: make migrate-create NAME=migration_name" && exit 1)
	migrate create -ext sql -dir db/migrations -seq $(NAME)

# Code generation (Story 4.3: sqlc)
gen:
	sqlc generate

sqlc-check:
	sqlc compile

# Cleanup
clean:
	docker compose down -v
	rm -rf bin/

# Worker (Story 8.2)
.PHONY: worker
worker:
	go run ./cmd/worker/main.go

# Scheduler (Story 9.2) - Periodic job scheduler
.PHONY: scheduler
scheduler:
	go run ./cmd/scheduler/main.go

