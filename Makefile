.PHONY: help dev build test test-integration lint clean migrate-up migrate-down migrate-down-all migrate-create gen sqlc-check up verify reset

# Help (default target)
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Development Workflow:"
	@echo "    up             - Start all services (postgres, redis, kafka, etc.)"
	@echo "    dev            - Start dependencies and run app"
	@echo "    verify         - Run lint + unit tests (fail-fast)"
	@echo "    reset          - Clean slate: stop containers, remove volumes, clean build"
	@echo ""
	@echo "  Testing:"
	@echo "    test           - Run tests with coverage"
	@echo "    test-integration - Run integration tests (requires Docker)"
	@echo "    lint           - Run golangci-lint"
	@echo ""
	@echo "  Build:"
	@echo "    build          - Build binary to bin/"
	@echo ""
	@echo "  Database:"
	@echo "    migrate-up     - Apply all pending migrations"
	@echo "    migrate-down N=1 - Rollback N migrations (default: 1)"
	@echo "    migrate-down-all - Rollback ALL migrations"
	@echo "    migrate-create NAME=x - Create new migration"
	@echo ""
	@echo "  Code Generation:"
	@echo "    gen            - Generate sqlc, protobuf, and GraphQL code"
	@echo ""
	@echo "  Other:"
	@echo "    clean          - Stop containers and clean build"
	@echo "    worker         - Run background job worker"
	@echo "    scheduler      - Run periodic job scheduler"

# Development
dev:
	@test -f .env || cp .env.example .env 2>/dev/null || true
	docker compose up -d
	go run cmd/server/main.go

# Start all services (containers only, no app) - waits for health checks
up:
	@test -f .env || (test -f .env.example && cp .env.example .env && echo "Created .env from example") || echo "WARNING: .env not found and .env.example missing. Defaults may fail."
	@echo "Starting all services..."
	docker compose up -d --wait
	@echo "✓ All services are up and healthy"

# Verify code quality: lint + unit tests (fail-fast)
verify: lint
	@$(MAKE) test
	@echo "✓ Verification complete: lint and tests passed"

# Clean slate: stop containers, remove volumes, clean build artifacts
reset:
	docker compose down -v
	rm -rf bin/

	@echo "✓ Reset complete: containers stopped, volumes removed, artifacts cleaned"

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

# Linting with policy pack
lint:
	golangci-lint run --config policy/golangci.yml ./...

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

# Updated gen target to include all generations
gen: gen-proto gen-gql
	sqlc generate

sqlc-check:
	sqlc compile

# Cleanup
clean:
	@$(MAKE) reset

# Worker (Story 8.2)
.PHONY: worker
worker:
	go run ./cmd/worker/main.go

# Scheduler (Story 9.2) - Periodic job scheduler
.PHONY: scheduler
scheduler:
	go run ./cmd/scheduler/main.go

# CLI Tool (Story 11.1) - bplat boilerplate generator
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BPLAT_LDFLAGS := -ldflags "\
	-X 'github.com/iruldev/golang-api-hexagonal/cmd/bplat/cmd.Version=$(VERSION)' \
	-X 'github.com/iruldev/golang-api-hexagonal/cmd/bplat/cmd.BuildDate=$(BUILD_DATE)' \
	-X 'github.com/iruldev/golang-api-hexagonal/cmd/bplat/cmd.GitCommit=$(GIT_COMMIT)'"

.PHONY: build-bplat
build-bplat: ## Build bplat CLI tool
	go build $(BPLAT_LDFLAGS) -o bin/bplat ./cmd/bplat

.PHONY: install-bplat
install-bplat: ## Install bplat CLI tool to GOPATH/bin
	go install $(BPLAT_LDFLAGS) ./cmd/bplat

# gRPC (Story 12.1) - Protobuf code generation
.PHONY: install-protoc gen-proto
install-protoc: ## Install protoc-gen-go and protoc-gen-go-grpc plugins
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

gen-proto: ## Generate Go code from .proto files
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/**/**/*.proto

# GraphQL (Story 12.3)
gen-gql: ## Generate GraphQL code from schema
	go run github.com/99designs/gqlgen generate

