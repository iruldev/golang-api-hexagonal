# Technical Research: Production-Grade Go Backend Boilerplate

**Project:** golang-api-hexagonal  
**Research Date:** 2024-12-24  
**Research Type:** Technical  
**Status:** Complete  

---

## Executive Summary

This research provides comprehensive analysis and recommendations for upgrading the `golang-api-hexagonal` project to an international production-grade backend boilerplate. The research addresses four priority areas: Security Hardening, Secret Management, Dependency Injection (Google Wire), and sqlc Adoption.

**Key Findings:**
1. **Security gaps** in current implementation require immediate attention (P0/P1 fixes)
2. **Google Wire** is the optimal DI choice for this boilerplate (compile-time, zero runtime overhead)
3. **Vault Agent injection** with `*_FILE` pattern provides vendor-agnostic secret management
4. **sqlc** can be adopted incrementally without disrupting existing transaction manager

---

## Table of Contents

1. [Security Hardening Best Practices](#1-security-hardening-best-practices)
2. [Secret Management Patterns](#2-secret-management-patterns)
3. [Dependency Injection with Google Wire](#3-dependency-injection-with-google-wire)
4. [sqlc Adoption Strategy](#4-sqlc-adoption-strategy)
5. [Implementation Roadmap](#5-implementation-roadmap)
6. [Conclusions & Recommendations](#6-conclusions--recommendations)

---

## 1. Security Hardening Best Practices

### 1.1 JWT Validation (golang-jwt/jwt/v5)

**Current Gap:** JWT middleware exists but may lack comprehensive claim validation.

#### Production Requirements

| Claim | Validation | Implementation |
|-------|------------|----------------|
| `exp` | Automatic | Library handles; configure leeway |
| `iss` | Required | `jwt.WithIssuer(expectedIssuer)` |
| `aud` | Required | `jwt.WithAudience(expectedAudience)` |
| `leeway` | Recommended | `jwt.WithLeeway(30 * time.Second)` |

**Recommended Implementation:**

```go
// middleware/auth.go
func JWTAuth(secret []byte, expectedIssuer, expectedAudience string, now func() time.Time) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... extract token from header ...
            
            token, err := jwt.Parse(tokenString, keyFunc,
                jwt.WithIssuer(expectedIssuer),
                jwt.WithAudience(expectedAudience),
                jwt.WithLeeway(30*time.Second),
                jwt.WithValidMethods([]string{"HS256"}),
            )
            // ... validate and proceed ...
        })
    }
}
```

**Sources:** [golang-jwt documentation](https://pkg.go.dev/github.com/golang-jwt/jwt/v5), [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)

---

### 1.2 No-Auth Guard for Production

**Current Gap:** `JWT_ENABLED=false` can accidentally be deployed to production.

#### Recommended Pattern: Fail-Closed in Production

```go
// config/config.go - Add validation
func (c *Config) Validate() error {
    // ... existing validation ...
    
    // CRITICAL: Fail-closed for production
    if c.IsProduction() && !c.JWTEnabled {
        return fmt.Errorf("JWT_ENABLED must be true in production environment")
    }
    
    // Optional: Require explicit ALLOW_UNAUTHENTICATED=true for staging
    if c.Env == "staging" && !c.JWTEnabled {
        if !c.AllowUnauthenticated {
            return fmt.Errorf("JWT_ENABLED=false in staging requires ALLOW_UNAUTHENTICATED=true")
        }
    }
    
    return nil
}
```

**Alternative: Compile-Time Guard (via build tags)**

```go
// +build production

func init() {
    if os.Getenv("JWT_ENABLED") != "true" {
        panic("JWT_ENABLED must be true in production builds")
    }
}
```

---

### 1.3 Authorization Consistency in Application Layer

**Current Gap:** `GetUserUseCase` has authz check, but `CreateUserUseCase` and `ListUsersUseCase` do not.

#### Recommended Pattern: Centralized Authorization

```go
// app/authz/policy.go
type Permission string

const (
    PermUserRead   Permission = "user:read"
    PermUserCreate Permission = "user:create"
    PermUserList   Permission = "user:list"
)

var rolePermissions = map[string][]Permission{
    "admin": {PermUserRead, PermUserCreate, PermUserList},
    "user":  {PermUserRead},  // Only own resources
}

// app/authz/checker.go
func RequirePermission(ctx context.Context, perm Permission) error {
    authCtx := GetAuthContext(ctx)
    if authCtx == nil {
        return app.ErrForbidden
    }
    
    perms, ok := rolePermissions[authCtx.Role]
    if !ok || !contains(perms, perm) {
        return app.ErrForbidden
    }
    return nil
}
```

**Apply in Use Cases:**

```go
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
    if err := authz.RequirePermission(ctx, authz.PermUserCreate); err != nil {
        return nil, err
    }
    // ... proceed with creation ...
}
```

---

### 1.4 Internal Endpoint Protection (/metrics)

**Current Gap:** `/metrics` endpoint exposed without authentication.

#### Options

| Option | Complexity | Security | Recommendation |
|--------|------------|----------|----------------|
| Separate port (internal-only) | Medium | High | ✅ Recommended |
| Basic auth middleware | Low | Medium | Alternative |
| Network firewall only | Low | Medium | Minimum |

**Recommended: Dual-Server Pattern**

```go
// cmd/api/main.go
func run() error {
    // Public API server
    publicRouter := httpTransport.NewRouter(...)
    publicSrv := &http.Server{Addr: ":8080", Handler: publicRouter}
    
    // Internal metrics server (different port)
    internalRouter := chi.NewRouter()
    internalRouter.Handle("/metrics", promhttp.HandlerFor(metricsReg, promhttp.HandlerOpts{}))
    internalRouter.Get("/health", healthHandler.ServeHTTP)
    internalRouter.Get("/ready", readyHandler.ServeHTTP)
    internalSrv := &http.Server{Addr: ":9090", Handler: internalRouter}
    
    // Start both servers...
}
```

**Kubernetes Deployment:**

```yaml
# Only internal port exposed to Prometheus service mesh
spec:
  containers:
    - name: api
      ports:
        - containerPort: 8080  # Public API
        - containerPort: 9090  # Internal metrics (not in Service)
```

---

### 1.5 Request-ID and Actor-ID Injection

**Current Gap:** Handler not populating `RequestID` and `ActorID` from context to use case request.

**Fix Pattern:**

```go
// handler/user.go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // ... parse body ...
    
    // CRITICAL: Inject correlation + actor from context
    ucReq := user.CreateUserRequest{
        Email:     body.Email,
        FirstName: body.FirstName,
        LastName:  body.LastName,
        RequestID: middleware.GetRequestID(ctx),  // From middleware
        ActorID:   extractActorID(ctx),           // From JWT claims
    }
    
    resp, err := h.createUserUC.Execute(ctx, ucReq)
    // ...
}

func extractActorID(ctx context.Context) domain.ID {
    authCtx := app.GetAuthContext(ctx)
    if authCtx == nil {
        return domain.ID{} // Empty for unauthenticated
    }
    id, err := domain.ParseID(authCtx.SubjectID)
    if err != nil {
        return domain.ID{} // Invalid UUID, treat as anonymous
    }
    return id
}
```

---

### 1.6 Proxy/IP Trust Configuration

**Current Gap:** `RealIP` middleware always active; `TRUST_PROXY` only affects rate limiting.

**Recommendation:** Conditional RealIP middleware

```go
// router.go
if rateLimitConfig.TrustProxy {
    r.Use(chiMiddleware.RealIP)
} else {
    // Use direct connection IP only (safer default)
}
```

**Warning:** Never trust proxy headers in production without proper network-level controls.

---

## 2. Secret Management Patterns

### 2.1 Env + *_FILE Pattern (12-Factor Compatible)

**Approach:** Application reads from environment or file, completely unaware of Vault/K8s.

#### Config Implementation

```go
// config/config.go
type Config struct {
    DatabaseURL     string `envconfig:"DATABASE_URL"`      // Direct value
    DatabaseURLFile string `envconfig:"DATABASE_URL_FILE"` // Or file path
    
    JWTSecret       string `envconfig:"JWT_SECRET"`
    JWTSecretFile   string `envconfig:"JWT_SECRET_FILE"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }
    
    // Resolve *_FILE patterns
    cfg.DatabaseURL = resolveSecret(cfg.DatabaseURL, cfg.DatabaseURLFile)
    cfg.JWTSecret = resolveSecret(cfg.JWTSecret, cfg.JWTSecretFile)
    
    return &cfg, cfg.Validate()
}

func resolveSecret(direct, filePath string) string {
    if direct != "" {
        return direct
    }
    if filePath == "" {
        return ""
    }
    data, err := os.ReadFile(filePath)
    if err != nil {
        slog.Warn("failed to read secret file", "path", filePath, "err", err)
        return ""
    }
    return strings.TrimSpace(string(data))
}
```

**Sources:** [12factor.net/config](https://12factor.net/config), [Docker secrets](https://docs.docker.com/engine/swarm/secrets/)

---

### 2.2 Vault Agent Injection (Kubernetes)

**Pattern:** Vault Agent sidecar writes secrets to pod filesystem; app reads as files.

**Kubernetes Annotations:**

```yaml
spec:
  serviceAccountName: golang-api
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "golang-api"
        vault.hashicorp.com/agent-inject-secret-db: "secret/data/golang-api/database"
        vault.hashicorp.com/agent-inject-template-db: |
          {{- with secret "secret/data/golang-api/database" -}}
          {{ .Data.data.url }}
          {{- end -}}
```

**App Environment:**

```yaml
env:
  - name: DATABASE_URL_FILE
    value: "/vault/secrets/db"
```

**Benefits:**
- App is Vault-unaware (portable)
- Automatic secret rotation via Vault Agent
- Least-privilege via Kubernetes service account

**Sources:** [HashiCorp Vault K8s Integration](https://developer.hashicorp.com/vault/docs/platform/k8s/injector)

---

### 2.3 Secret Rotation and Pool Lifecycle

**Current Gap:** `reconnectingDB` design can cause stale pool references.

**Impact on Secrets:** Dynamic database credentials require pool recreation.

**Recommended Approach:**

1. **Static credentials:** Use long-lived credentials from Vault; pool doesn't need rotation
2. **Dynamic credentials:** Pool must gracefully handle credential expiry

```go
// Option A: Simple - fail-fast, rely on pod restart
// Good for: Kubernetes with liveness probes
func (r *reconnectingDB) Ping(ctx context.Context) error {
    r.mu.RLock()
    pool := r.pool
    r.mu.RUnlock()
    
    if pool == nil {
        return fmt.Errorf("database pool not initialized")
    }
    
    // Just report error, don't try to reconnect
    return pool.Ping(ctx)
}

// Option B: Dynamic - recreate pool with new credentials
// Good for: Long-running services with dynamic secrets
type ReloadableDB struct {
    dsn     func() string  // Function that reads current DSN
    pool    atomic.Pointer[pgxpool.Pool]
    // ... 
}
```

---

### 2.4 Cloud Secret Manager Integration

**Pattern:** Use External Secrets Operator or CSI Driver to sync cloud secrets to K8s secrets.

| Cloud | Tool | Integration |
|-------|------|-------------|
| AWS | External Secrets Operator | Syncs to K8s Secret → env/file mount |
| GCP | Secret Manager CSI Driver | Direct file mount |
| Azure | CSI Driver + identity | File mount |

**App remains unaware** - just reads `*_FILE` or env vars.

---

## 3. Dependency Injection with Google Wire

### 3.1 Wire vs Fx Comparison

| Aspect | Google Wire | Uber Fx |
|--------|-------------|---------|
| **Approach** | Compile-time code generation | Runtime reflection |
| **Error Detection** | Compile-time | Runtime startup |
| **Performance** | Zero overhead | Small startup cost |
| **Lifecycle** | Manual (shutdown hooks) | Built-in OnStart/OnStop |
| **Learning Curve** | Lower | Higher |
| **Flexibility** | Less (static graph) | More (dynamic) |
| **Boilerplate Use** | ✅ Ideal | Overkill |

**Decision: Google Wire** - Aligns with hexagonal architecture goals (explicit, deterministic, minimal magic).

**Sources:** [Wire GitHub](https://github.com/google/wire), [Uber Fx GitHub](https://github.com/uber-go/fx), [Wire vs Fx comparison](https://medium.com/@pliutau/google-wire-vs-uber-fx)

---

### 3.2 Wire Migration Strategy

**Current:** `cmd/api/main.go` has manual wiring (~100 lines)

**Target Structure:**

```
cmd/api/
├── main.go           # Minimal: run() and graceful shutdown
├── wire.go           # Wire injector definition (compile-time)
└── wire_gen.go       # Generated code (do not edit)

internal/
└── bootstrap/
    ├── providers.go  # Wire provider functions
    └── sets.go       # Wire provider sets
```

**Step 1: Create Provider Functions**

```go
// internal/bootstrap/providers.go
package bootstrap

import (
    "github.com/iruldev/golang-api-hexagonal/internal/infra/config"
    "github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
    // ...
)

func ProvideConfig() (*config.Config, error) {
    return config.Load()
}

func ProvidePool(cfg *config.Config) (*postgres.Pool, func(), error) {
    pool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
    if err != nil {
        return nil, nil, err
    }
    cleanup := func() { pool.Close() }
    return pool, cleanup, nil
}

func ProvideUserRepo() domain.UserRepository {
    return postgres.NewUserRepo()
}

// ... more providers ...
```

**Step 2: Create Provider Sets**

```go
// internal/bootstrap/sets.go
package bootstrap

import "github.com/google/wire"

var InfraSet = wire.NewSet(
    ProvideConfig,
    ProvidePool,
    ProvideQuerier,
    ProvideTxManager,
)

var RepoSet = wire.NewSet(
    ProvideUserRepo,
    ProvideAuditEventRepo,
    ProvideIDGenerator,
)

var UseCaseSet = wire.NewSet(
    user.NewCreateUserUseCase,
    user.NewGetUserUseCase,
    user.NewListUsersUseCase,
    audit.NewAuditService,
)
```

**Step 3: Create Injector**

```go
// cmd/api/wire.go
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/iruldev/golang-api-hexagonal/internal/bootstrap"
    httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
)

type App struct {
    Router  chi.Router
    Cleanup func()
}

func InitializeApp() (*App, error) {
    wire.Build(
        bootstrap.InfraSet,
        bootstrap.RepoSet,
        bootstrap.UseCaseSet,
        bootstrap.HandlerSet,
        ProvideRouter,
        wire.Struct(new(App), "*"),
    )
    return nil, nil
}
```

**Step 4: CI Integration**

```yaml
# .github/workflows/ci.yml
- name: Check Wire generated code
  run: |
    go install github.com/google/wire/cmd/wire@latest
    wire ./cmd/api/...
    git diff --exit-code cmd/api/wire_gen.go || (echo "wire_gen.go is out of date" && exit 1)
```

---

### 3.3 Lifecycle Management

Wire doesn't provide lifecycle hooks. Implement manually:

```go
// cmd/api/main.go
func run() error {
    app, err := InitializeApp()
    if err != nil {
        return err
    }
    defer app.Cleanup() // Cleanup all resources in reverse order
    
    srv := &http.Server{Addr: ":8080", Handler: app.Router}
    
    // Graceful shutdown...
}
```

---

## 4. sqlc Adoption Strategy

### 4.1 sqlc + pgx Integration

**Setup:** `sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/infra/postgres/queries/"
    schema: "migrations/"
    gen:
      go:
        package: "sqlcgen"
        out: "internal/infra/postgres/sqlcgen"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: true
```

**Sources:** [sqlc.dev](https://docs.sqlc.dev/en/stable/)

---

### 4.2 Transaction Bridging Pattern

**Challenge:** sqlc generates methods on `*Queries` struct; need to bridge to existing `TxManager`.

**Solution: DBTX Interface**

sqlc generates with interface:

```go
// sqlcgen/db.go (generated)
type DBTX interface {
    Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
    Query(context.Context, string, ...interface{}) (pgx.Rows, error)
    QueryRow(context.Context, string, ...interface{}) pgx.Row
}
```

**Bridging:**

```go
// internal/infra/postgres/user_repo.go
type UserRepo struct {
    // Embed sqlc queries for convenience
}

func (r *UserRepo) Create(ctx context.Context, q domain.Querier, user *domain.User) error {
    // Bridge domain.Querier to sqlc's DBTX
    sqlcQ := sqlcgen.New(q.Underlying()) // q.Underlying() returns *pgxpool.Pool or pgx.Tx
    
    return sqlcQ.CreateUser(ctx, sqlcgen.CreateUserParams{
        ID:        user.ID.UUID(),
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    })
}
```

**Alternative: Modify domain.Querier**

```go
// domain/querier.go
type Querier interface {
    Underlying() any // Returns driver-specific type (pool or tx)
}
```

---

### 4.3 Domain Isolation

**Rule:** sqlc-generated code stays in `infra/postgres/sqlcgen/`. Repository wrappers map to domain types.

```
internal/
├── domain/
│   └── user.go           # domain.User (clean)
└── infra/postgres/
    ├── sqlcgen/          # sqlc generated (DO NOT import in domain/app)
    │   ├── db.go
    │   ├── models.go
    │   └── user.sql.go
    └── user_repo.go      # Maps sqlcgen.User → domain.User
```

---

### 4.4 Testing Strategy

**Unit Tests (Mocked):**

```go
// Use mockgen on domain.UserRepository, not sqlc Queries
mockRepo := mocks.NewMockUserRepository(ctrl)
uc := user.NewCreateUserUseCase(mockRepo, ...)
```

**Integration Tests:**

```go
//go:build integration

func TestUserRepo_Create(t *testing.T) {
    pool := setupTestDB(t)
    repo := postgres.NewUserRepo()
    querier := postgres.NewPoolQuerier(pool)
    
    user := &domain.User{ID: domain.NewID(), Email: "test@example.com", ...}
    err := repo.Create(context.Background(), querier, user)
    require.NoError(t, err)
    
    // Verify with direct SQL
    var count int
    pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE id = $1", user.ID.UUID()).Scan(&count)
    assert.Equal(t, 1, count)
}
```

**CI Job:**

```yaml
- name: Integration Tests
  run: go test -tags=integration ./internal/infra/...
  env:
    DATABASE_URL: postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable
```

---

## 5. Implementation Roadmap

### Phase 1: P0 Bug Fixes (1-2 days)

| # | Issue | Fix | Files |
|---|-------|-----|-------|
| 1 | AUDIT_REDACT_EMAIL not used | Use `cfg.AuditRedactEmail` instead of hardcode | `main.go` |
| 2 | reconnectingDB pool reset | Remove pool close in Ping(); just return error | `main.go` |
| 3 | Missing RequestID/ActorID | Inject from context in handlers | `handler/user.go` |

### Phase 2: Security Baseline (3-5 days)

| # | Task | Priority |
|---|------|----------|
| 1 | Add no-auth guard for production | P1 |
| 2 | Consistent authz in all use cases | P1 |
| 3 | Fix UUID v7 error handling | P1 |
| 4 | JWT claim validation (iss/aud/leeway) | P1 |
| 5 | Separate metrics port | P1 |

### Phase 3: Secret Management (2-3 days)

| # | Task |
|---|------|
| 1 | Implement `*_FILE` pattern in config |
| 2 | Document Vault Agent Kubernetes annotations |
| 3 | Update .env.example with all `*_FILE` variants |

### Phase 4: Google Wire DI (3-5 days)

| # | Task |
|---|------|
| 1 | Create `internal/bootstrap/` providers |
| 2 | Define provider sets |
| 3 | Create `cmd/api/wire.go` injector |
| 4 | Migrate main.go to use Wire |
| 5 | Add CI check for wire_gen.go |

### Phase 5: sqlc Adoption (5-7 days)

| # | Task |
|---|------|
| 1 | Add sqlc.yaml configuration |
| 2 | Create query files for existing tables |
| 3 | Generate sqlc code |
| 4 | Update repositories to use sqlc |
| 5 | Add `Underlying()` to Querier interface |
| 6 | Add integration test CI job |

---

## 6. Conclusions & Recommendations

### Final Recommendations

| Area | Decision | Rationale |
|------|----------|-----------|
| **DI Framework** | Google Wire | Compile-time safety, zero runtime overhead, deterministic |
| **Secret Management** | Env + `*_FILE` pattern | Portable, Vault-unaware, 12-factor compliant |
| **Vault Integration** | Agent injection | App stays simple, rotation handled by platform |
| **sqlc** | Adopt infra-only | Type-safe SQL, domain isolation maintained |
| **Production Auth** | Fail-closed | `JWT_ENABLED=false` blocked in production |

### Priority Order

1. **Immediate (P0):** Fix bugs (config not used, pool reset, missing metadata)
2. **High (P1):** Security baseline (authz, no-auth guard, JWT validation)
3. **Medium (P2):** Secret management (`*_FILE` pattern)
4. **Medium (P2):** Wire migration (cleaner wiring)
5. **Lower (P3):** sqlc adoption (can be incremental)

### Minimal Changes for Secure-by-Default

```diff
# Production readiness checklist
+ [ ] config.Validate() blocks JWT_ENABLED=false in production
+ [ ] All use cases have authz.RequirePermission() check
+ [ ] RequestID/ActorID injected in all handlers
+ [ ] /metrics on separate internal port (9090)
+ [ ] JWT validation includes iss, aud, leeway
+ [ ] *_FILE pattern supported for all secrets
```

---

## Sources

1. [golang-jwt/jwt documentation](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
2. [HashiCorp Vault Agent Injector](https://developer.hashicorp.com/vault/docs/platform/k8s/injector)
3. [Google Wire GitHub](https://github.com/google/wire)
4. [Uber Fx GitHub](https://github.com/uber-go/fx)
5. [sqlc documentation](https://docs.sqlc.dev/)
6. [12-Factor App Config](https://12factor.net/config)
7. [OWASP JWT Security](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
