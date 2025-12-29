# Source Tree Analysis: golang-api-hexagonal

> **Exhaustive Scan** - Annotated Directory Structure
> **Tanggal:** 2025-12-29
> **Total Files:** 140 Go files

---

## Struktur Lengkap dengan Anotasi

```
golang-api-hexagonal/
│
├── cmd/                                # Entry Points
│   └── api/
│       ├── main.go                     # Application bootstrap dengan Uber Fx
│       │                               # - Loads config
│       │                               # - Initializes DI container
│       │                               # - Starts HTTP servers (public + internal)
│       ├── wiring_test.go              # DI wiring validation tests
│       ├── migration_test.go           # Migration integration tests
│       └── smoke_test.go               # Smoke tests for application startup
│
├── internal/                           # Private Application Code
│   │
│   ├── domain/                         # DOMAIN LAYER (Business Core)
│   │   │                               # RULE: Hanya boleh import stdlib!
│   │   │
│   │   ├── user.go                     # User entity + UserRepository interface
│   │   ├── user_test.go                # Unit tests untuk User validation
│   │   │
│   │   ├── audit.go                    # AuditEvent entity + AuditEventRepository
│   │   ├── audit_test.go               # Unit tests untuk AuditEvent
│   │   │
│   │   ├── id.go                       # ID value object (UUID wrapper)
│   │   ├── id_test.go                  # Unit tests untuk ID
│   │   │
│   │   ├── pagination.go               # ListParams value object
│   │   ├── pagination_test.go          # Unit tests untuk pagination
│   │   │
│   │   ├── querier.go                  # Querier interface (DB abstraction)
│   │   ├── tx.go                       # TxManager interface
│   │   ├── redactor.go                 # Redactor interface
│   │   └── errors.go                   # Domain errors (ErrNotFound, etc.)
│   │
│   ├── app/                            # APPLICATION LAYER (Use Cases)
│   │   │                               # RULE: Hanya boleh import domain!
│   │   │
│   │   ├── user/                       # User use cases
│   │   │   ├── create_user.go          # CreateUserUseCase
│   │   │   ├── create_user_test.go     # Unit tests dengan mocks
│   │   │   ├── get_user.go             # GetUserUseCase
│   │   │   ├── get_user_test.go        # Unit tests dengan mocks
│   │   │   ├── list_users.go           # ListUsersUseCase
│   │   │   └── list_users_test.go      # Unit tests dengan mocks
│   │   │
│   │   ├── audit/                      # Audit service
│   │   │   ├── service.go              # AuditService
│   │   │   └── service_test.go         # Unit tests
│   │   │
│   │   ├── auth.go                     # AuthParser interface
│   │   ├── auth_test.go                # Unit tests
│   │   ├── errors.go                   # AppError type
│   │   └── errors_test.go              # Unit tests
│   │
│   ├── transport/                      # TRANSPORT LAYER (Inbound Adapters)
│   │   │
│   │   └── http/                       # HTTP transport
│   │       │
│   │       ├── router.go               # Chi router setup + middleware chain
│   │       ├── router_test.go          # Router integration tests
│   │       │
│   │       ├── handler/                # HTTP handlers
│   │       │   ├── health.go           # /health endpoint
│   │       │   ├── health_test.go
│   │       │   ├── ready.go            # /ready endpoint (checks DB)
│   │       │   ├── ready_test.go
│   │       │   ├── user.go             # /api/v1/users endpoints
│   │       │   ├── user_test.go        # Unit tests (22KB - comprehensive!)
│   │       │   ├── helpers_test.go     # Test helpers
│   │       │   ├── integration_test.go # Integration tests
│   │       │   ├── integration_idor_test.go  # IDOR security tests
│   │       │   └── metrics_audit_test.go     # Metrics + audit tests
│   │       │
│   │       ├── middleware/             # HTTP middleware (22 files)
│   │       │   ├── auth.go             # JWT authentication
│   │       │   ├── auth_test.go        # Comprehensive auth tests (23KB)
│   │       │   ├── auth_bridge.go      # Auth adapter
│   │       │   ├── auth_bridge_test.go
│   │       │   ├── auth_test_helper_test.go
│   │       │   ├── logging.go          # Request logging
│   │       │   ├── logging_test.go
│   │       │   ├── metrics.go          # Prometheus metrics
│   │       │   ├── metrics_test.go
│   │       │   ├── tracing.go          # OpenTelemetry tracing
│   │       │   ├── tracing_test.go
│   │       │   ├── security.go         # Security headers (OWASP)
│   │       │   ├── security_test.go
│   │       │   ├── ratelimit.go        # Rate limiting
│   │       │   ├── ratelimit_test.go
│   │       │   ├── body_limiter.go     # Request size limit
│   │       │   ├── body_limiter_test.go
│   │       │   ├── requestid.go        # X-Request-ID
│   │       │   ├── requestid_test.go
│   │       │   ├── response_wrapper.go # JSON envelope
│   │       │   └── response_wrapper_test.go
│   │       │
│   │       ├── contract/               # Request/Response DTOs
│   │       │   ├── json.go             # JSON utilities
│   │       │   ├── json_test.go
│   │       │   ├── error.go            # Error response (RFC 7807)
│   │       │   ├── error_test.go
│   │       │   ├── user.go             # User DTOs
│   │       │   ├── user_test.go
│   │       │   ├── response.go         # Generic response wrapper
│   │       │   └── validation.go       # Request validation utilities
│   │       │
│   │       └── ctxutil/                # Context utilities
│   │           ├── claims.go           # JWT claims context
│   │           ├── claims_test.go
│   │           ├── trace.go            # Trace context
│   │           ├── trace_test.go
│   │           └── request_id.go       # Request ID context
│   │
│   ├── infra/                          # INFRASTRUCTURE LAYER (Outbound Adapters)
│   │   │
│   │   ├── postgres/                   # PostgreSQL implementation
│   │   │   ├── pool.go                 # Connection pool config
│   │   │   ├── pool_test.go
│   │   │   ├── resilient_pool.go       # Auto-reconnecting pool
│   │   │   ├── resilient_pool_test.go
│   │   │   ├── querier.go              # Querier implementation
│   │   │   ├── tx_manager.go           # Transaction manager
│   │   │   ├── user_repo.go            # UserRepository implementation
│   │   │   ├── user_repo_test.go
│   │   │   ├── audit_event_repo.go     # AuditEventRepository implementation
│   │   │   ├── audit_event_repo_test.go
│   │   │   ├── id_generator.go         # ID generation (UUID v7)
│   │   │   ├── citext_integration_test.go  # CITEXT integration test
│   │   │   │
│   │   │   └── sqlcgen/                # Generated code (sqlc)
│   │   │       ├── db.go
│   │   │       ├── models.go
│   │   │       ├── querier.go
│   │   │       ├── users.sql.go
│   │   │       └── audit.sql.go
│   │   │
│   │   ├── config/                     # Configuration
│   │   │   ├── config.go               # Config struct + loading
│   │   │   ├── config_test.go
│   │   │   └── config_pool_validation_test.go
│   │   │
│   │   ├── observability/              # Observability setup
│   │   │   ├── logger.go               # slog setup
│   │   │   ├── logger_test.go
│   │   │   ├── metrics.go              # Prometheus registry
│   │   │   ├── metrics_test.go
│   │   │   ├── tracer.go               # OTEL tracer
│   │   │   └── tracer_test.go
│   │   │
│   │   └── fx/                         # Uber Fx DI modules
│   │       ├── module.go               # All DI wiring
│   │       └── module_test.go          # DI graph tests
│   │
│   ├── shared/                         # SHARED (Cross-cutting)
│   │   │
│   │   ├── metrics/                    # Metrics interfaces
│   │   │   └── http_metrics.go         # HTTPMetrics interface
│   │   │
│   │   └── redact/                     # PII redaction
│   │       ├── redactor.go
│   │       ├── redactor_test.go
│   │       ├── benchmark_test.go       # Performance tests
│   │       └── robustness_test.go      # Edge case tests
│   │
│   └── testutil/                       # TEST UTILITIES (Shared)
│       │
│       ├── testutil.go                 # Common test utilities
│       │
│       ├── assert/                     # Custom assertions
│       │   └── assert.go               # Domain-aware assertions
│       │
│       ├── fixtures/                   # Test fixtures
│       │   └── fixtures.go             # Standard test data
│       │
│       ├── mocks/                      # Generated mocks (mockgen)
│       │   ├── doc.go                  # Package documentation
│       │   ├── user_repository_mock.go # Mock for UserRepository
│       │   └── audit_event_repository_mock.go  # Mock for AuditEventRepository
│       │
│       └── containers/                 # Testcontainers integration
│           ├── README.md               # Container usage guide
│           ├── containers.go           # Container orchestration
│           ├── postgres.go             # PostgreSQL testcontainer
│           ├── postgres_test.go        # Container tests
│           ├── migrate.go              # Migration runner
│           ├── truncate.go             # Table truncation
│           └── tx.go                   # Transaction helpers
│
├── migrations/                         # Database Migrations (goose)
│   ├── 20251216000000_init.sql         # Initial schema
│   ├── 20251217000000_create_users.sql # Users table
│   ├── 20251219000000_create_audit_events.sql  # Audit events
│   └── 20251226084756_add_citext_email.sql     # CITEXT for email
│
├── queries/                            # sqlc Query Definitions
│   ├── users.sql                       # User queries
│   └── audit_events.sql                # Audit queries
│
├── docs/                               # Documentation
│   ├── index.md                        # Master index
│   ├── architecture.md                 # Architecture docs
│   ├── openapi.yaml                    # OpenAPI 3.1 spec
│   ├── observability.md                # Observability guide
│   ├── local-development.md            # Dev setup guide
│   └── guides/                         # Additional guides
│
├── .github/workflows/                  # CI/CD
│   └── ci.yml                          # GitHub Actions workflow
│
└── Configuration Files
    ├── go.mod                          # Go modules
    ├── go.sum                          # Dependencies lock
    ├── Makefile                        # Development commands
    ├── .gitignore
    ├── .gitleaks.toml                  # Secret scanning config
    ├── .golangci.yml                   # Linter config + layer rules
    ├── .spectral.yaml                  # OpenAPI linting
    ├── .env.example                    # Environment template
    ├── docker-compose.yaml             # Local infrastructure
    └── sqlc.yaml                       # sqlc configuration
```

