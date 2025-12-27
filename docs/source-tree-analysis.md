# Source Tree Analysis: golang-api-hexagonal

> **Deep Scan** - Annotated Directory Structure  
> **Tanggal:** 2025-12-27

---

## Struktur Lengkap dengan Anotasi

```
golang-api-hexagonal/
│
├── cmd/                                # 🚀 Entry Points
│   └── api/
│       └── main.go                     # Application bootstrap dengan Uber Fx
│                                       # - Loads config
│                                       # - Initializes DI container
│                                       # - Starts HTTP servers (public + internal)
│
├── internal/                           # 📦 Private Application Code
│   │
│   ├── domain/                         # 🔴 DOMAIN LAYER (Business Core)
│   │   │                               # ⚠️ RULE: Hanya boleh import stdlib!
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
│   ├── app/                            # 🟡 APPLICATION LAYER (Use Cases)
│   │   │                               # ⚠️ RULE: Hanya boleh import domain!
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
│   ├── transport/                      # 🔵 TRANSPORT LAYER (Inbound Adapters)
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
│   │       ├── middleware/             # HTTP middleware (21 files!)
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
│   │       │   ├── security.go         # Security headers
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
│   │       │   └── user_test.go
│   │       │
│   │       └── ctxutil/                # Context utilities
│   │           ├── claims.go           # JWT claims context
│   │           ├── claims_test.go
│   │           ├── trace.go            # Trace context
│   │           └── trace_test.go
│   │
│   ├── infra/                          # 🟢 INFRASTRUCTURE LAYER (Outbound Adapters)
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
│   │   │   ├── id_generator.go         # ID generation
│   │   │   ├── citext_integration_test.go  # CITEXT integration test
│   │   │   ├── test_helpers_test.go    # ⚠️ Empty file (22 bytes)
│   │   │   │
│   │   │   └── sqlcgen/                # ⚠️ Generated code (NO TESTS)
│   │   │       ├── db.go
│   │   │       ├── models.go
│   │   │       ├── querier.go
│   │   │       └── users.sql.go
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
│   └── shared/                         # 🟣 SHARED (Cross-cutting)
│       │
│       ├── metrics/                    # ⚠️ NO TESTS
│       │   └── http_metrics.go         # HTTPMetrics interface
│       │
│       └── redact/                     # PII redaction
│           ├── redactor.go
│           ├── redactor_test.go
│           ├── benchmark_test.go       # Performance tests
│           └── robustness_test.go      # Edge case tests
│
├── migrations/                         # 📊 Database Migrations (goose)
│   ├── 20251216000000_init.sql         # Initial schema
│   ├── 20251217000000_create_users.sql # Users table
│   ├── 20251219000000_create_audit_events.sql  # Audit events
│   └── 20251226084756_add_citext_email.sql     # CITEXT for email
│
├── queries/                            # 📝 sqlc Query Definitions
│   ├── users.sql                       # User queries
│   └── audit_events.sql                # Audit queries
│
├── docs/                               # 📚 Documentation
│   ├── index.md                        # Master index (this scan)
│   ├── architecture.md                 # Architecture docs
│   ├── openapi.yaml                    # OpenAPI 3.1 spec
│   ├── observability.md                # Observability guide
│   ├── local-development.md            # Dev setup guide
│   └── guides/                         # Additional guides
│
├── .github/workflows/                  # 🔄 CI/CD
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

## Critical Entry Points

| Entry Point | Path | Purpose |
|-------------|------|---------|
| **Main** | `cmd/api/main.go` | Bootstrap Fx application |
| **DI Wiring** | `internal/infra/fx/module.go` | All dependency injection |
| **Router** | `internal/transport/http/router.go` | HTTP routing + middleware |
| **Config** | `internal/infra/config/config.go` | Configuration loading |

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

## ⚠️ Area yang Perlu Perhatian

### 1. Files/Packages Tanpa Tests
- `internal/infra/postgres/sqlcgen/` - Generated code, expected
- `internal/shared/metrics/` - Interface only, tapi sebaiknya ada tests
- `internal/infra/postgres/test_helpers_test.go` - File kosong (22 bytes)

### 2. Large Test Files (mungkin perlu split)
- `internal/transport/http/middleware/auth_test.go` - 23KB
- `internal/transport/http/handler/user_test.go` - 22KB

### 3. Temporary/Review Files
- `internal/transport/internal_review_tmp/` - Folder review sementara

---

*Dokumentasi ini dihasilkan oleh BMad Method - Document Project Workflow*
