# AI Assistant Guide (AGENTS.md)

This document serves as a contract between AI assistants (e.g., GitHub Copilot, Claude, ChatGPT) and the codebase. Following these guidelines ensures consistent, high-quality contributions.

---

## ✅ DO

### Architecture

- **DO** follow hexagonal (ports and adapters) architecture
- **DO** respect layer boundaries: domain → usecase → interface → infra
- **DO** define interfaces in the domain layer (ports)
- **DO** implement interfaces in the infra layer (adapters)

### Code Style

- **DO** use standard Go idioms and conventions
- **DO** prefer the standard library over third-party packages
- **DO** write clear, self-documenting code
- **DO** add comments for non-obvious logic
- **DO** use meaningful variable and function names

### Patterns

- **DO** use the repository pattern for data access
- **DO** use the response envelope pattern for HTTP responses
- **DO** use table-driven tests with AAA pattern
- **DO** use dependency injection via constructors
- **DO** use sentinel errors for domain errors

### Testing

- **DO** write unit tests for all new code
- **DO** use mocks for external dependencies
- **DO** maintain ≥70% test coverage
- **DO** write integration tests for HTTP handlers

---

## ❌ DON'T

### Architecture

- **DON'T** import from `interface/` or `infra/` in the domain layer
- **DON'T** bypass layers (e.g., handler calling repo directly)
- **DON'T** put business logic in handlers
- **DON'T** put HTTP concerns in use cases

### Code Style

- **DON'T** use `panic` for error handling (except truly unrecoverable)
- **DON'T** ignore error returns
- **DON'T** use global state
- **DON'T** write clever code over clear code

### Patterns

- **DON'T** create new patterns without referencing existing ones
- **DON'T** skip validation in domain entities
- **DON'T** return raw database errors to HTTP layer
- **DON'T** use magic numbers/strings

### Testing

- **DON'T** skip tests for "simple" code
- **DON'T** write tests that depend on external services
- **DON'T** use `time.Sleep` in tests (use mocks/channels)
- **DON'T** leave commented-out test code

---

## 📁 File Structure Conventions

### Per Domain Structure

```
internal/
├── domain/{name}/
│   ├── entity.go           # Main entity struct with Validate()
│   ├── entity_test.go      # Entity unit tests
│   ├── errors.go           # Domain-specific sentinel errors
│   └── repository.go       # Repository interface (port)
│
├── usecase/{name}/
│   ├── usecase.go          # Business logic with repo dependency
│   └── usecase_test.go     # Unit tests with mock repository
│
├── interface/http/{name}/
│   ├── handler.go                    # HTTP handlers
│   ├── handler_test.go               # Handler unit tests
│   ├── handler_integration_test.go   # Integration tests (build-tagged)
│   └── dto.go                        # Request/Response DTOs
│
└── infra/postgres/{name}/
    └── (sqlc-generated files)
```

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | snake_case.go | `note_handler.go` |
| Packages | lowercase | `note`, `postgres` |
| Types | PascalCase | `NoteHandler` |
| Exported funcs | PascalCase | `NewHandler()` |
| Private funcs | camelCase | `handleError()` |
| Variables | camelCase | `noteID` |
| Constants | PascalCase | `MaxTitleLength` |
| Errors | Err prefix | `ErrNoteNotFound` |

### Database Conventions

| Element | Location |
|---------|----------|
| Migrations | `db/migrations/YYYYMMDDHHMMSS_description.{up,down}.sql` |
| SQLC queries | `db/queries/{name}.sql` |
| Generated code | `internal/infra/postgres/{name}/` |

---

## 🧪 Testing Requirements

### Unit Tests

```go
// Required: Table-driven test with AAA pattern
func TestUsecase_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {"valid", "Title", nil},
        {"empty", "", ErrEmptyTitle},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            repo := &MockRepository{}
            uc := NewUsecase(repo)
            
            // Act
            _, err := uc.Create(ctx, tt.input, "content")
            
            // Assert
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### Coverage Requirements

| Layer | Minimum Coverage |
|-------|------------------|
| Domain | 90% |
| Use Case | 80% |
| Handler | 70% |
| Overall | 70% |

### Mock Pattern

```go
// Mock repository for testing
type MockRepository struct {
    CreateFunc func(ctx context.Context, n *Note) error
    GetFunc    func(ctx context.Context, id uuid.UUID) (*Note, error)
    // ... other methods
}

