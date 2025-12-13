# Backend Service Golang Boilerplate

Enterprise-grade "golden template" untuk membangun backend services di Go dengan **observability-first architecture**.

## Quick Start

```bash
# Clone repository
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal

# Download dependencies
go mod download

# Verify compilation
go build ./...
```

## Documentation

- [Architecture](docs/architecture.md) - Design decisions and patterns
- [PRD](docs/prd.md) - Product requirements

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24.x |
| Router | chi v5 |
| Database | PostgreSQL + pgx v5 |
| Query | sqlc |
| Logger | zap |
| Config | koanf v2 |
| Tracing | OpenTelemetry |

## Project Structure

```
├── cmd/
│   ├── server/         # Main application entry point
│   ├── worker/         # Background job worker
│   ├── scheduler/      # Cron job scheduler
│   └── bplat/          # CLI tool for scaffolding
├── internal/
│   ├── app/            # Application wiring
│   ├── config/         # Configuration
│   ├── domain/         # Business entities
│   ├── usecase/        # Business logic
│   ├── infra/          # Infrastructure adapters
│   ├── interface/      # HTTP handlers
│   ├── observability/  # Logging/tracing
│   └── runtimeutil/    # Utilities
├── db/
│   ├── migrations/     # SQL migrations
│   └── queries/        # sqlc queries
└── docs/               # Documentation
```

## CLI Tool (bplat)

The `bplat` CLI tool provides code scaffolding utilities:

```bash
# Build CLI tool
make build-bplat

# Check version
./bin/bplat version

# View help
./bin/bplat --help

# Initialize a new service
./bin/bplat init service myservice

# With custom module path
./bin/bplat init service myservice --module github.com/myorg/myservice
```


## Adding New Modules

Follow this guide to create a new domain module by copying the example `note` module.

### Step 1: Copy Source Files

```bash
# Replace "task" with your new module name
NEW_MODULE="task"

# Copy domain layer
cp -r internal/domain/note internal/domain/$NEW_MODULE

# Copy usecase layer
cp -r internal/usecase/note internal/usecase/$NEW_MODULE

# Copy HTTP handler layer
cp -r internal/interface/http/note internal/interface/http/$NEW_MODULE
```

### Step 2: Rename Package and Types

```bash
# macOS sed syntax (use sed -i'' for Linux)
NEW_MODULE="task"

# Update package names and types
find internal/domain/$NEW_MODULE -type f -name "*.go" \
  -exec sed -i '' "s/note/$NEW_MODULE/g; s/Note/Task/g" {} \;

find internal/usecase/$NEW_MODULE -type f -name "*.go" \
  -exec sed -i '' "s/note/$NEW_MODULE/g; s/Note/Task/g" {} \;

find internal/interface/http/$NEW_MODULE -type f -name "*.go" \
  -exec sed -i '' "s/note/$NEW_MODULE/g; s/Note/Task/g" {} \;
```

### Step 3: Create Database Migration

```bash
# Create migration files
touch db/migrations/$(date +%Y%m%d%H%M%S)_create_tasks.up.sql
touch db/migrations/$(date +%Y%m%d%H%M%S)_create_tasks.down.sql

# Create SQLC queries
touch db/queries/task.sql
```

### Step 4: Generate SQLC and Wire Up

```bash
# Generate SQLC code
make sqlc

# Register routes in router.go
# Import: "github.com/iruldev/golang-api-hexagonal/internal/interface/http/task"
# Add: taskHandler.Routes(r)
```

### Checklist

- [ ] **Domain Layer**
  - [ ] `internal/domain/{name}/entity.go` - Entity with `Validate()`
  - [ ] `internal/domain/{name}/errors.go` - Domain-specific errors
  - [ ] `internal/domain/{name}/repository.go` - Repository interface
  - [ ] `internal/domain/{name}/entity_test.go` - Entity tests

- [ ] **Use Case Layer**
  - [ ] `internal/usecase/{name}/usecase.go` - Business logic
  - [ ] `internal/usecase/{name}/usecase_test.go` - Unit tests with mocks

- [ ] **Interface Layer**
  - [ ] `internal/interface/http/{name}/handler.go` - HTTP handlers
  - [ ] `internal/interface/http/{name}/dto.go` - Request/Response DTOs
  - [ ] `internal/interface/http/{name}/handler_test.go` - Handler tests
  - [ ] `internal/interface/http/{name}/handler_integration_test.go` - Integration tests

- [ ] **Infrastructure Layer**
  - [ ] `db/migrations/{timestamp}_{description}.up.sql` - Up migration
  - [ ] `db/migrations/{timestamp}_{description}.down.sql` - Down migration
  - [ ] `db/queries/{name}.sql` - SQLC queries
  - [ ] Run `make sqlc` to generate repository

- [ ] **Wiring**
  - [ ] Register routes in `internal/interface/http/router.go`
  - [ ] Run tests: `go test ./...`
  - [ ] Run lint: `golangci-lint run ./...`

## Operability

### Prometheus Alerting

The service includes pre-configured Prometheus alerting rules:

```bash
# View alerting rules
cat deploy/prometheus/alerts.yaml
```

**Included Alerts:**
- **HTTP Service:** HighErrorRate, HighLatency, ServiceDown
- **Database:** DBConnectionExhausted, DBSlowQueries
- **Job Queue:** JobQueueBacklog, JobFailureRate

Rules are automatically loaded by Prometheus. See [AGENTS.md](AGENTS.md#-prometheus-alerting) for customization.

## License

MIT
