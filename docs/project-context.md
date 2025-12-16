---
project_name: 'golang-api-hexagonal'
user_name: 'Chat'
date: '2025-12-16'
sections_completed: ['technology_stack', 'layer_rules', 'naming_conventions', 'error_handling', 'api_formats', 'logging', 'testing', 'anti_patterns']
status: 'complete'
rule_count: 45
optimized_for_llm: true
---

# Project Context for AI Agents

_Critical rules for implementing code in golang-api-hexagonal. Focus on what's unobvious._

---

## Technology Stack (Exact Versions)

| Component | Package | Version |
|-----------|---------|---------|
| Language | Go | 1.23+ (CI: 1.22, 1.23) |
| HTTP Router | github.com/go-chi/chi/v5 | v5.x |
| PostgreSQL | github.com/jackc/pgx/v5 | v5.x |
| Config | github.com/kelseyhightower/envconfig | latest |
| Validation | github.com/go-playground/validator/v10 | v10.x |
| JWT | github.com/golang-jwt/jwt/v5 | v5.x |
| Rate Limit | github.com/go-chi/httprate | latest |
| Migrations | github.com/pressly/goose/v3 | v3.x |
| UUID | github.com/google/uuid | latest |
| Tracing | go.opentelemetry.io/otel | latest |
| Testing | github.com/stretchr/testify | latest |

---

## Critical Layer Rules

### Domain Layer (`internal/domain/`)

```
✅ ALLOWED: stdlib only (errors, context, time, strings, fmt)
❌ FORBIDDEN: slog, uuid, pgx, chi, otel, ANY external package
```

**Domain Purity:**
- Use `type ID string` for identifiers — UUID parsing/generation happens at boundaries
- Define sentinel errors: `var ErrUserNotFound = errors.New("user not found")`
- Interfaces define contracts: `UserRepository`, `Querier`, `TxManager`
- NO logging in domain — ever

### App Layer (`internal/app/`)

```
✅ ALLOWED: domain imports only
❌ FORBIDDEN: net/http, pgx, slog, otel, uuid, transport, infra
```

**Use Case Rules:**
- Authorization checks happen HERE (not middleware)
- Use `TxManager.WithTx()` for multi-step operations
- Convert domain errors to typed `AppError` with `Code`
- NO logging — tracing context only

### Transport Layer (`internal/transport/http/`)

```
✅ ALLOWED: domain, app, chi, uuid, stdlib
❌ FORBIDDEN: pgx, direct infra/observability imports
```