func (m *MockRepository) Create(ctx context.Context, n *Note) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, n)
    }
    return nil
}
```

### Integration Tests

```go
//go:build integration
// +build integration

func TestHandler_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup httptest.NewServer with real router
    srv := httptest.NewServer(router)
    defer srv.Close()
    
    // Make real HTTP requests
    resp, _ := http.Get(srv.URL + "/api/v1/notes")
    // Assert response
}
```

---

## 🔧 Common Tasks

### Adding a New Domain

1. Create entity: `internal/domain/{name}/entity.go`
2. Create errors: `internal/domain/{name}/errors.go`
3. Create repository interface: `internal/domain/{name}/repository.go`
4. Create migration: `db/migrations/{timestamp}_{description}.up.sql`
5. Create SQLC queries: `db/queries/{name}.sql`
6. Run `make sqlc`
7. Create usecase: `internal/usecase/{name}/usecase.go`
8. Create HTTP handler: `internal/interface/http/{name}/handler.go`
9. Create DTOs: `internal/interface/http/{name}/dto.go`
10. Register routes in router
11. Write tests for all layers

### Error Handling Flow

```
Domain Error (ErrNoteNotFound)
    ↓
Usecase (returns domain error)
    ↓
Handler (maps to HTTP status)
    ↓
Response (JSON envelope with error code)
```

### Adding Auth Middleware

The auth middleware interface enables pluggable authentication providers. See `internal/interface/http/middleware/auth.go`.

> **📚 For comprehensive architecture documentation including SSO/IDP integration patterns, OAuth2/OIDC examples, and security best practices, see [`docs/architecture.md#Security-Architecture`](docs/architecture.md#security-architecture).**

#### Quick Guide: Implementing Custom Auth Provider

To integrate with external identity providers (Auth0, Okta, Azure AD, etc.), implement the `Authenticator` interface:

> [!TIP]
> For OIDC/JWKS validation, use `github.com/lestrrat-go/jwx/v2/jwk` and `github.com/lestrrat-go/jwx/v2/jwt` packages.

```go
// 1. Define your authenticator struct
type MyOIDCAuthenticator struct {
    keySet   jwk.Set // JWKS for token validation (from github.com/lestrrat-go/jwx/v2/jwk)
    issuer   string
    audience string
}

// 2. Implement the Authenticate method
func (a *MyOIDCAuthenticator) Authenticate(r *http.Request) (middleware.Claims, error) {
    // Extract bearer token
    authHeader := r.Header.Get("Authorization")
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return middleware.Claims{}, middleware.ErrUnauthenticated
    }
    token := strings.TrimPrefix(authHeader, "Bearer ")
    
    // Validate with your IDP's JWKS
    parsed, err := jwt.Parse(token, jwt.WithKeySet(a.keySet))
    if err != nil {
        return middleware.Claims{}, middleware.ErrTokenInvalid
    }
    
    // Map claims to internal struct
    return middleware.Claims{
        UserID:      parsed.Subject(),
        Roles:       extractRoles(parsed),
        Permissions: extractPermissions(parsed),
    }, nil
}

// 3. Use with AuthMiddleware
r.Use(middleware.AuthMiddleware(myOIDCAuth))
```

#### Common Mistakes to Avoid

| Mistake | Problem | Solution |
|---------|---------|----------|
| Hardcoding secrets | Security vulnerability | Use env vars or secret provider |
| Short JWT secret | Weak cryptography | Use ≥32 bytes for HMAC-SHA256 |
| Missing issuer/audience validation | Token from wrong source accepted | Always validate `iss` and `aud` |
| Logging tokens/keys | Credential leakage | Never log auth credentials |
| Using 401 for authorization failures | Incorrect HTTP semantics | Use 401 for authn, 403 for authz |
| Rate limiting after auth | DoS on expensive validation | Rate limit before authentication |
| Not handling token expiry | Stale sessions | Return `ErrTokenExpired` appropriately |

#### Core Components

| Component | Location | Purpose |
|-----------|----------|---------|
| `Authenticator` interface | `middleware/auth.go` | Port for auth providers |
| `Claims` struct | `middleware/auth.go` | Authenticated user info |
| `AuthMiddleware` | `middleware/auth.go` | HTTP middleware wrapper |
| Sentinel errors | `middleware/auth.go` | Error type checking |

#### Implementing an Authenticator

