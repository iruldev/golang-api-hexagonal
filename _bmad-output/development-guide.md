# Development Guide

**Project:** golang-api-hexagonal  
**Last Updated:** 2024-12-24  

## Prerequisites

- **Go** 1.24+ ([download](https://go.dev/dl/))
- **Docker** & Docker Compose ([download](https://docs.docker.com/get-docker/))
- **Make** (usually pre-installed on macOS/Linux)

## Quick Start

```bash
# 1. Clone and setup
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal
make setup

# 2. Start infrastructure
make infra-up

# 3. Run migrations
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
make migrate-up

# 4. Run the service
make run

# 5. Verify
curl http://localhost:8080/health
```

---

## Development Setup

### 1. Install Tools

```bash
make setup
```

This installs:
- `golangci-lint` - Linter
- `goose` - Database migrations
- Go module dependencies

### 2. Start Infrastructure

```bash
make infra-up
```

This starts PostgreSQL in Docker with:
- Host: `localhost:5432`
- User: `postgres`
- Password: `postgres`
- Database: `golang_api_hexagonal`

### 3. Configure Environment

```bash
# Required
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"

# Optional (defaults shown)
export PORT=8080
export LOG_LEVEL=info
export ENV=development
```

Or source from `.env.example`:

```bash
cp .env.example .env
source .env
```

### 4. Run Migrations

```bash
make migrate-up
```

### 5. Run the Service

```bash
make run
```

---

## Make Targets

### Development

| Target | Description |
|--------|-------------|
| `make setup` | Install tools and dependencies |
| `make build` | Build binary |
| `make run` | Run application |
| `make test` | Run tests with race detector |
| `make coverage` | Check coverage threshold (≥80%) |
| `make lint` | Run linter |
| `make clean` | Clean build artifacts |

### CI Pipeline

| Target | Description |
|--------|-------------|
| `make ci` | Run full CI locally (tidy, fmt, lint, test) |
| `make check-mod-tidy` | Verify go.mod is tidy |
| `make check-fmt` | Verify code formatting |

### Infrastructure

| Target | Description |
|--------|-------------|
| `make infra-up` | Start PostgreSQL (waits for healthy) |
| `make infra-down` | Stop PostgreSQL (preserve data) |
| `make infra-reset` | Stop and delete volumes (DESTRUCTIVE) |
| `make infra-logs` | View container logs |
| `make infra-status` | Show container status |

### Migrations

| Target | Description |
|--------|-------------|
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-status` | Show migration status |
| `make migrate-create name=X` | Create new migration |
| `make migrate-validate` | Validate migration syntax |

### OpenAPI

| Target | Description |
|--------|-------------|
| `make openapi` | Validate OpenAPI spec (via Docker/Spectral) |
| `make openapi-view` | Preview API docs in browser |

---

## Testing

### Run All Tests

```bash
make test
```

### Run With Coverage

```bash
make coverage
```

This checks that `domain` and `app` packages maintain ≥80% coverage.

### Run Specific Package

```bash
go test -v ./internal/domain/...
go test -v ./internal/app/user/...
```

### Run Integration Tests

Integration tests require a running database:

```bash
make infra-up
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
go test -v ./internal/transport/http/handler/...
```

---

## Code Style

### Linting

```bash
make lint
```

Configuration in `.golangci.yml` includes:
- Import ordering
- Unused code detection
- Error handling checks
- **depguard** - Enforces hexagonal architecture boundaries

### Formatting

```bash
gofmt -w .
```

---

## Project Structure

```
cmd/api/                # Entry point
internal/
├── domain/             # Business entities (no external deps)
├── app/                # Use cases
├── transport/http/     # HTTP handlers & middleware
└── infra/              # Config, postgres, observability
migrations/             # Database migrations
docs/                   # Documentation
```

### Layer Rules

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| Domain | stdlib | app, transport, infra |
| App | domain | transport, infra |
| Transport | domain, app | infra |
| Infra | domain, external | app, transport |

---

## Common Tasks

### Create New Migration

```bash
make migrate-create name=add_orders_table
```

### Add New Endpoint

1. Add domain entity/interface in `internal/domain/`
2. Create use case in `internal/app/{module}/`
3. Implement repository in `internal/infra/postgres/`
4. Add handler in `internal/transport/http/handler/`
5. Register route in `internal/transport/http/router.go`
6. Wire in `cmd/api/main.go`

See [Adding Module Guide](./guides/adding-module.md) for details.

### Enable JWT Authentication

```bash
export JWT_ENABLED=true
export JWT_SECRET="your-super-secret-key-at-least-32-bytes-long"
make run
```

### Enable Tracing

```bash
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
make run
```

---

## Troubleshooting

### Database Connection Failed

```bash
# Check if PostgreSQL is running
make infra-status

# Check logs
make infra-logs

# Restart infrastructure
make infra-down
make infra-up
```

### Migration Errors

```bash
# Check current status
make migrate-status

# Validate syntax
make migrate-validate

# Rollback and retry
make migrate-down
make migrate-up
```

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

---

## IDE Setup

### VS Code

Recommended extensions:
- Go (Official)
- Error Lens
- GitLens

Settings (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "gofmt",
  "editor.formatOnSave": true
}
```

### GoLand / IntelliJ

1. Enable golangci-lint integration
2. Configure File Watchers for gofmt
3. Set Go SDK to 1.24+
