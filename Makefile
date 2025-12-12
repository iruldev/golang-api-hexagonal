.PHONY: help dev build test lint clean migrate-up migrate-down migrate-down-all migrate-create gen sqlc-check

# Help (default target)
help:
	@echo "Available targets:"
	@echo "  dev            - Start dependencies and run app"
	@echo "  test           - Run tests with coverage"
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

# Testing
test:
	go test -v -cover -race ./...

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