---

## Statistik Kode

| Metrik | Jumlah |
|--------|--------|
| **Total Go Files** | 140 |
| **Production Files** | ~70 |
| **Test Files** | ~70 |
| **Domain Layer** | 12 files |
| **Application Layer** | 11 files |
| **Transport Layer** | 36 files |
| **Infrastructure Layer** | 24 files |
| **Test Utilities** | 13 files |
| **Migrations** | 4 files |

---

## Critical Entry Points

| Entry Point | Path | Purpose |
|-------------|------|---------|
| **Main** | `cmd/api/main.go` | Bootstrap Fx application |
| **DI Wiring** | `internal/infra/fx/module.go` | All dependency injection |
| **Router** | `internal/transport/http/router.go` | HTTP routing + middleware |
| **Config** | `internal/infra/config/config.go` | Configuration loading |

---

## Layer Dependencies (Enforced by depguard)

```
┌─────────────────────────────────────────────────────────────────┐
│                         DOMAIN LAYER                            │
│   (stdlib only - no external dependencies)                      │
│   user.go, audit.go, id.go, pagination.go, errors.go            │
└───────────────────────────────┬─────────────────────────────────┘
                                │ implements
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER                          │
│   (domain only - no infra, transport, slog, otel, uuid, http)   │
│   user/create_user.go, user/get_user.go, audit/service.go       │
└───────────────────────────────┬─────────────────────────────────┘
                                │ orchestrates
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       TRANSPORT LAYER                           │
│   (domain + app - no infra packages)                            │
│   http/router.go, http/handler/*.go, http/middleware/*.go       │
└───────────────────────────────┬─────────────────────────────────┘
                                │ uses
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE LAYER                         │
│   (domain + external packages)                                  │
│   postgres/*.go, config/*.go, observability/*.go, fx/*.go       │
└─────────────────────────────────────────────────────────────────┘
```

