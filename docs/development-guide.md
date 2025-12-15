# Development Guide

Panduan pengembangan untuk **Backend Service Golang Boilerplate**.

---

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| **Go** | 1.24.x | Runtime |
| **Docker** | 20.x+ | Container runtime |
| **Docker Compose** | 2.x+ | Local infrastructure |
| **golangci-lint** | Latest | Code linting |
| **sqlc** | Latest | SQL code generation |
| **golang-migrate** | Latest | Database migrations |
| **Make** | - | Build automation |

---

## Quick Start

```bash
# Clone repository
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal

# Start dependencies (PostgreSQL, Redis, Prometheus, Grafana)
docker-compose up -d

# Download Go dependencies
go mod download

# Run database migrations
make migrate-up

# Run application in development mode
make dev
```

---

## Environment Setup

### 1. Copy Environment File

```bash
cp .env.example .env
```

### 2. Key Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| **Application** |||
| `APP_ENV` | `development` | Environment (development/staging/production) |
| `APP_PORT` | `8080` | HTTP server port |
| **Database** |||
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | - | Database password |
| `DB_NAME` | `app` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| **Redis** |||
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | - | Redis password (optional) |
| `REDIS_DB` | `0` | Redis database number |
| **Authentication** |||
| `JWT_SECRET` | - | JWT signing key (≥32 bytes) |
| `JWT_ISSUER` | - | JWT issuer (optional) |
| `JWT_AUDIENCE` | - | JWT audience (optional) |
| `API_KEYS` | - | API key pairs (`key:service,key:service`) |
| **Observability** |||
| `OTEL_EXPORTER_OTLP_ENDPOINT` | - | OpenTelemetry collector endpoint |
| `PROMETHEUS_ENABLED` | `true` | Enable Prometheus metrics |
| **Feature Flags** |||
| `FF_*` | `false` | Feature flags (e.g., `FF_NEW_FEATURE=true`) |

---

## Local Development

### Docker Compose Services

```bash
# Start all services
docker-compose up -d

# View running services
docker-compose ps

# View logs
docker-compose logs -f app
```

| Service | Port | Purpose |
|---------|------|---------|
| **PostgreSQL** | 5432 | Primary database |
| **Redis** | 6379 | Cache, job queue |
| **Prometheus** | 9090 | Metrics collection |
| **Grafana** | 3000 | Metrics visualization |
| **Jaeger** | 16686 | Distributed tracing |

### Running the Server

```bash
# Development mode (hot reload with air if installed)
make dev

# Or directly with go
go run ./cmd/server

# Run background worker
go run ./cmd/worker

# Run scheduler
go run ./cmd/scheduler
```

### GraphQL Playground

Available in development mode at: `http://localhost:8080/playground`

> ⚠️ Disabled in staging/production environments.

---

## Database Operations

### Migrations

```bash
# Apply all migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create name=add_users_table
```

### SQLC Code Generation

```bash
# Generate Go code from SQL queries
make sqlc

# Location of query files: db/queries/
# Generated code: internal/infra/postgres/db.sqlc.go
```

---

## Build Commands

```bash
# Build all binaries
make build

# Build specific binary
make build-server
make build-worker
make build-bplat

# Install bplat CLI globally
make install-bplat
```

---

## Testing

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-cover

# Run specific package
go test ./internal/domain/...

# Run integration tests (requires Docker)
go test -tags=integration ./...
```

### Test Patterns

| Type | Location | Pattern |
|------|----------|---------|
| Unit tests | `*_test.go` next to source | `go test ./...` |
| Integration tests | `*_integration_test.go` | Uses testcontainers |
| Handler tests | `handler_test.go` | HTTP test recorder |

---

## Code Generation

### Generate New Module with bplat

```bash
# Build bplat CLI
make build-bplat

# Generate new domain module
./bin/bplat generate module payment

# With custom entity name
./bin/bplat generate module orders --entity Order
```

### Generated Structure

```
internal/
├── domain/payment/
│   ├── entity.go
│   ├── errors.go
│   ├── repository.go
│   └── entity_test.go
├── usecase/payment/
│   ├── usecase.go
│   └── usecase_test.go
└── interface/http/payment/
    ├── handler.go
    ├── dto.go
    └── handler_test.go

db/
├── migrations/{timestamp}_payment.up.sql
├── migrations/{timestamp}_payment.down.sql
└── queries/payment.sql
```

---

## Linting

```bash
# Run linter
make lint

# Auto-fix issues
golangci-lint run --fix ./...
```

### Linter Configuration

See `.golangci.yml` for enabled linters and rules.

---

## Adding New Features

### 1. Adding a New Domain Module

1. Create domain: `internal/domain/{name}/entity.go`
2. Create errors: `internal/domain/{name}/errors.go`
3. Create repository interface: `internal/domain/{name}/repository.go`
4. Create migration: `db/migrations/`
5. Create queries: `db/queries/{name}.sql`
6. Run: `make sqlc`
7. Create usecase: `internal/usecase/{name}/usecase.go`
8. Create handler: `internal/interface/http/{name}/handler.go`
9. Register routes in `router.go`

### 2. Adding New Middleware

1. Create middleware in `internal/interface/http/middleware/`
2. Register in `router.go`
3. Add tests

### 3. Adding Background Jobs

1. Define task in `internal/worker/tasks/`
2. Register handler in `internal/worker/registry.go`
3. See `docs/async-jobs.md` for patterns

---

## Common Tasks

| Task | Command |
|------|---------|
| Start dev server | `make dev` |
| Run tests | `make test` |
| Run linter | `make lint` |
| Generate sqlc | `make sqlc` |
| Apply migrations | `make migrate-up` |
| Build all | `make build` |
| Clean build | `make clean` |

---

*Generated by BMad Method document-project workflow on 2025-12-15*
