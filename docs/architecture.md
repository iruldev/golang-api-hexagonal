# Arsitektur: golang-api-hexagonal

> **Dokumentasi Brownfield** - Exhaustive Scan Analysis
> **Tanggal:** 2025-12-29
> **Total Files:** 140 Go files (70 production + 70 tests)

---

## 1. Executive Summary

**golang-api-hexagonal** adalah production-ready Go API yang mengimplementasikan **Hexagonal Architecture (Ports & Adapters)** dengan:

- ✅ Layer boundaries yang di-enforce via golangci-lint depguard
- ✅ Domain layer yang murni (hanya stdlib)
- ✅ Dependency injection dengan Uber Fx
- ✅ Comprehensive observability (OTEL + Prometheus + slog)
- ✅ Security middleware (JWT, rate limiting, body limiter)
- ✅ 70 test files dengan 80% coverage threshold
- ✅ Testcontainers untuk integration testing
- ✅ Security-focused tests (IDOR, auth edge cases)

---

## 2. Arsitektur Pattern

### 2.1 Hexagonal Architecture Overview

```
                    ┌─────────────────────────────────────┐
                    │         External World              │
                    │  (HTTP Clients, Databases, etc.)    │
                    └──────────────┬──────────────────────┘
                                   │
           ┌───────────────────────┼───────────────────────┐
           │                       │                       │
           ▼                       ▼                       ▼
    ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
    │  HTTP Port  │         │  Metrics    │         │  Database   │
    │  (Chi)      │         │  (Prometheus)│        │  (PostgreSQL)│
    └──────┬──────┘         └──────┬──────┘         └──────┬──────┘
           │                       │                       │
           │     ┌─────────────────┼─────────────────┐     │
           │     │                                   │     │
           ▼     ▼                                   ▼     ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                        transport/http                       │
    │                    (Inbound Adapters)                       │
    │    handler/ │ middleware/ │ contract/ │ ctxutil/            │
    └───────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                           app/                              │
    │                    (Application Layer)                      │
    │              user/ │ audit/ │ auth.go                       │
    └───────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                         domain/                             │
    │                    (Business Core)                          │
    │     User │ Audit │ ID │ Pagination │ Querier │ TxManager    │
    └─────────────────────────────────────────────────────────────┘
                                │
                                ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                          infra/                             │
    │                   (Outbound Adapters)                       │
    │        postgres/ │ config/ │ observability/ │ fx/           │
    └─────────────────────────────────────────────────────────────┘
```

### 2.2 Layer Rules (Enforced)

