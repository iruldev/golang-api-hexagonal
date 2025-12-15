# Source Tree Analysis

Annotated directory structure for **Backend Service Golang Boilerplate**.

---

## Root Structure

```
golang-api-hexagonal/
├── .agent/                  # AI agent configurations
├── .bmad/                   # BMad Method workflow configs
├── .github/                 # GitHub Actions workflows
├── api/                     # OpenAPI specifications
├── cmd/                     # Application entry points ⭐
│   ├── bplat/               # CLI scaffolding tool
│   ├── scheduler/           # Cron job scheduler
│   ├── server/              # HTTP API server (main)
│   └── worker/              # Background job worker
├── db/                      # Database assets ⭐
│   ├── migrations/          # SQL migration files
│   └── queries/             # sqlc query definitions
├── deploy/                  # Deployment configurations
│   ├── grafana/             # Grafana dashboards
│   └── prometheus/          # Prometheus alerting rules
├── docs/                    # Documentation
│   ├── runbook/             # Incident response runbooks
│   └── async-jobs.md        # Async job patterns guide
├── internal/                # Private application code ⭐
│   ├── app/                 # Application wiring
│   ├── config/              # Configuration management
│   ├── domain/              # Domain layer (entities/interfaces)
│   ├── usecase/             # Use case layer (business logic)
│   ├── interface/           # Interface layer (adapters)
│   ├── infra/               # Infrastructure layer
│   ├── observability/       # Logging, metrics, tracing
│   ├── runtimeutil/         # Runtime utilities
│   ├── testing/             # Test helpers
│   └── worker/              # Background job implementations
├── proto/                   # Protocol buffer definitions
├── AGENTS.md                # AI assistant guide
├── ARCHITECTURE.md          # Architecture documentation
├── docker-compose.yaml      # Local development stack
├── go.mod                   # Go module definition
├── gqlgen.yml               # GraphQL code generation config
├── Makefile                 # Build and development tasks
└── sqlc.yaml                # sqlc code generation config
```

---

## Internal Package Structure (Hexagonal Architecture)

### Domain Layer (`internal/domain/`)

**Purpose:** Business entities, repository interfaces, and domain errors.

```
internal/domain/
├── auth/                    # Authentication domain
│   ├── rbac.go              # Role-based access control
│   └── rbac_test.go
├── note/                    # Note domain (reference implementation)
│   ├── entity.go            # Note entity with Validate()
│   ├── entity_test.go
│   ├── errors.go            # Domain-specific errors
│   └── repository.go        # Repository interface (port)
├── doc.go                   # Package documentation
├── errors.go                # Common domain errors
└── errors_test.go
```

**Key Pattern:** Each domain module contains:
- `entity.go` - Entity struct with `Validate()` method
- `errors.go` - Domain-specific error definitions
- `repository.go` - Repository interface (port)

---

### Use Case Layer (`internal/usecase/`)

**Purpose:** Application business logic orchestrating domain entities.

```
internal/usecase/
└── note/                    # Note use cases
    ├── usecase.go           # Business logic implementation
    ├── usecase_mock.go      # Mock for testing
    └── usecase_test.go      # Unit tests
```

**Key Pattern:** Usecases:
- Accept domain interfaces (dependency injection)
- Coordinate domain entities
- Apply business rules before persistence
- Only depend on domain layer

---

### Interface Layer (`internal/interface/`)

**Purpose:** Adapters for external communication (HTTP, gRPC, GraphQL).

```
internal/interface/
├── http/                    # HTTP API (primary)
│   ├── admin/               # Admin endpoints (feature flags, roles)
│   │   ├── features.go      # Feature flag management
│   │   ├── handler.go       # Admin health check
│   │   └── roles.go         # User role management
│   ├── handlers/            # Common handlers
│   │   └── health.go        # Health/readiness endpoints
│   ├── middleware/          # HTTP middleware (29 files)
│   │   ├── apikey.go        # API key authentication
│   │   ├── auth.go          # Auth interface
│   │   ├── jwt.go           # JWT authentication
│   │   ├── ratelimit.go     # Rate limiting
│   │   ├── rbac.go          # Role-based access control
│   │   ├── requestid.go     # Request ID propagation
│   │   ├── security.go      # Security headers
│   │   └── ...
│   ├── note/                # Note HTTP handlers
│   │   ├── dto.go           # Request/Response DTOs
│   │   ├── handler.go       # HTTP handlers
│   │   └── handler_test.go
│   ├── response/            # Response envelope pattern
│   │   └── envelope.go      # Standardized responses
│   ├── httpx/               # HTTP utilities
│   ├── router.go            # Chi router setup ⭐
│   ├── routes.go            # Route registration
│   └── routes_admin.go      # Admin routes
├── grpc/                    # gRPC API
│   ├── interceptor*.go      # gRPC interceptors
│   └── server.go            # gRPC server setup
└── graphql/                 # GraphQL API
    ├── generated.go         # gqlgen generated code
    ├── resolver.go          # Query resolvers
    └── schema.graphqls      # GraphQL schema
```