```go
// JWT Authenticator example
type JWTAuthenticator struct {
    secretKey []byte
}

func (a *JWTAuthenticator) Authenticate(r *http.Request) (middleware.Claims, error) {
    token := r.Header.Get("Authorization")
    if token == "" {
        return middleware.Claims{}, middleware.ErrUnauthenticated
    }
    // Validate and parse token...
    return middleware.Claims{
        UserID: "user-123",
        Roles:  []string{"admin"},
    }, nil
}
```

#### Using JWTAuthenticator (Built-in)

The framework includes a ready-to-use JWT authenticator. See `internal/interface/http/middleware/jwt.go`.

```go
import "github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"

// Create authenticator (secret must be ≥32 bytes)
jwtAuth, err := middleware.NewJWTAuthenticator(
    []byte(os.Getenv("JWT_SECRET")),
    middleware.WithIssuer("my-app"),     // Optional: validates "iss" claim
    middleware.WithAudience("my-api"),   // Optional: validates "aud" claim
)
if err != nil {
    log.Fatal("JWT config error:", err)  // ErrSecretKeyTooShort if <32 bytes
}

// Use with AuthMiddleware
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtAuth))
    r.Get("/api/v1/protected", protectedHandler)
})
```

#### JWT Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | HMAC-SHA256 secret (≥32 bytes) | `your-secret-key-at-least-32-bytes!!` |
| `JWT_ISSUER` | (Optional) Expected token issuer | `my-app` |
| `JWT_AUDIENCE` | (Optional) Expected token audience | `my-api` |

#### Using Auth Middleware in Routes

```go
// Protected routes
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtAuth))
    r.Get("/api/v1/notes", noteHandler.List)
})
```

#### Extracting Claims in Handlers

```go
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    claims, err := middleware.FromContext(r.Context())
    if err != nil {
        // Handle error (shouldn't occur if middleware applied)
    }
    
    if claims.HasRole("admin") {
        // Admin-specific logic
    }
    
    if claims.HasPermission("notes:delete") {
        // Permission-specific logic
    }
}
```

#### Auth Error Types

| Error | When Returned | HTTP Status |
|-------|---------------|-------------|
| `ErrUnauthenticated` | No/invalid credentials | 401 |
| `ErrTokenExpired` | Token has expired | 401 |
| `ErrTokenInvalid` | Malformed/bad signature | 401 |
| `ErrNoClaimsInContext` | Claims missing from ctx | 500 |

#### Using APIKeyAuthenticator (Built-in)

The framework includes a ready-to-use API key authenticator for service-to-service auth. See `internal/interface/http/middleware/apikey.go`.

```go
import "github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"

// Create validator from environment variables
// Format: API_KEYS="key1:svc-payments,key2:svc-inventory"
validator := middleware.NewEnvKeyValidator("API_KEYS")

// Create authenticator (validator is required)
apiAuth, err := middleware.NewAPIKeyAuthenticator(
    validator,
    middleware.WithHeaderName("X-Custom-Key"),  // Optional: default is X-API-Key
)
if err != nil {
    log.Fatal("API Key config error:", err)  // ErrValidatorRequired if nil
}

// Use with AuthMiddleware
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(apiAuth))
    r.Get("/api/v1/internal", internalHandler)
})
```

#### API Key Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `API_KEYS` | Comma-separated key:service pairs | `abc123:svc-payments,xyz789:svc-inventory` |

#### Custom Key Validators

Implement the `KeyValidator` interface for database or external validation:

```go
type KeyValidator interface {
    Validate(ctx context.Context, key string) (*KeyInfo, error)
}

// Example: Database validator
type DBKeyValidator struct {
    db *sql.DB
}

func (v *DBKeyValidator) Validate(ctx context.Context, key string) (*middleware.KeyInfo, error) {
    var serviceID string
    err := v.db.QueryRowContext(ctx, "SELECT service_id FROM api_keys WHERE key = $1", key).Scan(&serviceID)
    if err != nil {
        return nil, middleware.ErrTokenInvalid
    }
    return &middleware.KeyInfo{
        ServiceID: serviceID,
        Roles:     []string{"service"},
    }, nil
}
```

#### RBAC Authorization Middleware

After authentication, use RBAC middleware to control access based on roles or permissions. See `internal/interface/http/middleware/rbac.go` and `internal/domain/auth/rbac.go`.

##### Available Roles

