# Architecture Documentation

**Project:** golang-api-hexagonal  
**Architecture Pattern:** Hexagonal (Ports & Adapters)  
**Last Updated:** 2024-12-24  

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Layer Structure](#layer-structure)
3. [Dependency Rules](#dependency-rules)
4. [Component Details](#component-details)
5. [Data Flow](#data-flow)
6. [Configuration](#configuration)
7. [Observability](#observability)
8. [Security](#security)
9. [Testing Strategy](#testing-strategy)

---

## Architecture Overview

This project implements **Hexagonal Architecture** (also known as Ports & Adapters), which separates business logic from external concerns through well-defined interfaces.

```
┌──────────────────────────────────────────────────────────────┐
│                    TRANSPORT LAYER                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │  HTTP/Chi   │  │  Handlers   │  │     Middleware      │   │
│  │   Router    │  │  (User,     │  │  (Auth, Logging,    │   │
│  │             │  │   Health)   │  │   Metrics, Security)│   │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘   │
└─────────┼────────────────┼─────────────────────┼─────────────┘
          │                │                     │
          ▼                ▼                     ▼
┌──────────────────────────────────────────────────────────────┐
│                   APPLICATION LAYER                          │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                    Use Cases                             │ │
│  │  CreateUserUseCase, GetUserUseCase, ListUsersUseCase    │ │
│  │  AuditService                                            │ │
│  └─────────────────────────┬───────────────────────────────┘ │
│                            │ Uses interfaces from domain     │
└────────────────────────────┼─────────────────────────────────┘
                             ▼
┌──────────────────────────────────────────────────────────────┐
│                     DOMAIN LAYER                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │  Entities   │  │   Ports     │  │     Value Objects   │   │
│  │  (User,     │  │ (UserRepo,  │  │  (ID, Pagination,   │   │
│  │  AuditEvent)│  │  TxManager, │  │   Errors)           │   │
│  │             │  │  Querier)   │  │                     │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
                             ▲
                             │ Implements interfaces
┌────────────────────────────┼─────────────────────────────────┐
│                   INFRASTRUCTURE LAYER                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │  Postgres   │  │   Config    │  │    Observability    │   │
│  │  (Pool,     │  │  (envconfig)│  │  (Logger, Tracer,   │   │
│  │  Repos,     │  │             │  │   Metrics)          │   │
│  │  TxManager) │  │             │  │                     │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
```

---

## Layer Structure

### 1. Domain Layer (`internal/domain/`)

The **core business logic** with zero external dependencies.

| File | Purpose |
|------|---------|
| `user.go` | User entity and UserRepository interface |
| `audit.go` | AuditEvent entity and AuditEventRepository interface |
| `querier.go` | Database abstraction interface |
| `tx.go` | Transaction manager interface (TxManager) |
| `id.go` | UUID-based ID value object |
| `pagination.go` | Pagination parameters value object |
| `errors.go` | Domain-specific error types |
| `redactor.go` | PII redaction configuration |

**Key Interfaces (Ports):**

```go
// Querier - Database abstraction
type Querier interface {
    Exec(ctx context.Context, sql string, args ...any) (any, error)
    Query(ctx context.Context, sql string, args ...any) (any, error)
    QueryRow(ctx context.Context, sql string, args ...any) any
}

// TxManager - Transaction management
type TxManager interface {
    WithTx(ctx context.Context, fn func(tx Querier) error) error
}

// UserRepository - User persistence
type UserRepository interface {
    Create(ctx context.Context, q Querier, user *User) error
    GetByID(ctx context.Context, q Querier, id ID) (*User, error)
    List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)
}
```

### 2. Application Layer (`internal/app/`)

**Orchestrates business operations** using domain entities and interfaces.

| Directory | Purpose |
|-----------|---------|
| `app/user/` | User use cases (Create, Get, List) |
| `app/audit/` | Audit service for recording events |
| `app/auth.go` | Authentication context helpers |
| `app/errors.go` | Application-level errors |

**Use Case Pattern:**

```go
type CreateUserUseCase struct {
    userRepo     domain.UserRepository
    auditService AuditService
    idGen        domain.IDGenerator
    txManager    domain.TxManager
    querier      domain.Querier
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
    // 1. Validate request
    // 2. Create user entity
    // 3. Execute in transaction:
    //    - Save user to repository
    //    - Record audit event
    // 4. Return response
}
```

### 3. Transport Layer (`internal/transport/http/`)

**HTTP delivery mechanism** - handles requests and responses.

| Directory | Purpose |
|-----------|---------|
| `handler/` | HTTP handlers (Health, Ready, User) |
| `middleware/` | HTTP middleware (Auth, Logging, Metrics, Security) |
| `contract/` | Request/Response DTOs, RFC 7807 errors |
| `ctxutil/` | Context utility functions (see [ctxutil Pattern](#ctxutil-pattern)) |
| `router.go` | Chi router configuration |

#### ctxutil Pattern

The `ctxutil` package is a **leaf package** designed to hold context key/value utilities. It was created to prevent import cycles between middleware and infrastructure packages.

**Problem Solved:**
```
middleware/requestid.go → observability/logger.go → middleware/requestid.go
                       ↑_____IMPORT CYCLE_____↑
```

**Solution:**
```
middleware/requestid.go → ctxutil/request_id.go ← observability/logger.go
                       ↑_____NO CYCLE_____↑
```

**Package Contents:**

| File | Functions | Purpose |
|------|-----------|---------|
| `request_id.go` | `GetRequestID`, `SetRequestID` | Request ID context management |
| `trace.go` | `GetTraceID`, `SetTraceID`, `GetSpanID`, `SetSpanID` | Trace/span ID context management |

**Usage Pattern:**

```go
// In middleware (setter)
import "github.com/.../ctxutil"

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqID := generateID()
        ctx := ctxutil.SetRequestID(r.Context(), reqID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// In any package (getter)
import "github.com/.../ctxutil"

func SomeFunction(ctx context.Context) {
    reqID := ctxutil.GetRequestID(ctx)
    if reqID != "" {
        // Use request ID
    }
}
```

**Constants for Zero Values:**
```go
// Used to filter out zero/empty trace IDs
ctxutil.EmptyTraceID // "00000000000000000000000000000000"
ctxutil.EmptySpanID  // "0000000000000000"
```

**When to Use ctxutil:**
- ✅ Adding new context-propagated values (correlation IDs, tenant IDs, etc.)
- ✅ Breaking import cycles between middleware and other packages
- ✅ When multiple packages need to read/write the same context value
- ❌ For package-specific context values that don't need external access

**Middleware Stack Order:**
1. SecureHeaders - Security headers (FIRST)
2. RequestID - Generate/passthrough request ID
3. Tracing - OpenTelemetry spans (if enabled)
4. Metrics - Prometheus recording
5. RequestLogger - Structured logging
6. RealIP - Extract client IP
7. BodyLimiter - Request size limit
8. Recoverer - Panic recovery

### 4. Infrastructure Layer (`internal/infra/`)

**External service adapters** - implements domain interfaces.

| Directory | Purpose |
|-----------|---------|
| `config/` | Environment configuration loading |
| `postgres/` | Database pool, repositories, transactions |
| `observability/` | Logger, Tracer, Metrics factories |

---

## Dependency Rules

The following **unidirectional dependency rules** are enforced:

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| **Domain** | stdlib only | app, transport, infra |
| **App** | domain | transport, infra |
| **Transport** | domain, app | infra |
| **Infra** | domain, external packages | app, transport |

These rules are enforced by:
- **golangci-lint depguard** configuration
- **CI pipeline** validation

---

## Component Details

### Transaction Flow

```
Handler → UseCase → WithTx(func(tx Querier) error {
    repo.Create(ctx, tx, entity)
    auditService.Record(ctx, tx, event)
})
```

- All write operations are wrapped in transactions
- `TxManager.WithTx()` handles commit/rollback
- Repositories accept `Querier` interface (pool or transaction)

### Audit Trail Integration

```go
// In use case after successful operation:
auditInput := audit.AuditEventInput{
    EventType:  domain.EventUserCreated,
    ActorID:    req.ActorID,
    EntityType: "user",
    EntityID:   user.ID,
    Payload:    user,  // Auto-redacted
    RequestID:  req.RequestID,
}
auditService.Record(ctx, tx, auditInput)
```

---

## Configuration

All configuration via **environment variables** (no .env auto-loading):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | ✅ | - | PostgreSQL connection string |
| `PORT` | ❌ | `8080` | HTTP server port |
| `LOG_LEVEL` | ❌ | `info` | debug, info, warn, error |
| `ENV` | ❌ | `development` | Environment identifier |
| `OTEL_ENABLED` | ❌ | `false` | Enable OpenTelemetry tracing |
| `JWT_ENABLED` | ❌ | `false` | Enable JWT authentication |
| `JWT_SECRET` | ❌* | - | JWT signing secret (required if JWT_ENABLED) |
| `RATE_LIMIT_RPS` | ❌ | `100` | Requests per second limit |

See `.env.example` for complete list.

---

## Observability

### Logging
- **Structured JSON** via `log/slog`
- Attributes: `service`, `env`, `requestId`, `traceId`
- Log levels: debug, info, warn, error

### Tracing
- **OpenTelemetry** with OTLP exporter
- Spans for HTTP requests and database operations
- Trace context propagation via headers

### Metrics
- **Prometheus** format at `/metrics`
- `http_requests_total{method, route, status}`
- `http_request_duration_seconds{method, route}`
- Go runtime metrics (goroutines, memory)

---

## Security

### HTTP Security Headers
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'none'`
- `Strict-Transport-Security: max-age=31536000` (HTTPS only)

### Authentication
- **JWT (HS256)** middleware (optional)
- Token validation with expiry check
- Claims extraction to context

### Rate Limiting
- IP-based limiting (configurable RPS)
- Proxy-aware with `TRUST_PROXY` setting
- Per-user limiting when JWT enabled

### PII Protection
- Automatic redaction in audit payloads
- Configurable email redaction modes (`full`/`partial`)

---

## Testing Strategy

### Test Levels

| Layer | Test Type | Coverage Target |
|-------|-----------|-----------------|
| Domain | Unit tests | ≥80% |
| App | Unit tests (mocked deps) | ≥80% |
| Transport | Handler tests | - |
| Infra | Integration tests | - |

### Running Tests

```bash
# Run all tests with race detector
make test

# Check coverage threshold (domain+app ≥80%)
make coverage

# Run linter
make lint

# Full CI pipeline locally
make ci
```

---

## Extension Points

### Adding a New Module

1. **Domain**: Add entity and repository interface in `internal/domain/`
2. **App**: Create use cases in `internal/app/{module}/`
3. **Infra**: Implement repository in `internal/infra/postgres/`
4. **Transport**: Add handler in `internal/transport/http/handler/`
5. **Router**: Register routes in `router.go`
6. **Main**: Wire dependencies in `cmd/api/main.go`

See [Adding Module Guide](./guides/adding-module.md) for detailed steps.

### Adding a New Adapter

See [Adding Adapter Guide](./guides/adding-adapter.md) for:
- Database adapters (Redis, MongoDB)
- External service integrations
- Message queue consumers