**Handler Rules:**
- Receive `*slog.Logger` via injection (don't import infra/observability)
- Generate UUID v7 for new entities, parse UUID from path params
- Pass `domain.ID` (string) to use case
- Map `AppError.Code` → HTTP status + RFC 7807
- ONLY place that knows HTTP status codes

### Infra Layer (`internal/infra/`)

```
✅ ALLOWED: domain, pgx, slog, otel, uuid, external packages
❌ FORBIDDEN: app, transport
```

**Repository Rules:**
- Implement domain interfaces
- Accept `Querier` parameter (works with pool or tx)
- Wrap errors with `op` string: `fmt.Errorf("%s: %w", op, err)`
- Convert `domain.ID` ↔ `uuid.UUID` at repository boundary

---

## UUID v7 Handling

**Key Rule:** UUID v7 is generated at transport/infra boundary, NOT in database.

```go
// Transport layer: generate UUID v7 for new entities
import "github.com/google/uuid"

func (h *UserHandler) CreateUser(...) {
    id, _ := uuid.NewV7()  // Generate at boundary
    req.ID = domain.ID(id.String())
    // Pass to use case...
}

// Infra layer: parse UUID from domain.ID
func (r *UserRepo) Create(ctx context.Context, q Querier, user *domain.User) error {
    id, err := uuid.Parse(string(user.ID))
    if err != nil {
        return fmt.Errorf("userRepo.Create: invalid ID: %w", err)
    }
    // Use id (uuid.UUID) in SQL...
}
```

**Database schema:** No UUID default — ID provided by application.

```sql
CREATE TABLE users (
    id uuid PRIMARY KEY,  -- No DEFAULT, app provides UUID v7
    email VARCHAR(255) NOT NULL,
    ...
);
```

---

## Naming Conventions (Strict)

### Database (PostgreSQL)
```sql
-- Tables: snake_case, plural
CREATE TABLE users (...);
CREATE TABLE audit_events (...);

-- Columns: snake_case
created_at, updated_at, first_name

-- Foreign Keys: {referenced_table_singular}_id
user_id, audit_event_id, created_by_user_id

-- Indexes
idx_users_email        -- regular
uniq_users_email       -- unique

-- Primary Key: id uuid (NO default, app generates UUID v7)
id uuid PRIMARY KEY
```

### Go Code
```go
// Packages: lowercase, 1-2 words, NO underscores
package user      // ✅
package auditlog  // ✅
package audit_log // ❌

// Exports: PascalCase
type User struct {}
func CreateUser() {}

// Private: camelCase
func validate() {}
var defaultSize = 20

// Constants: PascalCase (NOT ALL_CAPS)
const DefaultPageSize = 20  // ✅
const DEFAULT_PAGE_SIZE = 20 // ❌ (except env vars)

// Domain Errors: Err prefix
var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already exists")
```

### API
```
Paths:       kebab-case, plural     /api/v1/users, /api/v1/audit-events
Params:      Chi syntax             /api/v1/users/{id}
Query:       camelCase              ?page=1&pageSize=20
JSON fields: camelCase              { "firstName": "John", "createdAt": "..." }
Timestamps:  RFC 3339 UTC           "2025-12-16T10:30:00Z"
IDs:         UUID v7 string         "019400a0-1234-7abc-..."
```

---

## Error Handling (Layered Strategy)

```
Domain  →  var ErrUserNotFound = errors.New("user not found")
           (sentinel errors only)

Infra   →  fmt.Errorf("userRepo.GetByID: %w", domain.ErrUserNotFound)
           (wrap with op string)

App     →  &AppError{Op: "GetUser", Code: "USER_NOT_FOUND", Err: err}
           (typed error with machine-readable Code)

Handler →  HTTP 404 + RFC 7807 JSON
           (ONLY place that maps to HTTP)
```

**AppError Structure:**
```go
type AppError struct {
    Op      string // "CreateUser"
    Code    string // "USER_NOT_FOUND", "VALIDATION_ERROR"
    Message string // Human-readable
    Err     error  // Wrapped error
}
```

---

## API Response Formats

**Success (single):**
```json
{ "data": { "id": "...", "email": "...", "createdAt": "2025-12-16T10:30:00Z" } }
```

**Success (list):**
```json
{
  "data": [...],
  "pagination": { "page": 1, "pageSize": 20, "totalItems": 150, "totalPages": 8 }
}
```

**Error (RFC 7807):**
```json
{
  "type": "https://api.example.com/problems/validation-error",
  "title": "Validation Error",
  "status": 400,
  "detail": "One or more fields failed validation",
  "instance": "/api/v1/users",
  "code": "VALIDATION_ERROR",
  "validationErrors": [{ "field": "email", "message": "must be valid email" }]
}
```

---

## Logging Rules

**Required fields (ALL log entries):**
```go
slog.Info("user created",
    "service", cfg.ServiceName,
    "env", cfg.Environment,
    "requestId", middleware.GetRequestID(ctx),
    "traceId", middleware.GetTraceID(ctx),  // empty if tracing off
    "userId", user.ID,
)
```

**Audit Events ≠ Logs:**
- Audit → DB sink with full payload (PII redacted)
- Log → metadata only (event type, entity ID)

**NEVER log:** passwords, tokens, PII, full request bodies

---

## Testing Patterns

| Type | Location | Build Tag | Command |
|------|----------|-----------|---------|
| Unit (domain/app) | `*_test.go` co-located | none | `go test ./internal/domain/... ./internal/app/...` |
| Integration (repo) | `*_test.go` co-located | `integration` | `go test -tags=integration ./internal/infra/...` |

**Coverage requirement:** domain + app ≥ 80%

**Test style:** Table-driven with testify assertions
```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserRequest
        wantErr bool
    }{
        {"valid user", CreateUserRequest{...}, false},
        {"missing email", CreateUserRequest{...}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

---

## Anti-Patterns (NEVER DO)

### Layer Violations
```go
// ❌ Domain importing external package
package domain
import "github.com/google/uuid"  // WRONG

// ❌ App layer logging
package user
import "log/slog"  // WRONG

// ❌ Handler knowing database
package handler
import "github.com/jackc/pgx/v5"  // WRONG

// ❌ Domain returning HTTP status
func GetUser() (*User, error) {
    return nil, &DomainError{Status: 404}  // WRONG
}
```

### UUID Handling
```go
// ❌ Database generates UUID
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid()  // WRONG - use UUID v7 from app
);

// ❌ Domain using uuid package
package domain
import "github.com/google/uuid"
type User struct {
    ID uuid.UUID  // WRONG - use type ID string
}
```

### Error Handling
```go
// ❌ Swallowing errors
if err != nil {
    return nil  // WRONG - always handle or propagate
}

// ❌ Logging and returning
if err != nil {
    slog.Error("failed", "err", err)
    return err  // WRONG - double logging upstream
}

// ❌ HTTP status in app layer
return &AppError{Status: 404}  // WRONG - use Code, not Status
```

### Configuration
```go
// ❌ Hardcoded values
const DatabaseURL = "postgres://..."  // WRONG

// ❌ Config files
viper.ReadConfig("config.yaml")  // WRONG - use env vars only
```

---

## Quick Reference

### File Locations
```
cmd/api/main.go              → Entry point, DI wiring
internal/domain/*.go         → Entities, interfaces, errors (type ID string)
internal/app/{module}/*.go   → Use cases
internal/transport/http/     → Handlers, middleware, router (UUID generation)
internal/infra/postgres/     → Repository implementations (UUID parsing)
internal/infra/config/       → Config struct
internal/shared/             → Internal utilities
migrations/                  → goose SQL files
```

### Makefile Commands
```bash
make setup      # Install tools
make run        # Run locally
make test       # Run unit tests
make lint       # Run golangci-lint
make migrate    # Run migrations
```

### depguard Enforcement
Boundaries enforced in CI via `.golangci.yml` depguard rules.
Violations = build failure.

---

## Usage Guidelines

**For AI Agents:**
- Read this file before implementing any code
- Follow ALL rules exactly as documented
- When in doubt, prefer the more restrictive option
- Layer boundaries are enforced by CI — violations fail the build

**For Humans:**
- Keep this file lean and focused on agent needs
- Update when technology stack changes
- Review quarterly for outdated rules
- Remove rules that become obvious over time

**Last Updated:** 2025-12-16

---