| Role | Constant | Purpose |
|------|----------|---------|
| Admin | `auth.RoleAdmin` | Full system access |
| Service | `auth.RoleService` | Service-to-service auth |
| User | `auth.RoleUser` | Standard user access |

##### Available Permissions

Permissions follow the `resource:action` pattern:

| Permission | Constant | Description |
|------------|----------|-------------|
| `note:create` | `auth.PermNoteCreate` | Create new notes |
| `note:read` | `auth.PermNoteRead` | Read notes |
| `note:update` | `auth.PermNoteUpdate` | Update notes |
| `note:delete` | `auth.PermNoteDelete` | Delete notes |
| `note:list` | `auth.PermNoteList` | List all notes |

##### RequireRole Middleware (OR Logic)

Requires user to have **at least one** of the specified roles:

```go
import "github.com/iruldev/golang-api-hexagonal/internal/domain/auth"

r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtAuth))
    r.Use(middleware.RequireRole(string(auth.RoleAdmin), string(auth.RoleService)))
    r.Delete("/admin/users/{id}", deleteUserHandler)
})
```

##### RequirePermission Middleware (AND Logic)

Requires user to have **all** of the specified permissions:

```go
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtAuth))
    r.Use(middleware.RequirePermission(string(auth.PermNoteCreate), string(auth.PermNoteRead)))
    r.Post("/notes", createNoteHandler)
})
```

##### RequireAnyPermission Middleware (OR Logic)

Requires user to have **at least one** of the specified permissions:

```go
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtAuth))
    r.Use(middleware.RequireAnyPermission(string(auth.PermNoteUpdate), string(auth.PermNoteDelete)))
    r.Patch("/notes/{id}", modifyNoteHandler)
})
```

##### Combined Auth + RBAC Pattern

```go
r.Route("/api/v1", func(r chi.Router) {
    // Apply auth to all API routes
    r.Use(middleware.AuthMiddleware(jwtAuth))
    
    // User-accessible routes
    r.Get("/notes", noteHandler.List)
    
    // Admin-only routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireRole(string(auth.RoleAdmin)))
        r.Delete("/notes/{id}", noteHandler.Delete)
    })
})
```

##### RBAC Error Types

| Error | When Returned | HTTP Status |
|-------|---------------|-------------|
| `ErrForbidden` | Authorization failed | 403 |
| `ErrInsufficientRole` | Missing required role | 403 |
| `ErrInsufficientPermission` | Missing required permission | 403 |

### Rate Limiting Middleware

The rate limiting middleware protects endpoints from abuse using a token bucket algorithm. See `internal/interface/http/middleware/ratelimit.go`.

#### Core Components

| Component | Location | Purpose |
|-----------|----------|---------|
| `TokenBucket` | `middleware/ratelimit.go` | Token bucket algorithm implementation |
| `InMemoryRateLimiter` | `middleware/ratelimit.go` | Thread-safe in-memory limiter |
| `RateLimitMiddleware` | `middleware/ratelimit.go` | HTTP middleware wrapper |
| `IPKeyExtractor` | `middleware/ratelimit.go` | Extract client IP for rate limiting |
| `UserIDKeyExtractor` | `middleware/ratelimit.go` | Extract user ID from auth claims |

#### Basic Usage

```go
import "github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
import "github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"

// Create rate limiter: 100 requests per minute
limiter := middleware.NewInMemoryRateLimiter(
    middleware.WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
)
defer limiter.Stop()

// Apply to routes
r.Use(middleware.RateLimitMiddleware(limiter))
```

#### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithDefaultRate(rate)` | 100 req/min | Default rate for new keys |
| `WithCleanupInterval(d)` | 5 min | Interval for cleaning expired buckets |
| `WithBucketTTL(d)` | 10 min | How long to keep inactive buckets |
| `WithKeyExtractor(fn)` | IP address | Function to extract key from request |
| `WithRetryAfterSeconds(n)` | dynamic | Seconds to return in Retry-After header |

#### Per-Endpoint Rate Limits

```go
// Strict rate limit for sensitive endpoints
strictLimiter := middleware.NewInMemoryRateLimiter(
    middleware.WithDefaultRate(runtimeutil.NewRate(10, time.Minute)),
)

// Normal rate limit for public endpoints  
normalLimiter := middleware.NewInMemoryRateLimiter(
    middleware.WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
)

r.Group(func(r chi.Router) {
    r.Use(middleware.RateLimitMiddleware(strictLimiter))
    r.Post("/auth/login", loginHandler)
})

r.Group(func(r chi.Router) {
    r.Use(middleware.RateLimitMiddleware(normalLimiter))
    r.Get("/notes", notesHandler)
})
```

