# golang-api-hexagonal Makefile
# Run `make help` to see available targets

.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=api
COVERAGE_THRESHOLD ?= 80

# Docker parameters
DOCKER_COMPOSE=docker compose
DOCKER_VOLUME_PGDATA ?= golang-api-hexagonal-pgdata
INFRA_TIMEOUT ?= 60
INFRA_CONFIRM ?=

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
	@echo ""
	@echo "  Installing golangci-lint..."
	@if ! golangci-lint --version 2>/dev/null | grep -q "1.64.2"; then \
		echo "    ⚠️  golangci-lint missing or version mismatch, installing v1.64.2..."; \
		go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@v1.64.2; \
	else \
		echo "    ✅ golangci-lint v1.64.2 already installed"; \
	fi
	@echo ""
	@echo "  Installing goose..."
	@if ! goose --version 2>/dev/null | grep -q "v3.26.0"; then \
		echo "    ⚠️  goose missing or version mismatch, installing v3.26.0..."; \
		go install github.com/pressly/goose/v3/cmd/goose@v3.26.0; \
	else \
		echo "    ✅ goose v3.26.0 already installed"; \
	fi
	@echo ""
	@echo "  Installing sqlc..."
	@if ! sqlc version 2>/dev/null | grep -q "v1.28.0"; then \
		echo "    ⚠️  sqlc missing or version mismatch, installing v1.28.0..."; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.28.0; \
	else \
		echo "    ✅ sqlc v1.28.0 already installed"; \
	fi
	@echo ""
	@echo "  Creating .env.local..."
	@if [ ! -f .env.local ]; then \
		cp .env.example .env.local; \
		echo "    ✅ Created .env.local from .env.example"; \
	else \
		echo "    ✅ .env.local already exists"; \
	fi
	@echo ""
	@echo "📦 Downloading Go modules..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo ""
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Start infrastructure:  make infra-up"
	@echo "  2. Run migrations:        export DATABASE_URL=\"postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable\""
	@echo "                            make migrate-up"
	@echo "  3. Run the service:       make run"
	@echo ""
	@echo "Run 'make help' to see all available targets."

## generate: Run sqlc to generate type-safe SQL code (Story 5.3)
.PHONY: generate
generate:
	@echo "🔧 Generating sqlc code..."
	@which sqlc > /dev/null || (echo "❌ sqlc not found. Run 'make setup' first." && exit 1)
	sqlc generate
	@echo "✅ Code generation complete"


## build: Build the application
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/api

## run: Run the application
.PHONY: run
run:
	$(GOCMD) run ./cmd/api

## test: Run all tests (usage: make test ARGS="-run TestName")
.PHONY: test
test:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./... $(ARGS)

## test-integration: Run integration tests (requires DATABASE_URL with *_test database)
## Story 6.3: Integration tests require a test database to be running
.PHONY: test-integration
test-integration:
	@echo "🧪 Running integration tests..."
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "❌ DATABASE_URL not set. Set it to a test database (e.g., *_test):"; \
		echo "  export DATABASE_URL=\"postgres://postgres:postgres@localhost:5432/golang_api_hexagonal_test?sslmode=disable\""; \
		exit 1; \
	fi
	@if echo "$$DATABASE_URL" | grep -qv "_test"; then \
		echo "❌ DATABASE_URL must end in '_test' to prevent accidental data loss."; \
		echo "   Current: $$DATABASE_URL"; \
		exit 1; \
	fi
	$(GOTEST) -v -race -tags=integration ./... $(ARGS)
	@echo "✅ Integration tests complete"

## coverage: Check test coverage meets 80% threshold for domain+app
.PHONY: coverage
coverage:
	@echo "📊 Running tests with coverage (domain+app)..."
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic \
		./internal/domain/... \
		./internal/app/...
	@echo ""
	@echo "📈 Coverage report:"
	@go tool cover -func=coverage.out | tail -1
	@THRESHOLD="$(COVERAGE_THRESHOLD)"; \
	COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{gsub(/%/,"",$$3); print $$3}'); \
	if awk 'BEGIN {exit !('"$$COVERAGE"' < '"$$THRESHOLD"')}'; then \
		echo ""; \
		echo "❌ Coverage $$COVERAGE% is below $$THRESHOLD% threshold"; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ Coverage $$COVERAGE% meets $$THRESHOLD% threshold"; \
	fi

## lint: Run linter
.PHONY: lint
lint:
	golangci-lint run ./...

# =============================================================================
# CI Pipeline
# =============================================================================

## ci: Run full CI pipeline locally (mod-tidy, fmt, lint, test)
.NOTPARALLEL: ci
.PHONY: ci
ci:
	@echo ""
	@echo "🚀 Running Local CI Pipeline"
	@echo "=============================="
	@echo ""
	@$(MAKE) check-mod-tidy
	@$(MAKE) check-fmt
	@$(MAKE) lint
	@$(MAKE) test
	@echo ""
	@echo "=============================="
	@echo "✅ All CI checks passed!"
	@echo ""

