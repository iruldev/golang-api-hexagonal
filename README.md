# golang-api-hexagonal

[![CI](https://github.com/iruldev/golang-api-hexagonal/actions/workflows/ci.yml/badge.svg)](https://github.com/iruldev/golang-api-hexagonal/actions/workflows/ci.yml)
<!-- Badge updates automatically after CI runs with GIST_SECRET/GIST_ID configured -->
<!-- Coverage badge (requires setup, see CI workflow) -->
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/iruldev/GIST_ID/raw/coverage.json)](https://github.com/iruldev/golang-api-hexagonal/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/iruldev/golang-api-hexagonal)](https://goreportcard.com/report/github.com/iruldev/golang-api-hexagonal)

A production-ready Go API built with hexagonal architecture, featuring comprehensive observability, security, and developer experience.

## 🚀 Quick Start

⏱️ **Time to first API call: ~10 minutes**

### Option 1: One-Command Setup (Recommended)

```bash
# Clone and run quick-start
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal
make quick-start
```

This command will automatically:
1. ✅ Check prerequisites (Go, Docker)
2. ✅ Install development tools
3. ✅ Start PostgreSQL
4. ✅ Run database migrations
5. ✅ Verify everything works

After completion, run `make run` to start the API server.

### Option 2: Step-by-Step Setup

```bash
# 1. Clone and bootstrap
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal
make bootstrap  # Install dev tools (first time only)
make setup      # Setup project dependencies

# 2. Start infrastructure (PostgreSQL)
make infra-up

# 3. Run migrations
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
make migrate-up

# 4. Run the service
make run

# 5. Verify endpoints
curl http://localhost:8080/health
# Expected: {"data":{"status":"ok"}}

curl http://localhost:8080/ready
# Expected: {"data":{"status":"ready","checks":{"database":"ok"}}}

curl http://localhost:8081/metrics
# Expected: Prometheus metrics in text format
```

### 🧪 First API Call (Optional)

Once the service is running, try the Users API:

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","firstName":"John","lastName":"Doe"}'
# Expected: {"data":{"id":"<uuid>","email":"john@example.com",...}}

# Get the user (replace <id> with the returned UUID)
curl http://localhost:8080/api/v1/users/<id>

# List all users
curl http://localhost:8080/api/v1/users
# Expected: {"data":[...],"pagination":{"page":1,"pageSize":20,...}}
```

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

| Tool | Version | Check Command | Download |
|------|---------|---------------|----------|
| **Go** | 1.24+ | `go version` | [go.dev/dl](https://go.dev/dl/) |
| **Docker** | Latest | `docker --version` | [docker.com](https://www.docker.com/products/docker-desktop/) |
| **Make** | Any | `make --version` | Usually pre-installed |

**Quick check:** Run `make check-prereqs` to verify all prerequisites are met.

## 🛠️ Make Targets

Run `make help` to see all available targets:

### Quick Start & Verification
| Target | Description |
|--------|-------------|
| `make quick-start` | Complete setup from clone to running API (~10 min) |
| `make check-prereqs` | Verify all prerequisites are installed |
| `make verify-setup` | Check if environment is correctly configured |

### Development
| Target | Description |
|--------|-------------|
| `make bootstrap` | Install dev tools with pinned versions (first time) |
| `make setup` | Install development tools and dependencies |
| `make build` | Build the application binary |
| `make run` | Run the application |
| `make test` | Run all tests with race detector |
| `make lint` | Run golangci-lint |
| `make clean` | Clean build artifacts |

### Infrastructure
| Target | Description |
|--------|-------------|
| `make infra-up` | Start PostgreSQL (waits for healthy) |
| `make infra-down` | Stop infrastructure (preserve data) |
| `make infra-reset` | Stop and remove volumes (DESTRUCTIVE) |
| `make infra-logs` | View container logs |
| `make infra-status` | Show container status |

### Database Migrations
| Target | Description |
|--------|-------------|
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Rollback the last migration |
| `make migrate-status` | Show migration status |
| `make migrate-create name=X` | Create new migration |
| `make migrate-validate` | Validate migration files |

## 🏗️ Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters):

```
cmd/
└── api/                    # Application entry point
    └── main.go

internal/
├── domain/                 # Business logic (entities, value objects)
├── app/                    # Application services (use cases)
├── transport/              # Inbound adapters
│   └── http/
│       ├── router.go
│       └── handler/        # HTTP handlers
└── infra/                  # Outbound adapters
    ├── config/             # Configuration management
    └── postgres/           # Database implementation
```

### Layer Rules

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| Domain | stdlib only | app, transport, infra |
| App | domain | transport, infra |
| Transport | domain, app | infra |
| Infra | domain, external packages | app, transport |

## 🔧 Configuration

Configuration via environment variables (using `envconfig`):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | ✅ | - | PostgreSQL connection string |
| `PORT` | ❌ | `8080` | HTTP server port |
| `INTERNAL_PORT` | ❌ | `8081` | Internal server port (metrics) |
| `INTERNAL_BIND_ADDRESS` | ❌ | `127.0.0.1` | Internal bind address |
| `LOG_LEVEL` | ❌ | `info` | Logging level (debug, info, warn, error) |
| `ENV` | ❌ | `development` | Environment (development, staging, production, test) |
| `SERVICE_NAME` | ❌ | `golang-api-hexagonal` | Service name for observability |

Example:
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
export PORT=8080
export INTERNAL_PORT=8081
export LOG_LEVEL=info
export ENV=development
```

## 📡 API Endpoints

### Health Checks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Liveness probe (no DB check) |
| `/ready` | GET | Readiness probe (checks DB) |

**Response format:**
```json
{
  "data": {
    "status": "ok"
  }
}
```

## 🔐 Internal Port (Metrics Protection)

The application runs **two HTTP servers** to isolate internal endpoints:

| Server | Port | Endpoints | Purpose |
|--------|------|-----------|---------|
| **Public** | `PORT` (8080) | `/health`, `/ready`, `/api/v1/*` | External traffic |
| **Internal** | `INTERNAL_PORT` (8081) | `/metrics` | Metrics scraping only |

### Why Separate Ports?

The `/metrics` endpoint is protected on a separate internal port to:
- **Prevent information disclosure**: Metrics can reveal sensitive operational data
- **Enable network-level isolation**: Firewall rules can restrict access to internal port
- **Support secure Prometheus scraping**: Only internal services can scrape metrics

### Accessing /metrics

```bash
# Local development
curl http://localhost:8081/metrics

# Docker (expose internal port)
docker run -p 8080:8080 -p 8081:8081 your-image

# Kubernetes: Create ClusterIP service for internal port
apiVersion: v1
kind: Service
metadata:
  name: myapp-internal
spec:
  type: ClusterIP
  ports:
    - port: 8081
      targetPort: 8081
      name: metrics
```

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `INTERNAL_PORT` | `8081` | Port for internal endpoints |
| `INTERNAL_BIND_ADDRESS` | `127.0.0.1` | Bind address (loopback for security) |

> **Note:** In production, set `INTERNAL_BIND_ADDRESS=0.0.0.0` if exposing to container network.

## 🗄️ Database Migrations


Migrations use [goose](https://github.com/pressly/goose) and are located in `migrations/`.

**File format:** `YYYYMMDDHHMMSS_description.sql`

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE ...;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ...;
-- +goose StatementEnd
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/transport/http/handler/...
```

For detailed testing conventions, patterns, and best practices, see [Testing Patterns](docs/testing-patterns.md).

For copy-paste friendly code templates, see [Code Patterns](docs/patterns.md) and [Copy-Paste Kit](docs/copy-paste-kit/README.md).

For architectural decisions and their rationale, see [Architecture Decision Records](docs/adr/index.md).

For operational incident response procedures, see [Operational Runbooks](docs/runbooks/index.md).

### Coverage

Coverage is enforced at **≥80%** for `internal/domain/...` and `internal/app/...` packages.

```bash
# Check coverage meets threshold (fails if < 80%)
make coverage

# Generate HTML coverage report for visual review
make coverage-html
open coverage.html  # macOS

# Show per-package coverage breakdown
make coverage-report

# Show uncovered functions for PR review
make coverage-detail
```

**Coverage Targets:**
| Target | Description |
|--------|-------------|
| `make coverage` | Check 80% threshold (fails CI if below) |
| `make coverage-html` | Generate HTML report |
| `make coverage-report` | Per-package breakdown |
| `make coverage-detail` | Show uncovered functions |

> **Note:** Only domain and app layers are subject to coverage requirements. Infrastructure and transport layers are tested via integration tests.


## 📁 Project Structure

```
.
├── cmd/api/                # Application entry point
├── internal/               # Private application code
│   ├── domain/             # Business entities
│   ├── app/                # Use cases
│   ├── transport/http/     # HTTP handlers
│   └── infra/              # Infrastructure (config, postgres)
├── migrations/             # Database migrations (goose)
├── docs/                   # Documentation
├── .github/workflows/      # CI/CD pipelines
├── docker-compose.yaml     # Local infrastructure
├── Makefile                # Development commands
├── .env.example            # Environment template
└── go.mod                  # Go modules
```

## 🔧 Troubleshooting

### Docker Issues

**❌ Error: "Docker is not running"**
```
❌ Docker daemon not running - start Docker Desktop
```
**Solution:** Start Docker Desktop application and wait for it to fully initialize.

**❌ Error: "Port 5432 already in use"**
```
Error response from daemon: Ports are not available: exposing port TCP 0.0.0.0:5432
```
**Solution:** Stop the conflicting service or change the port:
```bash
# Find what's using port 5432
lsof -i :5432

# Stop local PostgreSQL if running
brew services stop postgresql  # macOS
sudo systemctl stop postgresql  # Linux
```

### Database Issues

**❌ Error: "DATABASE_URL not set"**
```
❌ DATABASE_URL is not set.
```
**Solution:** Export the variable before running migrations:
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
make migrate-up
```

**❌ Error: "Connection refused"**
```
dial tcp 127.0.0.1:5432: connect: connection refused
```
**Solution:** Ensure PostgreSQL is running:
```bash
make infra-up
make infra-status  # Should show "running"
```

### Go Issues

**❌ Error: "Go version mismatch"**
```
go: go.mod requires go >= 1.24
```
**Solution:** Update Go to version 1.24 or higher:
```bash
# Check current version
go version

# Download latest from: https://go.dev/dl/
```

### Verification Commands

Use these commands to diagnose issues:

| Command | Purpose |
|---------|---------|
| `make check-prereqs` | Verify all prerequisites are installed |
| `make verify-setup` | Check if development environment is ready |
| `make infra-status` | Show PostgreSQL container status |
| `make infra-logs` | View PostgreSQL container logs |

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.
