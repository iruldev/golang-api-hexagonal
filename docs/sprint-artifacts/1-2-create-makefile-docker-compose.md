# Story 1.2: Create Makefile & Docker Compose

Status: Ready for Review

## Story

As a developer,
I want to start all dependencies with a single command,
So that I don't need to manually configure external services.

## Acceptance Criteria

### AC1: make dev starts dependencies ✅
**Given** docker and docker-compose are installed
**When** I run `make dev`
**Then** PostgreSQL container starts on port 5432
**And** Jaeger container starts on port 16686 (optional)
**And** the Go application compiles and runs

### AC2: make test runs tests ✅
**Given** I want to run tests
**When** I run `make test`
**Then** all tests execute with coverage report

### AC3: make lint checks code quality ✅
**Given** I want to check code quality
**When** I run `make lint`
**Then** golangci-lint runs with project configuration

---

## Tasks / Subtasks

- [x] **Task 1: Create docker-compose.yaml** (AC: #1)
  - [x] Define PostgreSQL service (port 5432, volume for data)
  - [x] Define Jaeger service (ports 16686, 6831, 4317)
  - [x] Configure network for service communication
  - [x] Add healthcheck for PostgreSQL

- [x] **Task 2: Create Makefile** (AC: #1, #2, #3)
  - [x] Add `help` target: show all available commands
  - [x] Add `dev` target: docker-compose up + go run
  - [x] Add `test` target: go test ./... -cover
  - [x] Add `lint` target: golangci-lint run
  - [x] Add `build` target: go build ./...
  - [x] Add `migrate-up` target (placeholder for Story 4.x)
  - [x] Add `migrate-down` target (placeholder for Story 4.x)
  - [x] Add `gen` target (placeholder for sqlc)

- [x] **Task 3: Create .golangci.yml** (AC: #3)
  - [x] Configure linters (errcheck, govet, staticcheck, etc.)
  - [x] Set cyclomatic complexity limit ≤ 15
  - [x] Exclude generated files
  - [x] Match project_context.md standards

- [x] **Task 4: Verify all make targets work** (AC: #1, #2, #3)
  - [x] Test `make help` shows all targets
  - [x] Test `make test` runs with coverage
  - [x] Test `make build` creates bin/server

---

## Dev Notes

### docker-compose.yaml Structure

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: golang-api-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: app
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  jaeger:
    image: jaegertracing/all-in-one:1.53
    container_name: golang-api-jaeger
    ports:
      - "16686:16686"  # UI
      - "6831:6831/udp"  # Thrift compact
      - "4317:4317"  # OTLP gRPC (for OpenTelemetry)
    environment:
      COLLECTOR_OTLP_ENABLED: true

volumes:
  postgres_data:
```

**Note:** `version` key is deprecated in modern docker-compose - do NOT include it.

### Makefile Structure

```makefile
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
```

### .golangci.yml Structure

```yaml
run:
  timeout: 5m
  
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gocyclo
    - gofmt
    - goimports

linters-settings:
  gocyclo:
    min-complexity: 15

issues:
  exclude-dirs:
    - vendor
```

### Technology Stack (from Story 1.1)

| Tech | Package | Purpose |
|------|---------|---------|
| Go | 1.24.x | Runtime |
| Router | go-chi/chi/v5 | HTTP routing |
| Database | jackc/pgx/v5 | PostgreSQL driver |
| Logger | go.uber.org/zap | Structured logging |
| Tracing | go.opentelemetry.io/otel | Observability |

### Previous Story Learnings

From Story 1.1:
- Project structure with internal/ hierarchy is complete
- go.mod has all dependencies (chi, pgx, zap, koanf, otel, testify)
- Basic main.go exists at cmd/server/main.go
- Tests pass with `go test ./...`

### Critical Don'ts

❌ Don't add actual app logic to main.go (that's Story 2.x+)
❌ Don't implement migrations yet (Story 4.x)
❌ Don't add config loading yet (Story 2.x)
❌ Don't hardcode database credentials in code
❌ Don't include `version:` key in docker-compose (deprecated)

---

## References

- [Source: docs/architecture.md#Infrastructure-DX]
- [Source: docs/architecture.md#Docker-Compose]
- [Source: docs/project_context.md#Code-Quality]
- [Source: docs/prd.md#FR3] - Developer can start all dependencies with single command (`make dev`)
- [Source: docs/epics.md#Story-1.2]

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Enhanced by validate-create-story quality review.

### Agent Model Used

Dev implementation, Code review with fix.

### Debug Log References

None.

### Completion Notes List

- Story created: 2025-12-10
- Quality validation applied: 2025-12-10
- Code review fixes applied: 2025-12-11
  - Fixed .golangci.yml: added version: "2"
  - Added restart: unless-stopped to containers
  - Added explicit app-network for service communication
  - Updated File List to past tense

### File List

Files created:
- `docker-compose.yaml`
- `Makefile`
- `.golangci.yml`
- `.env.example`