## check-mod-tidy: Verify go.mod and go.sum are tidy
.PHONY: check-mod-tidy
check-mod-tidy:
	@echo "📦 Checking go.mod is tidy..."
	@if [ -z "$$ALLOW_DIRTY" ] && ! git diff --exit-code > /dev/null 2>&1; then \
		echo ""; \
		echo "❌ Working tree is not clean (required for mod tidy check)"; \
		echo "   Commit or stash changes, then rerun"; \
		echo "   (or rerun with ALLOW_DIRTY=1 if you intentionally want to skip this clean-tree guard)"; \
		echo ""; \
		git --no-pager diff --name-only; \
		exit 1; \
	fi
	@$(GOMOD) tidy
	@if ! git diff --exit-code go.mod go.sum > /dev/null 2>&1; then \
		echo ""; \
		echo "❌ go.mod or go.sum is not tidy"; \
		echo "   Run 'go mod tidy' and commit the changes"; \
		git --no-pager diff --stat go.mod go.sum; \
		exit 1; \
	fi
	@echo "✅ go.mod is tidy"

## check-fmt: Verify code is formatted with gofmt
.PHONY: check-fmt
check-fmt:
	@echo "📐 Checking code formatting (gofmt)..."
	@if [ -z "$$ALLOW_DIRTY" ] && ! git diff --exit-code > /dev/null 2>&1; then \
		echo ""; \
		echo "❌ Working tree is not clean (required for gofmt check)"; \
		echo "   Commit or stash changes, then rerun"; \
		echo "   (or rerun with ALLOW_DIRTY=1 if you intentionally want to skip this clean-tree guard)"; \
		echo ""; \
		git --no-pager diff --name-only; \
		exit 1; \
	fi
	@if [ -n "$$ALLOW_DIRTY" ]; then \
		echo "Skipping gofmt write due to ALLOW_DIRTY=1 (manual gofmt recommended before commit)"; \
	else \
		FILES=$$(git ls-files '*.go'); \
		if [ -n "$$FILES" ]; then \
			gofmt -w $$FILES; \
		fi; \
	fi; \
	if [ -z "$$ALLOW_DIRTY" ] && ! git diff --exit-code > /dev/null 2>&1; then \
		echo ""; \
		echo "❌ gofmt would change files"; \
		echo "   Run 'gofmt -w .' and commit the changes"; \
		echo ""; \
		echo "Changed files:"; \
		git --no-pager diff --name-only; \
		exit 1; \
	fi
	@echo "✅ All files are formatted"

## clean: Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME) coverage.out
	$(GOCMD) clean

# =============================================================================
# Infrastructure
# =============================================================================

## infra-up: Start infrastructure (PostgreSQL) (INFRA_TIMEOUT=60)
.PHONY: infra-up
infra-up:
	@echo "🐘 Starting PostgreSQL..."
	$(DOCKER_COMPOSE) up -d postgres
	@echo "⏳ Waiting for PostgreSQL to be healthy (timeout: $(INFRA_TIMEOUT)s)..."
	@set -e; \
	timeout=$(INFRA_TIMEOUT); \
	cid=$$($(DOCKER_COMPOSE) ps -q postgres); \
	if [ -z "$$cid" ]; then \
		echo "❌ PostgreSQL container not found after startup"; \
		exit 1; \
	fi; \
	while true; do \
		status=$$(docker inspect --format='{{.State.Health.Status}}' "$$cid" 2>/dev/null || echo "unknown"); \
		if [ "$$status" = "healthy" ]; then \
			break; \
		fi; \
		if [ "$$status" = "unhealthy" ]; then \
			echo "❌ PostgreSQL reported unhealthy"; \
			echo "   Run 'make infra-logs' to check container logs"; \
			exit 1; \
		fi; \
		timeout=$$((timeout - 2)); \
		if [ $$timeout -le 0 ]; then \
			echo "❌ Timeout: PostgreSQL did not become healthy in $(INFRA_TIMEOUT)s"; \
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
	@set -e; \
	cid=$$($(DOCKER_COMPOSE) ps -q postgres); \
	if [ -n "$$cid" ]; then \
		$(DOCKER_COMPOSE) stop postgres; \
		$(DOCKER_COMPOSE) rm -f postgres; \
	fi
	@echo "✅ Infrastructure stopped (data preserved)"