#### User-Based Rate Limiting

```go
// Rate limit by authenticated user ID instead of IP
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtAuth))
    r.Use(middleware.RateLimitMiddleware(limiter,
        middleware.WithKeyExtractor(middleware.UserIDKeyExtractor),
    ))
    r.Get("/api/v1/me", userHandler)
})
```

#### Custom Key Extractor

```go
// Rate limit by API key header
apiKeyExtractor := func(r *http.Request) string {
    apiKey := r.Header.Get("X-API-Key")
    if apiKey == "" {
        return middleware.IPKeyExtractor(r) // Fallback
    }
    return "apikey:" + apiKey
}

r.Use(middleware.RateLimitMiddleware(limiter,
    middleware.WithKeyExtractor(apiKeyExtractor),
))
```

#### Response Format

When rate limit is exceeded:

```json
HTTP/1.1 429 Too Many Requests
Retry-After: 60
Content-Type: application/json

{
  "success": false,
  "error": {
    "code": "ERR_RATE_LIMITED",
    "message": "Rate limit exceeded"
  }
}
```

#### Security Considerations

| Consideration | Implementation |
|---------------|----------------|
| Fail-open | On limiter error, requests are allowed through |
| Memory limits | Automatic cleanup of expired buckets |
| Thread-safety | Uses sync.Map and mutexes |
| X-Forwarded-For | Documented spoofing risk when behind proxy |

#### Redis-Backed Rate Limiter (Distributed)

For multi-instance deployments, use the Redis rate limiter. See `internal/infra/redis/ratelimiter.go`.

##### Setup

```go
import (
    infraredis "github.com/iruldev/golang-api-hexagonal/internal/infra/redis"
    "github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
    "github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// Create Redis rate limiter (shared across instances)
redisLimiter := infraredis.NewRedisRateLimiter(
    redisClient.Client(),
    infraredis.WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
    infraredis.WithKeyPrefix("api:ratelimit:"),
)

// Use with middleware
r.Use(middleware.RateLimitMiddleware(redisLimiter))
```

##### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithRedisDefaultRate(rate)` | 100 req/min | Default rate for new keys |
| `WithKeyPrefix(string)` | `rl:` | Redis key prefix |
| `WithRedisTimeout(duration)` | 100ms | Redis operation timeout |
| `WithFallbackLimiter(limiter)` | nil | Fallback when Redis fails |
| `WithCircuitBreakerConfig(threshold, recovery)` | 5 failures, 30s | Circuit breaker configuration |

##### Fallback Configuration

```go
// Create in-memory fallback
fallback := middleware.NewInMemoryRateLimiter(
    middleware.WithDefaultRate(runtimeutil.NewRate(50, time.Minute)),
)
defer fallback.Stop()

// Redis limiter with fallback
redisLimiter := infraredis.NewRedisRateLimiter(
    redisClient.Client(),
    infraredis.WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
    infraredis.WithFallbackLimiter(fallback),
    infraredis.WithCircuitBreakerConfig(5, 30*time.Second),
)
```

##### Circuit Breaker Behavior

| State | Condition | Behavior |
|-------|-----------|----------|
| Closed | < threshold failures | Uses Redis |
| Open | ≥ threshold failures | Uses fallback, logs warning |
| Half-Open | After recovery time | Attempts Redis, resets on success |

##### Deployment Considerations

| Consideration | Recommendation |
|---------------|----------------|
| Redis availability | Use Redis Cluster/Sentinel for HA |
| Network latency | Set appropriate timeout (100ms default) |
| Memory | Keys have TTL, but monitor at high traffic |
| Fail-open | Requests allowed if Redis fails (controlled degradation) |

### Feature Flags

The feature flag interface enables toggling features without deployment. See `internal/runtimeutil/featureflags.go`.

#### Core Components

| Component | Location | Purpose |
|-----------|----------|---------|
| `FeatureFlagProvider` interface | `runtimeutil/featureflags.go` | Port for feature flag providers |
| `EvalContext` struct | `runtimeutil/featureflags.go` | Context for user targeting |
| `EnvFeatureFlagProvider` | `runtimeutil/featureflags.go` | Environment variable provider |
| `NopFeatureFlagProvider` | `runtimeutil/featureflags.go` | No-op provider for testing |

