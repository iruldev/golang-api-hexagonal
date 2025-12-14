# Backend Service Golang Boilerplate

Enterprise-grade "golden template" untuk membangun backend services di Go dengan **observability-first architecture**.

## Quick Start

```bash
# Clone repository
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal

# Start dependencies (PostgreSQL, Redis, Prometheus, etc.)
docker-compose up -d

# Download dependencies
go mod download

# Run migrations
make migrate-up

# Run application
make dev
```

### GraphQL Playground

When running in development mode (`APP_ENV=development` or `APP_ENV=local`), you can explore the GraphQL API interactively:

- **URL:** `http://localhost:8080/playground`
- **GraphQL Endpoint:** `/query`

> ⚠️ **Security Note:** The playground is automatically disabled in `staging` and `production` environments.

## V2 Features (Platform Evolution)

V2 extends the golden template foundation with enterprise-grade capabilities:

### 🔄 Async Job Patterns

| Pattern | Use Case | Location |
|---------|----------|----------|
| **Fire-and-Forget** | Analytics, audit logs, non-critical tasks | `internal/worker/patterns/fireandforget.go` |
| **Scheduled Jobs** | Periodic cleanup, reports (cron syntax) | `cmd/scheduler/`, `internal/worker/patterns/scheduled.go` |
| **Fanout** | One event → multiple handlers | `internal/worker/patterns/fanout.go` |
| **Idempotency** | Prevent duplicate processing, critical operations | `internal/worker/idempotency/` |

