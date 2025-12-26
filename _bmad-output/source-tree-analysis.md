# Source Tree Analysis

**Project:** golang-api-hexagonal  
**Analysis Date:** 2024-12-24  
**Repository Type:** Monolith  

## Directory Structure

```
golang-api-hexagonal/
├── cmd/                          # Application entry points
│   └── api/
│       └── main.go               # ★ Primary entry point - wires all dependencies
│
├── internal/                     # Private application code (Go convention)
│   │
│   ├── domain/                   # ★ CORE: Business entities & interfaces
│   │   ├── audit.go              # AuditEvent entity, event type constants
│   │   ├── audit_test.go         # Unit tests for audit entity
│   │   ├── errors.go             # Domain error types (ErrNotFound, etc.)
│   │   ├── id.go                 # UUID-based ID value object
│   │   ├── id_test.go            # Unit tests for ID
│   │   ├── pagination.go         # ListParams value object
│   │   ├── pagination_test.go    # Unit tests for pagination
│   │   ├── querier.go            # Querier interface (DB abstraction)
│   │   ├── redactor.go           # PII redaction config
│   │   ├── tx.go                 # TxManager interface (transactions)
│   │   ├── user.go               # User entity, UserRepository interface
│   │   └── user_test.go          # Unit tests for user entity
│   │
│   ├── app/                      # ★ APPLICATION: Use cases & services
│   │   ├── audit/
│   │   │   ├── audit_service.go  # AuditService with PII redaction
│   │   │   └── audit_service_test.go
│   │   ├── user/
│   │   │   ├── create_user.go    # CreateUserUseCase
│   │   │   ├── create_user_test.go
│   │   │   ├── get_user.go       # GetUserUseCase
│   │   │   ├── get_user_test.go
│   │   │   ├── list_users.go     # ListUsersUseCase
│   │   │   └── list_users_test.go
│   │   ├── auth.go               # Auth context helpers (ActorID extraction)
│   │   ├── auth_test.go
│   │   ├── errors.go             # Application error types
│   │   └── errors_test.go
│   │
│   ├── transport/                # ★ INBOUND ADAPTERS: HTTP delivery
│   │   ├── http/
│   │   │   ├── router.go         # Chi router configuration & middleware stack
│   │   │   ├── contract/         # Request/Response DTOs
│   │   │   │   ├── request.go    # Request binding helpers
│   │   │   │   ├── response.go   # Standard response format
│   │   │   │   ├── pagination.go # Pagination request/response
│   │   │   │   ├── problem.go    # RFC 7807 error responses
│   │   │   │   ├── user.go       # User DTOs
│   │   │   │   └── *_test.go     # Contract tests
│   │   │   ├── ctxutil/          # Context utilities
│   │   │   │   └── ctxutil.go    # RequestID, TraceID extraction
│   │   │   ├── handler/          # HTTP handlers
│   │   │   │   ├── health.go     # Liveness probe handler
│   │   │   │   ├── health_test.go
│   │   │   │   ├── ready.go      # Readiness probe handler
│   │   │   │   ├── ready_test.go
│   │   │   │   ├── user.go       # User CRUD handlers
│   │   │   │   ├── user_test.go
│   │   │   │   └── integration_test.go
│   │   │   └── middleware/       # HTTP middleware
│   │   │       ├── auth.go           # JWT authentication
│   │   │       ├── auth_bridge.go    # Auth context bridge
│   │   │       ├── body_limiter.go   # Request size limiting
│   │   │       ├── logging.go        # Structured request logging
│   │   │       ├── metrics.go        # Prometheus metrics
│   │   │       ├── ratelimit.go      # Rate limiting
│   │   │       ├── requestid.go      # Request ID generation
│   │   │       ├── response_wrapper.go # Response capture
│   │   │       ├── security.go       # OWASP security headers
│   │   │       ├── tracing.go        # OpenTelemetry tracing
│   │   │       └── *_test.go         # Middleware tests
│   │   └── internal_review_tmp/  # Temporary review directory
│   │
│   ├── infra/                    # ★ OUTBOUND ADAPTERS: External services
│   │   ├── config/
│   │   │   ├── config.go         # Environment config with validation
│   │   │   └── config_test.go
│   │   ├── observability/
│   │   │   ├── logger.go         # Structured JSON logger (slog)
│   │   │   ├── logger_test.go
│   │   │   ├── metrics.go        # Prometheus metrics factories
│   │   │   ├── metrics_test.go
│   │   │   ├── tracer.go         # OpenTelemetry tracer init
│   │   │   └── tracer_test.go
│   │   └── postgres/
│   │       ├── pool.go           # pgxpool wrapper
│   │       ├── querier.go        # Querier interface implementation
│   │       ├── tx_manager.go     # Transaction manager
│   │       ├── user_repo.go      # UserRepository implementation
│   │       ├── user_repo_test.go
│   │       ├── audit_repo.go     # AuditEventRepository implementation
│   │       ├── id_generator.go   # UUID generator
│   │       └── id_generator_test.go
│   │
│   └── shared/                   # Cross-cutting utilities
│       ├── ctxutil/              # Context helpers (if any)
│       ├── metrics/              # Shared metrics interfaces
│       │   └── http.go           # HTTPMetrics interface
│       └── redact/               # PII redaction utilities
│           ├── redactor.go       # PIIRedactor implementation
│           └── redactor_test.go
│
├── migrations/                   # Database migrations (Goose)
│   ├── 20251216000000_init.sql           # Schema info table
│   ├── 20251217000000_create_users.sql   # Users table
│   └── 20251219000000_create_audit_events.sql # Audit events table
│
├── docs/                         # Project documentation
│   ├── guides/                   # How-to guides
│   │   ├── adding-adapter.md     # Adding new adapters
│   │   ├── adding-audit-events.md # Adding audit event types
│   │   └── adding-module.md      # Adding new modules
│   ├── local-development.md      # Local dev setup guide
│   └── observability.md          # Observability configuration
│
├── .github/                      # GitHub configuration
│   └── workflows/
│       └── ci.yml                # CI pipeline (lint, test, build, migrate)
│
├── _bmad/                        # BMad Method configuration
│   └── ...                       # Agent configurations
│
├── _bmad-output/                 # BMad workflow outputs
│   └── bmm-workflow-status.yaml  # Workflow tracking
│
├── .env.example                  # Environment variable template
├── .env.docker                   # Docker-specific env
├── .golangci.yml                 # Linter configuration
├── docker-compose.yaml           # Local PostgreSQL
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── Makefile                      # Development commands
└── README.md                     # Project README
```