#### Basic Usage

```go
import "github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"

// Create provider with default settings (FF_ prefix, fail-closed)
provider := runtimeutil.NewEnvFeatureFlagProvider()

// Check if feature is enabled
enabled, err := provider.IsEnabled(ctx, "new_dashboard")
if enabled {
    // Render new dashboard
}
```

#### Environment Variable Naming

| Flag Name | Env Var | Value | Result |
|-----------|---------|-------|--------|
| `new_dashboard` | `FF_NEW_DASHBOARD` | `true` | enabled |
| `beta-feature` | `FF_BETA_FEATURE` | `1` | enabled |
| `dark_mode` | `FF_DARK_MODE` | `enabled` | enabled |
| `experimental` | `FF_EXPERIMENTAL` | `false` | disabled |
| `not_set` | (not set) | - | default (false) |

**Truthy values:** `true`, `1`, `enabled`, `on`, `yes` (case-insensitive)

#### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithEnvPrefix(string)` | `FF_` | Environment variable prefix |
| `WithEnvDefaultValue(bool)` | `false` | Default for unconfigured flags |
| `WithEnvStrictMode(bool)` | `false` | Error on unknown flags |

```go
// Custom prefix and fail-open behavior
provider := runtimeutil.NewEnvFeatureFlagProvider(
    runtimeutil.WithEnvPrefix("FEATURE_"),
    runtimeutil.WithEnvDefaultValue(true),  // Fail-open
)

// Strict mode (error on unknown flags)
provider := runtimeutil.NewEnvFeatureFlagProvider(
    runtimeutil.WithEnvStrictMode(true),
)
enabled, err := provider.IsEnabled(ctx, "unknown_flag")
// err == ErrFlagNotFound
```

#### Context-Based Evaluation

```go
// For future providers that support user targeting
evalCtx := runtimeutil.EvalContext{
    UserID: "user-123",
    Attributes: map[string]interface{}{
        "plan":    "premium",
        "country": "US",
    },
    Percentage: 50.0,  // For gradual rollouts
}

enabled, err := provider.IsEnabledForContext(ctx, "beta_feature", evalCtx)
```

> **Note:** EnvProvider ignores context; use LaunchDarkly, Split.io, etc. for advanced targeting.

#### Testing with NopProvider

```go
// All flags disabled (for testing)
provider := runtimeutil.NewNopFeatureFlagProvider(false)

// All flags enabled (for testing)
provider := runtimeutil.NewNopFeatureFlagProvider(true)
```

#### Error Types

| Error | When Returned | Description |
|-------|---------------|-------------|
| `ErrFlagNotFound` | Strict mode + unknown flag | Flag not configured |
| `ErrInvalidFlagName` | Empty or invalid characters | Flag name validation failed |

#### Usage in HTTP Handlers

```go
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    enabled, _ := h.featureFlags.IsEnabled(r.Context(), "new_list_ui")
    if enabled {
        // New UI logic
    } else {
        // Classic UI logic
    }
}
```

> **For comprehensive async job documentation, see [`docs/async-jobs.md`](docs/async-jobs.md)**

#### Step 1: Choose Your Pattern

| Scenario | Pattern | Package |
|----------|---------|---------|
| Non-critical background (analytics, audit) | Fire-and-Forget | `internal/worker/patterns/fireandforget.go` |
| Periodic tasks (cleanup, reports) | Scheduled | `internal/worker/patterns/scheduled.go` |
| Event → multiple handlers | Fanout | `internal/worker/patterns/fanout.go` |
| Critical operations (payments, orders) | Standard + Idempotency | `internal/worker/idempotency/` |

#### Step 2: Create Task Type

Add to `internal/worker/tasks/types.go`:

```go
const (
    Type{Name} = "{domain}:{action}"  // e.g., TypeEmailSend = "email:send"
)
```

#### Step 3: Create Task Handler

Create `internal/worker/tasks/{name}.go`:

```go
// 1. Payload struct
type {Name}Payload struct {
    ID uuid.UUID `json:"id"`
}

// 2. Task constructor
func New{Name}Task(id uuid.UUID) (*asynq.Task, error) {
    payload, err := json.Marshal({Name}Payload{ID: id})
    if err != nil {
        return nil, fmt.Errorf("marshal payload: %w", err)
    }
    return asynq.NewTask(Type{Name}, payload, asynq.MaxRetry(3)), nil
}

// 3. Handler struct with dependencies
type {Name}Handler struct {
    logger *zap.Logger
    // Add: repo, usecase, etc.
}

func New{Name}Handler(logger *zap.Logger) *{Name}Handler {
    return &{Name}Handler{logger: logger}
}

// 4. Handle method with validation
func (h *{Name}Handler) Handle(ctx context.Context, t *asynq.Task) error {
    var p {Name}Payload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("unmarshal: %v: %w", err, asynq.SkipRetry)
    }
    if p.ID == uuid.Nil {
        return fmt.Errorf("id required: %w", asynq.SkipRetry)
    }
    // Process task
    return nil
}
```

#### Step 4: Register Handler

Add to `cmd/worker/main.go`:

```go
handler := tasks.New{Name}Handler(logger)
srv.HandleFunc(tasks.Type{Name}, handler.Handle)
```

#### Step 5: Write Tests

Create `internal/worker/tasks/{name}_test.go` with tests for:
- Valid payload handling
- Invalid/empty payload (SkipRetry)
- Missing required fields (SkipRetry)
- Happy path processing

#### Async Job Creation Checklist

- [ ] Task type constant in `internal/worker/tasks/types.go`
- [ ] Payload struct with JSON tags
- [ ] Task constructor (`New{Name}Task`) with default options
- [ ] Handler struct with dependencies
- [ ] `Handle` method with validation
- [ ] `SkipRetry` for validation errors
- [ ] Handler registered in `cmd/worker/main.go`
- [ ] Unit tests in `internal/worker/tasks/{name}_test.go`

#### Copy Commands

```bash
# Copy reference task file
cp internal/worker/tasks/note_archive.go internal/worker/tasks/{name}.go
cp internal/worker/tasks/note_archive_test.go internal/worker/tasks/{name}_test.go

# Replace placeholders (macOS)
sed -i '' 's/NoteArchive/{Name}/g' internal/worker/tasks/{name}.go
sed -i '' 's/note:archive/{domain}:{action}/g' internal/worker/tasks/{name}.go
sed -i '' 's/NoteID/YourFieldID/g' internal/worker/tasks/{name}.go

# Linux: use sed -i without quotes: sed -i 's/NoteArchive/{Name}/g' ...

# Add type constant to types.go
# Manually add: Type{Name} = "{domain}:{action}"

# Register handler in cmd/worker/main.go
# Manually add: srv.HandleFunc(tasks.Type{Name}, handler.Handle)
```

#### Queue Selection Guide

| Priority | Queue | Weight | When to Use |
|----------|-------|--------|-------------|
| High | `critical` | 6 | User-facing (email, notifications) |
| Normal | `default` | 3 | Business logic (archival, sync) |
| Low | `low` | 1 | Analytics, cleanup, batch jobs |

#### Pattern Selection Decision Table

| Your Scenario | Use This Pattern | Queue |
|---------------|------------------|-------|
| Non-critical, best-effort | `patterns.FireAndForget()` | `low` |
| Scheduled cleanup/reports | `patterns.RegisterScheduledJobs()` | `default` |
| Single event → multiple actions | `patterns.Fanout()` | per-handler |
| Prevent duplicate processing | `idempotency.IdempotentHandler()` | any |
| Critical with confirmation | Standard enqueue | `critical` |

---

## 🛠️ CLI Tool (bplat)

The `bplat` CLI tool provides code scaffolding utilities for the boilerplate. Located in `cmd/bplat/`.

### Available Commands

| Command | Description |
|---------|-------------|
| `bplat version` | Print version, build date, git commit, and Go version |
| `bplat init service <name>` | Initialize a new service with complete project structure |
| `bplat generate module <name>` | Generate a new domain module with all layers |
| `bplat --help` | List all available commands |

### Init Service Command

Initialize a new service from the boilerplate template:

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

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--module` | `-m` | `github.com/user/<name>` | Go module path |
| `--dir` | `-d` | `.` | Output directory |
| `--force` | `-f` | `false` | Overwrite existing directory |

#### Generated Structure

```
myservice/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── config/
│   ├── domain/
│   ├── usecase/
│   ├── interface/
│   │   └── http/
│   └── infra/
├── db/
│   ├── migrations/
│   └── queries/
├── .env.example
├── go.mod
└── README.md
```

#### Validation Rules

| Rule | Pattern | Error Message |
|------|---------|---------------|
| Name empty | `len(name) == 0` | "service name is required" |
| Invalid chars | `^[a-z][a-z0-9_-]*$` | "service name must start with letter and contain only lowercase letters, numbers, hyphens, underscores" |
| Dir exists | `os.Stat(dir)` | "directory already exists, use --force to overwrite" |
### Generate Module Command

Generate a new domain module with all hexagonal architecture layers:

```bash
# Basic usage - creates payment module
bplat generate module payment