**Built on:** [asynq](https://github.com/hibiken/asynq) + Redis

```go
// Enqueue fire-and-forget job
patterns.FireAndForget(ctx, client, tasks.NewEmailNotificationTask(userID))

// Register scheduled job (cron)
patterns.RegisterScheduledJob(scheduler, "0 2 * * *", tasks.TypeDailyCleanup)
```

### 🔐 Security Features

| Feature | Purpose | Location |
|---------|---------|----------|
| **JWT Authentication** | Token-based auth with claims | `internal/interface/http/middleware/jwt.go` |
| **API Key Auth** | Service-to-service authentication | `internal/interface/http/middleware/apikey.go` |
| **RBAC** | Role/Permission-based authorization | `internal/domain/auth/rbac.go` |
| **Auth Middleware Interface** | Pluggable auth providers (OIDC, SSO) | `internal/interface/http/middleware/auth.go` |

**Roles:** `Admin`, `Service`, `User` | **Permissions:** `resource:action` pattern

```go
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtAuth))
    r.Use(middleware.RequireRole(string(auth.RoleAdmin)))
    r.Delete("/users/{id}", deleteUserHandler)
})
```

### 🚦 Rate Limiting

| Type | Use Case | Location |
|------|----------|----------|
| **In-Memory** | Single instance, development | `internal/interface/http/middleware/ratelimit.go` |
| **Redis-Backed** | Multi-instance, distributed | `internal/infra/redis/ratelimiter.go` |

```go
// 100 requests per minute per IP
limiter := middleware.NewInMemoryRateLimiter(
    middleware.WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
)
r.Use(middleware.RateLimitMiddleware(limiter))
```

### 🎛 Feature Flags

| Provider | Use Case | Location |
|----------|----------|----------|
| **EnvProvider** | Environment variable flags (simple) | `internal/runtimeutil/featureflags.go` |
| **NopProvider** | Testing (all on/off) | `internal/runtimeutil/featureflags.go` |

```go
provider := runtimeutil.NewEnvFeatureFlagProvider()
if enabled, _ := provider.IsEnabled(ctx, "new_dashboard"); enabled {
    // New feature logic
}
```

### 📤 Event Publishing (V3 Preview)

| Type | Use Case | Location |
|------|----------|----------|
| **Kafka Publisher** | High-throughput event-driven communication | `internal/infra/kafka/publisher.go` |
| **Nop Publisher** | Testing, disabled mode | `internal/runtimeutil/events.go` |

```go
// Publish event synchronously (waits for ack)
event, _ := runtimeutil.NewEvent("user.created", map[string]string{"user_id": "123"})
publisher.Publish(ctx, "users", event)

// Publish asynchronously (fire-and-forget)
publisher.PublishAsync(ctx, "analytics", event)
```

**Configuration:**

| Variable | Default | Description |
|----------|---------|-------------|
| `KAFKA_ENABLED` | `false` | Enable Kafka publisher |
| `KAFKA_BROKERS` | `localhost:9092` | Broker addresses |
| `KAFKA_CLIENT_ID` | `golang-api-hexagonal` | Client identifier |

### 📥 Event Consuming (V3)

| Type | Use Case | Location |
|------|----------|----------|
| **EventConsumer Interface** | Swappable event consumption (Kafka, RabbitMQ, NATS) | `internal/runtimeutil/events.go` |
| **NopConsumer** | Testing, disabled mode | `internal/runtimeutil/events.go` |
| **MockConsumer** | Behavior verification in tests | `internal/runtimeutil/events.go` |

```go
// Subscribe to events (blocks until ctx cancelled)
handler := func(ctx context.Context, event runtimeutil.Event) error {
    // Process event...
    return nil
}

// Optional: Wrap with DLQ handler for reliability
dlqCfg := runtimeutil.DefaultDLQConfig()
dlqCfg.TopicName = "orders.dlq"
dlqHandler := runtimeutil.NewDLQHandler(handler, dlq, dlqCfg, consumerCfg)

consumer.Subscribe(ctx, "orders", dlqHandler)
```

### 🛠 CLI Tool (bplat)

| Command | Purpose |
|---------|---------|
| `bplat init service <name>` | Scaffold new service from template |
| `bplat generate module <name>` | Generate domain module with all layers |
| `bplat version` | Display version info |

See [CLI Documentation](#cli-tool-bplat) below for full usage.

### 📊 Observability Enhancements

- **Prometheus Alerting Rules:** Pre-configured alerts for HTTP, DB, and Job Queue (`deploy/prometheus/alerts.yaml`)
- **Grafana Dashboards:** Ready-to-import dashboards (`deploy/grafana/dashboards/`)
- **Runbook Documentation:** Standardized incident response (`docs/runbook/`)

### 📋 V2 Quick Reference

#### V2 Environment Variables

| Variable | Purpose | Default | Required |
|----------|---------|---------|----------|
| **Redis** ||||
| `REDIS_HOST` | Redis server host | `localhost` | Yes |
| `REDIS_PORT` | Redis server port | `6379` | Yes |
| `REDIS_PASSWORD` | Redis password | - | No |
| `REDIS_DB` | Redis database number | `0` | No |
| **Authentication** ||||
| `JWT_SECRET` | JWT signing key (≥32 bytes) | - | For JWT |
| `JWT_ISSUER` | JWT issuer validation | - | Optional |
| `JWT_AUDIENCE` | JWT audience validation | - | Optional |
| `API_KEYS` | API key-service pairs (`key:svc,key:svc`) | - | For API Key |
| **Feature Flags** ||||
| `FF_*` | Feature flags (e.g., `FF_NEW_FEATURE=true`) | `false` | No |

#### Job Queue Pattern Decision Table

| Scenario | Pattern | Package | Retry |
|----------|---------|---------|-------|
| Non-critical, best-effort (analytics, audit) | Fire-and-Forget | `patterns.FireAndForget()` | No |
| Periodic tasks (cleanup, reports) | Scheduled | `patterns.RegisterScheduledJobs()` | No |
| One event → multiple handlers | Fanout | `patterns.Fanout()` | Per-handler |
| Prevent duplicate processing | Idempotency | `idempotency.IdempotentHandler()` | Yes |
| Critical operations (payments) | Standard | Direct enqueue | Yes |

---

## Documentation

- [Architecture](docs/architecture.md) - Design decisions and patterns
- [PRD](docs/prd.md) - Product requirements
- [AGENTS.md](AGENTS.md) - AI assistant guide and patterns

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
| **V2 Features** | |
| Job Queue | asynq + Redis |
| Rate Limiting | Token bucket (in-memory/Redis) |
| Auth | JWT/API Key + RBAC |

## Migration from V1

If upgrading from V1 (Epics 1-7) to V2 (Epics 8-11), follow these steps:

### New Dependencies

```bash
# Add asynq for background jobs
go get github.com/hibiken/asynq

# Add Redis client (if not already present)
go get github.com/redis/go-redis/v9

# Add testcontainers for integration tests
go get github.com/testcontainers/testcontainers-go
```

### New Configuration Options

| Environment Variable | Purpose | Default | Required |
|---------------------|---------|---------|----------|
| `REDIS_HOST` | Redis server host | `localhost` | Yes (V2) |
| `REDIS_PORT` | Redis server port | `6379` | Yes (V2) |
| `REDIS_PASSWORD` | Redis password | `` | No |
| `REDIS_DB` | Redis database number | `0` | No |
| `JWT_SECRET` | JWT signing key (≥32 bytes) | - | For JWT auth |
| `JWT_ISSUER` | JWT issuer validation | - | Optional |
| `JWT_AUDIENCE` | JWT audience validation | - | Optional |
| `API_KEYS` | API key-service pairs | - | For API key auth |
| `FF_*` | Feature flags (e.g., `FF_NEW_FEATURE=true`) | `false` | No |

### Breaking Changes

**No breaking changes in V2.** All V2 features are additive. Existing V1 code continues to work.

### Upgrade Steps

1. **Add Redis** (required for job queue):
   ```bash
   # Add to docker-compose.yaml if not present
   docker-compose up -d redis
   ```

2. **Configure environment** (add to `.env`):
   ```bash
   REDIS_HOST=localhost
   REDIS_PORT=6379
   ```

3. **Enable V2 features** (optional imports):
   ```go
   // For async jobs
   import "github.com/iruldev/golang-api-hexagonal/internal/worker"
   
   // For security middleware  
   import "github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
   
   // For feature flags
   import "github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
   ```

4. **Run migrations** (if using new tables):
   ```bash
   make migrate-up
   ```

### Feature Matrix: V1 vs V2

| Capability | V1 (Epics 1-7) | V2 (Epics 8-11) |
|------------|---------------|-----------------|
| HTTP API | ✅ Chi router, middleware | ✅ + Auth, Rate limiting |
| Database | ✅ PostgreSQL, sqlc | ✅ Same |
| Observability | ✅ Metrics, logging, tracing | ✅ + Alerting rules, dashboards |
| Background Jobs | ❌ | ✅ Fire-and-forget, scheduled, fanout |
| Authentication | ❌ | ✅ JWT, API Key |
| Authorization | ❌ | ✅ RBAC with roles/permissions |
| Rate Limiting | ❌ | ✅ In-memory, Redis-backed |
| Feature Flags | ❌ | ✅ Env-based provider |
| CLI Scaffolding | ❌ | ✅ bplat tool |
| Integration Tests | ✅ Manual setup | ✅ Testcontainers (self-contained) |

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

The `bplat` CLI tool provides code scaffolding utilities for rapid development.

### Quick Reference

| Command | Description | Example |
|---------|-------------|---------|
| `bplat version` | Print version info | `./bin/bplat version` |
| `bplat init service <name>` | Initialize new service | `./bin/bplat init service myapi` |
| `bplat generate module <name>` | Generate domain module | `./bin/bplat generate module payment` |
| `bplat --help` | Show all commands | `./bin/bplat --help` |

### Building and Installing

```bash
# Build to bin/ directory
make build-bplat

# Install to GOPATH/bin (available system-wide)
make install-bplat
```

### Init Service

Create a new service from the boilerplate template:

```bash
# Basic usage
bplat init service myservice

# With custom module path
bplat init service myservice --module github.com/myorg/myservice

# In a specific directory
bplat init service myservice --dir /path/to/projects

# Overwrite existing directory
bplat init service myservice --force
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--module` | `-m` | `github.com/user/<name>` | Go module path |
| `--dir` | `-d` | `.` | Output directory |
| `--force` | `-f` | `false` | Overwrite existing directory |

### Generate Module

Create a new domain module with all hexagonal architecture layers:

```bash
# Basic usage - creates payment module with Payment entity
bplat generate module payment

# With custom entity name
bplat generate module orders --entity Order
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--entity` | `-e` | Singularized module name | Custom entity name (PascalCase) |

**Generated Structure:**

```
internal/
├── domain/payment/           # Domain layer
│   ├── entity.go             # Entity with Validate()
│   ├── errors.go             # Domain-specific errors
│   ├── repository.go         # Repository interface
│   └── entity_test.go        # Entity tests
├── usecase/payment/          # Use case layer
│   ├── usecase.go            # Business logic
│   └── usecase_test.go       # Use case tests
└── interface/http/payment/   # Interface layer
    ├── handler.go            # HTTP handlers
    ├── dto.go                # Request/Response DTOs
    └── handler_test.go       # Handler tests

db/
├── migrations/
│   ├── {timestamp}_payment.up.sql
│   └── {timestamp}_payment.down.sql
└── queries/
    └── payment.sql           # sqlc queries
```

**Template Variables:**

| Variable | Description | Example |
|----------|-------------|---------|
| `ModuleName` | Lowercase module name | `payment` |
| `EntityName` | PascalCase entity name | `Payment` |
| `TableName` | Snake_case plural | `payments` |
| `Timestamp` | Migration timestamp | `20251214021630` |
| `ModulePath` | Go module path | `github.com/iruldev/golang-api-hexagonal` |

**Next Steps After Generation:**

1. Review and update entity fields in `internal/domain/{name}/entity.go`
2. Update migration in `db/migrations/{timestamp}_{name}.up.sql`
3. Update sqlc queries in `db/queries/{name}.sql`
4. Run: `make sqlc`
5. Register routes in `internal/interface/http/router.go`

---

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

### Runbook Documentation

Each alert has a corresponding runbook for incident response:

- **Location:** `docs/runbook/`
- **Template:** `docs/runbook/template.md`
- **Index:** `docs/runbook/README.md`

Runbooks include: symptoms, diagnosis steps, common causes, remediation, and escalation paths.

## License

MIT