| Layer | Dapat Import | Tidak Boleh Import |
|-------|--------------|-------------------|
| **domain/** | stdlib only | app, transport, infra |
| **app/** | domain | transport, infra, slog, http |
| **transport/** | domain, app, shared | infra |
| **infra/** | domain, shared, external packages | app, transport |
| **infra/fx/** | ALL (wiring layer) | - |

> **Enforcement:** `.golangci.yml` dengan depguard rules - CI akan gagal jika violated

---

## 3. Struktur Direktori

```
golang-api-hexagonal/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point (Fx bootstrap)
│
├── internal/
│   ├── domain/                  # 🔴 Business Core (stdlib only)
│   │   ├── user.go              # User entity + UserRepository interface
│   │   ├── audit.go             # AuditEvent entity + repository interface
│   │   ├── id.go                # ID value object + IDGenerator interface
│   │   ├── pagination.go        # ListParams value object
│   │   ├── querier.go           # Querier interface (DB abstraction)
│   │   ├── tx.go                # TxManager interface
│   │   ├── redactor.go          # Redactor interface (PII masking)
│   │   └── errors.go            # Domain errors
│   │
│   ├── app/                     # 🟡 Application Layer (use cases)
│   │   ├── user/
│   │   │   ├── create_user.go   # CreateUserUseCase
│   │   │   ├── get_user.go      # GetUserUseCase
│   │   │   └── list_users.go    # ListUsersUseCase
│   │   ├── audit/
│   │   │   └── service.go       # AuditService
│   │   ├── auth.go              # AuthParser interface
│   │   └── errors.go            # Application errors (AppError)
│   │
│   ├── transport/http/          # 🔵 Inbound Adapters (HTTP)
│   │   ├── router.go            # Chi router setup
│   │   ├── handler/             # HTTP handlers
│   │   │   ├── health.go
│   │   │   ├── ready.go
│   │   │   └── user.go
│   │   ├── middleware/          # HTTP middleware stack
│   │   │   ├── auth.go          # JWT authentication
│   │   │   ├── logging.go       # Request logging
│   │   │   ├── metrics.go       # Prometheus metrics
│   │   │   ├── tracing.go       # OpenTelemetry tracing
│   │   │   ├── security.go      # Security headers
│   │   │   ├── ratelimit.go     # Rate limiting
│   │   │   └── body_limiter.go  # Request size limiter
│   │   ├── contract/            # Request/Response DTOs
│   │   └── ctxutil/             # Context utilities
│   │
│   ├── infra/                   # 🟢 Outbound Adapters
│   │   ├── postgres/            # Database implementation
│   │   │   ├── pool.go          # Connection pool
│   │   │   ├── resilient_pool.go # Auto-reconnecting pool
│   │   │   ├── querier.go       # Querier implementation
│   │   │   ├── tx_manager.go    # Transaction manager
│   │   │   ├── user_repo.go     # UserRepository implementation
│   │   │   ├── audit_event_repo.go # AuditEventRepository implementation
│   │   │   └── sqlcgen/         # sqlc generated code
│   │   ├── config/              # Configuration loading
│   │   ├── observability/       # Logging, metrics, tracing setup
│   │   └── fx/                  # Uber Fx DI modules
│   │       └── module.go        # Dependency injection wiring
│   │
│   └── shared/                  # Cross-cutting concerns
│       ├── metrics/             # Shared metrics interfaces
│       ├── redact/              # PII redaction implementation
│       └── ctxutil/             # Shared context utilities
│
├── migrations/                  # Database migrations (goose)
│   ├── 20251216000000_init.sql
│   ├── 20251217000000_create_users.sql
│   ├── 20251219000000_create_audit_events.sql
│   └── 20251226084756_add_citext_email.sql
│
├── queries/                     # sqlc query definitions
│   ├── users.sql
│   └── audit_events.sql
│
└── docs/                        # Documentation
```

---

## 4. Dependency Injection (Uber Fx)

### 4.1 Module Structure

```go
// internal/infra/fx/module.go
var Module = fx.Options(
    ConfigModule,        // Configuration loading
    ObservabilityModule, // Logger, Metrics, Tracer
    PostgresModule,      // DB Pool, Querier, TxManager
    DomainModule,        // Repositories, IDGenerator, Redactor
    AppModule,           // Use Cases
    TransportModule,     // Handlers, Routers
)
```

### 4.2 Dependency Graph

```
config.Config
    │
    ├──> observability.Logger
    ├──> observability.Metrics
    ├──> observability.Tracer
    │
    ├──> postgres.ResilientPool
    │       │
    │       ├──> postgres.Querier ──> domain.Querier
    │       └──> postgres.TxManager ──> domain.TxManager
    │
    ├──> postgres.UserRepo ──> domain.UserRepository
    ├──> postgres.AuditEventRepo ──> domain.AuditEventRepository
    ├──> postgres.IDGenerator ──> domain.IDGenerator
    │
    ├──> audit.AuditService
    │       │
    │       └──> user.CreateUserUseCase
    │            user.GetUserUseCase
    │            user.ListUsersUseCase
    │
    └──> handler.UserHandler
         handler.HealthHandler
         handler.ReadyHandler
              │
              └──> http.Router (Chi)
```

---

## 5. Komponen Utama

### 5.1 Domain Layer

| Interface | Deskripsi | Implementasi |
|-----------|-----------|--------------|
| `UserRepository` | User CRUD operations | `postgres.UserRepo` |
| `AuditEventRepository` | Audit event storage | `postgres.AuditEventRepo` |
| `Querier` | DB query abstraction | `postgres.PoolQuerier` |
| `TxManager` | Transaction management | `postgres.TxManager` |
| `IDGenerator` | UUID generation | `postgres.IDGenerator` |
| `Redactor` | PII masking | `redact.PIIRedactor` |

### 5.2 Application Layer

| Use Case | Deskripsi |
|----------|-----------|
| `CreateUserUseCase` | Create user with audit trail |
| `GetUserUseCase` | Get user by ID |
| `ListUsersUseCase` | List users with pagination |
| `AuditService` | Record audit events |

### 5.3 Transport Layer

| Handler | Routes |
|---------|--------|
| `HealthHandler` | `GET /health` |
| `ReadyHandler` | `GET /ready` |
| `UserHandler` | `GET/POST /api/v1/users`, `GET /api/v1/users/{id}` |

### 5.4 Middleware Stack

```go
// Order matters - applied top to bottom
1. RequestID          // Generate X-Request-ID
2. Logging            // Request/response logging
3. Tracing            // OpenTelemetry spans
4. Metrics            // Prometheus metrics
5. Security           // Security headers
6. BodyLimiter        // Request size limit
7. RateLimiter        // Rate limiting
8. ResponseWrapper    // Standard JSON envelope
9. Auth (JWT)         // Protected routes only
```

---

## 6. Kepatuhan Standar Internasional

### 6.1 ✅ Yang Sudah Baik

| Aspek | Status | Detail |
|-------|--------|--------|
| **Architecture** | ✅ | Clean hexagonal dengan layer enforcement |
| **Dependency Injection** | ✅ | Uber Fx dengan proper wiring |
| **Testing** | ✅ | 49 test files, 80% coverage threshold |
| **CI/CD** | ✅ | 8 quality gates termasuk security scan |
| **API Design** | ✅ | RESTful dengan OpenAPI spec |
| **Observability** | ✅ | OTEL tracing, Prometheus metrics, structured logging |
| **Security** | ✅ | JWT auth, rate limiting, security headers |
| **Database** | ✅ | sqlc type-safe queries, migrations |
| **Linting** | ✅ | golangci-lint dengan depguard |

### 6.2 ⚠️ Area untuk Improvement (International Grade Gap Analysis)

| Area | Priority | Issue | Recommendation |
|------|----------|-------|----------------|
| **API Documentation** | HIGH | OpenAPI spec belum complete (missing examples, descriptions) | Enhance OpenAPI with examples, error codes, proper descriptions |
| **Error Standardization** | HIGH | Domain errors bisa lebih granular | Implement RFC 7807 fully dengan error codes taxonomy |
| **Missing Tests** | MEDIUM | `shared/metrics/` tanpa tests (interface only) | Add interface compliance tests |
| **Large Test Files** | MEDIUM | `auth_test.go` (23KB), `user_test.go` (22KB) | Split into smaller, focused test files |
| **Retry/Circuit Breaker** | MEDIUM | ResilientPool exists, tapi no application-level resilience | Add retry patterns untuk external calls |
| **API Rate Limit Headers** | LOW | Rate limiting ada, tapi no X-RateLimit-* headers | Add standard rate limit response headers |
| **Versioning Strategy** | LOW | Hanya v1, no deprecation strategy | Document API lifecycle dan deprecation policy |
| **Health Check Details** | LOW | /health basic | Add /health dengan dependency status details |

### 6.3 ✅ Recently Resolved (Found in Exhaustive Scan)

| Area | Status | Detail |
|------|--------|--------|
| **Integration Tests** | ✅ RESOLVED | Testcontainers implemented di `testutil/containers/` |
| **Test Fixtures** | ✅ RESOLVED | Standard fixtures di `testutil/fixtures/` |
| **Mock Generation** | ✅ RESOLVED | Centralized mocks di `testutil/mocks/` |
| **Security Tests** | ✅ RESOLVED | IDOR tests, comprehensive auth edge cases |
| **Benchmark Tests** | ✅ RESOLVED | Performance tests di `shared/redact/` |

---

## 7. Security Architecture

### 7.1 Authentication Flow

```
Client Request
      │
      ▼
┌─────────────────┐     JWT Token?     ┌──────────────────┐
│  /api/v1/*      │────────────────────│  Auth Middleware │
│  Protected      │                    │  JWT Validation  │
└─────────────────┘                    └────────┬─────────┘
                                                │
                                    ┌───────────┴───────────┐
                                    │                       │
                                    ▼                       ▼
                              Valid Token            Invalid Token
                                    │                       │
                                    ▼                       ▼
                              Set Claims           401 Unauthorized
                              in Context
```

### 7.2 Security Features

- **JWT Authentication**: HS256 signing, configurable via env
- **Rate Limiting**: httprate with configurable RPS
- **Body Limiter**: Prevent large payload attacks (default 1MB)
- **Security Headers**: X-Content-Type-Options, X-Frame-Options, etc.
- **Secret Scanning**: gitleaks in CI pipeline
- **Vulnerability Scanning**: govulncheck in CI

---

## 8. Database Architecture

### 8.1 Connection Management

```
┌────────────────────────────────────────────────────────────────┐
│                     ResilientPool                              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              pgxpool.Pool                               │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │   │
│  │  │  Conn   │ │  Conn   │ │  Conn   │ │  Conn   │ ...   │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                │
│  Features:                                                     │
│  - Auto-reconnection on failure                                │
│  - Configurable pool size (min/max conns)                      │
│  - Connection lifetime management                              │
│  - Health check for readiness probe                            │
└────────────────────────────────────────────────────────────────┘
```

### 8.2 Tables

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `users` | User data | id (UUID), email (CITEXT), first_name, last_name |
| `audit_events` | Audit trail | id, event_type, actor_id, entity_type, entity_id, payload |
| `goose_db_version` | Migration tracking | version_id, is_applied |

---

## 9. Observability Stack

```
┌─────────────────────────────────────────────────────────────┐
│                    Application                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   slog      │  │ Prometheus  │  │   OpenTelemetry     │  │
│  │   Logger    │  │   Metrics   │  │   Tracing           │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
└─────────┼────────────────┼────────────────────┼─────────────┘
          │                │                    │
          ▼                ▼                    ▼
     JSON stdout      :8081/metrics        OTLP gRPC
          │                │                    │
          ▼                ▼                    ▼
       Logging          Prometheus           Jaeger/Tempo
       Platform         + Grafana            + Grafana
```

---

## 10. International Grade Improvement Roadmap

### 10.1 Priority 1: API Excellence
- [ ] Enhance OpenAPI spec dengan examples, descriptions, dan error responses
- [ ] Implement complete RFC 7807 problem details dengan error codes taxonomy
- [ ] Add X-RateLimit-Remaining, X-RateLimit-Limit, X-RateLimit-Reset headers
- [ ] Document API versioning dan deprecation policy

### 10.2 Priority 2: Resilience Patterns
- [ ] Add application-level retry dengan exponential backoff
- [ ] Implement circuit breaker pattern untuk external dependencies
- [ ] Add timeout configuration per-endpoint
- [ ] Enhance health check dengan dependency status details

### 10.3 Priority 3: Code Quality
- [ ] Split large test files (auth_test.go, user_test.go) into focused modules
- [ ] Add tests untuk `shared/metrics/` interface
- [ ] Add contract tests untuk API backward compatibility
- [ ] Implement mutation testing untuk test quality assessment

### 10.4 Existing Best Practices to Maintain
1. **Architecture Constraints**: Patuhi layer rules di `.golangci.yml`
2. **Testing Standards**: Min 80% coverage untuk domain+app
3. **API Patterns**: Ikuti existing contract patterns di `internal/transport/http/contract/`
4. **DI Wiring**: Extend `internal/infra/fx/module.go` untuk dependencies baru
5. **Security**: Maintain JWT validation, rate limiting, security headers

---

## 11. Statistics Summary

| Metric | Count |
|--------|-------|
| Total Go Files | 140 |
| Production Files | ~70 |
| Test Files | ~70 |
| Domain Layer | 12 files |
| Application Layer | 11 files |
| Transport Layer | 36 files |
| Infrastructure Layer | 24 files |
| Test Utilities | 13 files |
| API Endpoints | 6 |
| Database Tables | 2 |
| Migrations | 4 |

---

*Dokumentasi ini dihasilkan oleh BMad Method - Document Project Workflow (Exhaustive Scan)*