## infra-reset: Stop infrastructure and remove volumes (DESTRUCTIVE) (INFRA_CONFIRM=y)
.PHONY: infra-reset
infra-reset:
	@echo "WARNING: removing volumes"
	@echo "⚠️  This will delete all database data!"
	@if [ "$(INFRA_CONFIRM)" != "y" ]; then \
		printf "Are you sure? [y/N] " && read confirm && [ "$$confirm" = "y" ] || exit 1; \
	fi
	@set -e; \
	cid=$$($(DOCKER_COMPOSE) ps -q postgres); \
	if [ -n "$$cid" ]; then \
		$(DOCKER_COMPOSE) stop postgres; \
		$(DOCKER_COMPOSE) rm -f postgres; \
	fi; \
	if docker volume inspect "$(DOCKER_VOLUME_PGDATA)" > /dev/null 2>&1; then \
		docker volume rm -f "$(DOCKER_VOLUME_PGDATA)"; \
	fi
	@echo "✅ Infrastructure reset complete"

## infra-logs: View infrastructure logs
.PHONY: infra-logs
infra-logs:
	$(DOCKER_COMPOSE) logs -f postgres

## infra-status: Show infrastructure status
.PHONY: infra-status
infra-status:
	$(DOCKER_COMPOSE) ps postgres

# =============================================================================
# Database Migrations
# =============================================================================

# Helper to check prerequisites
.PHONY: _check-goose _check-db-url

_check-goose:
	@which goose > /dev/null || (echo "❌ goose not found. Run 'make setup' first." && exit 1)

_check-db-url:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL is not set."; \
		echo ""; \
		echo "Set it with:"; \
		echo "  export DATABASE_URL=\"postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable\""; \
		echo ""; \
		echo "Or source from .env.example:"; \
		echo "  export \$$(grep DATABASE_URL .env.example | xargs)"; \
		exit 1; \
	fi

## migrate-up: Run all pending migrations
.PHONY: migrate-up
migrate-up: _check-goose _check-db-url
	@echo "🔄 Running migrations..."
	goose -dir migrations postgres "$(DATABASE_URL)" up
	@echo "✅ Migrations complete"

## migrate-down: Rollback the last migration
.PHONY: migrate-down
migrate-down: _check-goose _check-db-url
	@echo "⏪ Rolling back last migration..."
	goose -dir migrations postgres "$(DATABASE_URL)" down
	@echo "✅ Rollback complete"

## migrate-status: Show migration status
.PHONY: migrate-status
migrate-status: _check-goose _check-db-url
	goose -dir migrations postgres "$(DATABASE_URL)" status

## migrate-create: Create a new migration (usage: make migrate-create name=description)
.PHONY: migrate-create
migrate-create: _check-goose
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=description"; exit 1; fi
	goose -dir migrations create "$(name)" sql

## migrate-validate: Validate migration files syntax (no DB required)
.PHONY: migrate-validate
migrate-validate: _check-goose
	@echo "🔍 Validating migration files..."
	@echo "  Running goose validate..."
	@goose -dir migrations validate
	@echo ""
	@echo "  Checking goose annotations..."
	@for f in migrations/*.sql; do \
		if [ -f "$$f" ]; then \
			echo "    Checking $$f..."; \
			if ! grep -q -e "+goose Up" "$$f"; then \
				echo "      ❌ Missing '-- +goose Up' section"; \
				exit 1; \
			fi; \
			if ! grep -q -e "+goose Down" "$$f"; then \
				echo "      ❌ Missing '-- +goose Down' section"; \
				exit 1; \
			fi; \
			echo "      ✅ Annotations valid"; \
		fi; \
	done
	@echo ""
	@echo "✅ All migration files are valid"

# =============================================================================
# OpenAPI
# =============================================================================

## openapi: Validate OpenAPI spec (requires spectral or npx)
.PHONY: openapi
openapi:
	@echo "🔍 Validating OpenAPI spec..."
	@if command -v docker > /dev/null 2>&1; then \
		echo "🐳 Running Spectral via Docker..."; \
		docker run --rm -v $(CURDIR):/tmp stoplight/spectral:6.15.0 lint /tmp/docs/openapi.yaml --ruleset /tmp/.spectral.yaml; \
	elif command -v npx > /dev/null 2>&1; then \
		npx --yes @stoplight/spectral-cli lint docs/openapi.yaml; \
	elif command -v spectral > /dev/null 2>&1; then \
		spectral lint docs/openapi.yaml; \
	else \
		echo "⚠️  No validator found (docker, spectral, or npx)"; \
		echo "   Checking YAML syntax only..."; \
		if command -v python3 > /dev/null 2>&1; then \
			python3 -c "import yaml; yaml.safe_load(open('docs/openapi.yaml'))"; \
			echo "✅ YAML syntax is valid"; \
		else \
			echo "   Install Docker, Node.js (npx), or Spectral for full validation"; \
		fi; \
	fi

## openapi-view: View OpenAPI spec in browser (requires redoc-cli or npx)
.PHONY: openapi-view
openapi-view:
	@echo "🌐 Opening OpenAPI spec in browser..."
	@if command -v npx > /dev/null 2>&1; then \
		npx --yes @redocly/cli preview-docs docs/openapi.yaml; \
	else \
		echo "❌ npx not found. Install Node.js first."; \
		exit 1; \
	fi