---

## Integration Points

```
External Systems
       │
       ├──> HTTP Client ──> transport/http/handler/ ──> app/ ──> domain/
       │
       ├──> PostgreSQL <── infra/postgres/ <── domain interfaces
       │
       ├──> Prometheus <── :8081/metrics <── infra/observability/
       │
       └──> OTLP Collector <── middleware/tracing.go <── observability/tracer.go
```

---

## Test Infrastructure Analysis

### Unit Testing
- **Framework**: Go standard `testing` package
- **Mocks**: mockgen-generated mocks in `testutil/mocks/`
- **Assertions**: Custom assertions in `testutil/assert/`
- **Fixtures**: Standard test data in `testutil/fixtures/`

### Integration Testing
- **Container**: Testcontainers-go for PostgreSQL
- **Migration**: Automatic migration via `testutil/containers/migrate.go`
- **Cleanup**: Table truncation via `testutil/containers/truncate.go`
- **Transactions**: Per-test transactions via `testutil/containers/tx.go`

### Security Testing
- **IDOR Tests**: `integration_idor_test.go` - Insecure Direct Object Reference tests
- **Auth Tests**: Comprehensive JWT validation in `middleware/auth_test.go`

---

## Area yang Perlu Perhatian

### 1. Files/Packages Tanpa Tests
- `internal/shared/metrics/http_metrics.go` - Interface only, sebaiknya ada tests
- `internal/infra/postgres/sqlcgen/` - Generated code, expected tanpa tests

### 2. Large Test Files (mungkin perlu split)
- `internal/transport/http/middleware/auth_test.go` - 23KB
- `internal/transport/http/handler/user_test.go` - 22KB

### 3. Positive Findings
- Comprehensive test coverage (~50% files are tests)
- Dedicated test utilities package
- Testcontainers for realistic integration tests
- Security-focused tests (IDOR, auth)
- Benchmark tests for performance-critical code

---

*Dokumentasi ini dihasilkan oleh BMad Method - Document Project Workflow (Exhaustive Scan)*