# With custom entity name (default: singularized module name)
bplat generate module orders --entity Order
```

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--entity` | `-e` | singularized module name | Custom entity name (PascalCase) |

#### Generated Structure

```
internal/
├── domain/payment/
│   ├── entity.go           # Entity with Validate()
│   ├── errors.go           # Domain-specific errors
│   ├── repository.go       # Repository interface
│   └── entity_test.go      # Entity tests
├── usecase/payment/
│   ├── usecase.go          # Business logic
│   └── usecase_test.go     # Usecase tests
└── interface/http/payment/
    ├── handler.go          # HTTP handlers
    ├── dto.go              # Request/Response DTOs
    └── handler_test.go     # Handler tests

db/
├── migrations/
│   ├── {timestamp}_payment.up.sql    # Create table
│   └── {timestamp}_payment.down.sql  # Drop table
└── queries/
    └── payment.sql         # sqlc queries
```

#### Template Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `ModuleName` | Lowercase module name | `payment` |
| `EntityName` | PascalCase entity name | `Payment` |
| `TableName` | Snake_case plural | `payments` |
| `Timestamp` | Migration timestamp | `20251214021630` |
| `ModulePath` | Go module path | `github.com/iruldev/golang-api-hexagonal` |

#### Next Steps After Generation

1. Review and update entity fields in `internal/domain/{name}/entity.go`
2. Update migration in `db/migrations/{timestamp}_{name}.up.sql`
3. Update sqlc queries in `db/queries/{name}.sql`
4. Run: `make sqlc`
5. Register routes in router

#### Validation Rules

| Rule | Pattern | Error Message |
|------|---------|---------------|
| Name empty | `len(name) == 0` | "module name is required" |
| Invalid chars | `^[a-z][a-z0-9_-]*$` | "module name must start with letter and contain only lowercase letters, numbers, hyphens, underscores" |
| Already exists | `os.Stat(domain/path)` | "module already exists" |

### Building and Installing

```bash
# Build to bin/ directory with version info
make build-bplat

# Install to GOPATH/bin
make install-bplat
```

### Version Information

The CLI uses ldflags for build-time version injection:

```bash
# Output example:
bplat version v1.0.0
Build date: 2025-12-14T00:00:00Z
Git commit: abc1234
Go version: go1.24.10
```

### CLI Structure Pattern

Follow this pattern when adding new commands:

```
cmd/bplat/
├── main.go           # Entry point, calls cmd.Execute()
└── cmd/
    ├── root.go       # Root command with Execute() function
    ├── root_test.go  # Root command tests
    ├── version.go    # Version command
    ├── version_test.go
    ├── init.go       # Init parent command
    ├── init_service.go  # Init service subcommand
    └── init_test.go  # Init command tests
```

### Adding New Commands

1. Create `cmd/bplat/cmd/{name}.go` with cobra command
2. Register command in `root.go` via `init()`: `rootCmd.AddCommand({name}Cmd)`
3. Use `cmd.OutOrStdout()` for testable output
4. Create `cmd/bplat/cmd/{name}_test.go` with table-driven tests

### Command Implementation Pattern

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
    Use:   "example",
    Short: "Short description",
    Long:  `Longer description with usage details.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Fprintln(cmd.OutOrStdout(), "Output here")
    },
}

func init() {
    rootCmd.AddCommand(exampleCmd)
}
```

---

## 📋 Checklist for Code Review

Before submitting code, verify:

- [ ] Follows hexagonal architecture
- [ ] No layer violations
- [ ] Uses existing patterns
- [ ] Has unit tests
- [ ] Uses table-driven tests
- [ ] Follows AAA pattern
- [ ] Has meaningful test names
- [ ] Domain has validation
- [ ] Errors are sentinel errors
- [ ] HTTP uses response envelope
- [ ] No global state
- [ ] No `panic` for error handling
- [ ] Comments for non-obvious logic
- [ ] Follows naming conventions