---

### Infrastructure Layer (`internal/infra/`)

**Purpose:** External service implementations (database, cache, message brokers).

```
internal/infra/
├── postgres/                # PostgreSQL implementation
│   ├── db.go                # Database pool setup
│   ├── db.sqlc.go           # sqlc generated code
│   ├── note_repository.go   # Note repository implementation
│   └── ...
├── redis/                   # Redis implementation
│   ├── client.go            # Redis client
│   ├── ratelimiter.go       # Distributed rate limiter
│   └── ...
├── kafka/                   # Kafka implementation
│   ├── publisher.go         # Event publisher
│   └── ...
├── rabbitmq/                # RabbitMQ implementation
│   └── ...
└── doc.go
```

---

### Worker Layer (`internal/worker/`)

**Purpose:** Background job processing with Asynq.

```
internal/worker/
├── worker.go                # Worker setup
├── registry.go              # Task handler registration
├── idempotency/             # Idempotency patterns
│   ├── handler.go           # Idempotent handler wrapper
│   └── store.go             # Idempotency store interface
├── patterns/                # Async job patterns
│   ├── fanout.go            # One-to-many processing
│   ├── fireandforget.go     # Best-effort processing
│   └── scheduled.go         # Cron-based scheduling
└── tasks/                   # Task definitions
    ├── cleanup.go           # Cleanup task
    ├── email.go             # Email notification task
    └── ...
```

---

### Supporting Packages

#### `internal/observability/`

```
internal/observability/
├── audit.go                 # Audit logging
├── logger.go                # Zap logger setup
├── metrics.go               # Prometheus metrics
├── tracer.go                # OpenTelemetry tracer
└── ...
```

#### `internal/config/`

```
internal/config/
├── config.go                # Configuration struct
├── loader.go                # Koanf loader
└── ...
```

#### `internal/runtimeutil/`

```
internal/runtimeutil/
├── featureflags.go          # Feature flag providers
├── events.go                # Event publishing interfaces
├── rate.go                  # Rate limiting utilities
└── ...
```

---

## Entry Points

### HTTP Server (`cmd/server/`)

Primary entry point for HTTP API.

```
cmd/server/
├── main.go                  # Application bootstrap
└── main_test.go             # Integration tests
```

### Background Worker (`cmd/worker/`)

Asynq-based job processor.

```
cmd/worker/
└── main.go                  # Worker bootstrap
```

### Cron Scheduler (`cmd/scheduler/`)

Scheduled job runner.

```
cmd/scheduler/
└── main.go                  # Scheduler bootstrap
```

### CLI Tool (`cmd/bplat/`)

Code scaffolding utility.

```
cmd/bplat/
├── main.go                  # CLI entry point
├── cmd/
│   ├── generate.go          # Generate commands
│   ├── init.go              # Init commands
│   ├── root.go              # Root command
│   └── version.go           # Version command
└── templates/               # Code generation templates
```

---

## Database Assets (`db/`)

```
db/
├── migrations/              # SQL migrations (golang-migrate)
│   ├── 000001_*.up.sql
│   ├── 000001_*.down.sql
│   └── ...
└── queries/                 # sqlc query definitions
    └── note.sql             # Note CRUD queries
```

---

## Deployment (`deploy/`)

```
deploy/
├── grafana/
│   └── dashboards/          # Pre-configured Grafana dashboards
│       └── service-dashboard.json
└── prometheus/
    ├── prometheus.yml       # Prometheus config
    └── alerts.yaml          # Alerting rules
```

---

*Generated by BMad Method document-project workflow on 2025-12-15*