## Critical Directories

### `internal/domain/` - Business Core
- **Purpose**: Pure business logic with no external dependencies
- **Key Files**: Entity definitions, repository interfaces (ports)
- **Rules**: Only stdlib imports allowed

### `internal/app/` - Use Cases
- **Purpose**: Application workflows and service orchestration
- **Key Files**: Use case implementations, service classes
- **Rules**: Can import domain only

### `internal/transport/http/` - HTTP Delivery
- **Purpose**: Request handling, middleware, routing
- **Key Files**: Handlers, middleware, contract DTOs
- **Rules**: Can import domain and app

### `internal/infra/` - External Adapters
- **Purpose**: Implementation of domain interfaces
- **Key Files**: Database repos, config loader, observability
- **Rules**: Implements domain interfaces, uses external packages

## Entry Points

| File | Purpose |
|------|---------|
| `cmd/api/main.go` | Application bootstrap - loads config, wires dependencies, starts server |

## Integration Points

### Database
- **Technology**: PostgreSQL via pgxpool
- **Connection**: `internal/infra/postgres/pool.go`
- **Migrations**: `migrations/*.sql` (Goose format)

### External Services (Future)
- **Tracing**: OTLP endpoint (Jaeger/Tempo)
- **Metrics**: Prometheus scrape endpoint `/metrics`

## File Counts by Directory

| Directory | Files | Tests | Coverage |
|-----------|-------|-------|----------|
| domain | 12 | 4 | ≥80% |
| app | 12 | 6 | ≥80% |
| transport/http | 37 | ~18 | - |
| infra | 16 | 6 | - |
| shared | 5 | 2 | - |
| **Total** | **82** | **36** | - |
