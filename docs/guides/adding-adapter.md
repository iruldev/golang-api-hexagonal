# Guide: Adding a New Infrastructure Adapter

This guide provides step-by-step instructions for adding new infrastructure adapters to the golang-api-hexagonal project. It follows hexagonal architecture principles and uses existing adapters as reference implementations.

## Table of Contents

- [Overview](#overview)
- [Understanding Adapters](#understanding-adapters)
- [Step 1: Define the Interface (Port)](#step-1-define-the-interface-port)
- [Step 2: Implement the Adapter](#step-2-implement-the-adapter)
- [Step 3: Configure via Environment Variables](#step-3-configure-via-environment-variables)
- [Step 4: Write Tests](#step-4-write-tests)
- [Step 5: Wire in main.go](#step-5-wire-in-maingo)
- [Example: Redis Cache Adapter](#example-redis-cache-adapter)
- [Example: Email Service Adapter](#example-email-service-adapter)
- [Quick Reference Checklist](#quick-reference-checklist)

---

## Overview

In hexagonal architecture, **adapters** are the bridge between your application and external systems (databases, caches, message queues, email services, etc.). They implement interfaces defined in the domain or application layer, allowing the core business logic to remain decoupled from infrastructure concerns.

> [!IMPORTANT]
> **Layer boundaries are enforced by CI via golangci-lint depguard rules.** Violations will fail the build. Read [docs/architecture.md](../architecture.md) for details.

**Estimated Time:** 1-3 hours depending on complexity

**Prerequisites:**
- Familiarity with Go and hexagonal architecture concepts
- Project running locally (`make run`)
- Understanding of the external service you're integrating

---

## Understanding Adapters

### Adapter Types

In hexagonal architecture, there are two types of adapters:

| Type | Direction | Example | Location |
|------|-----------|---------|----------|
| **Driving Adapters** | External → Application | HTTP handlers, CLI, gRPC | `internal/transport/` |
| **Driven Adapters** | Application → External | Database, cache, email | `internal/infra/` |

This guide focuses on **driven adapters** in `internal/infra/`.

### Port vs Adapter

- **Port** = Interface defined in domain (or app layer for auxiliary services)
- **Adapter** = Concrete implementation in infra layer

```
┌─────────────────────────────────────────────────────────────┐
│  Domain Layer                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Port (Interface)                                    │    │
│  │  type UserRepository interface {                     │    │
│  │      Create(ctx, q, user) error                      │    │
│  │      GetByID(ctx, q, id) (*User, error)             │    │
│  │  }                                                   │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ implements
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Infra Layer                                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Adapter (Implementation)                            │    │
│  │  type UserRepo struct{}                              │    │
│  │  func (r *UserRepo) Create(...) error { ... }        │    │
│  │  func (r *UserRepo) GetByID(...) (*User, error) {...}│    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Existing Adapters

| Adapter | Directory | Purpose |
|---------|-----------|---------|
| PostgreSQL | `internal/infra/postgres/` | Database repositories, connection pool, transactions |
| Config | `internal/infra/config/` | Environment-based configuration loading |
| Observability | `internal/infra/observability/` | Logging, tracing, metrics |

### Layer Import Rules (Critical)

| Layer | Can Import | CANNOT Import |
|-------|------------|---------------|
| **Domain** | stdlib only | `slog`, `uuid`, `pgx`, `otel`, ANY external |
| **App** | `domain` only | `slog`, `otel`, `uuid`, `net/http`, `pgx`, `transport`, `infra` |
| **Transport** | `domain`, `app`, `chi`, `uuid`, stdlib | `pgx`, `internal/infra` |
| **Infra** | `domain`, `pgx`, `slog`, `otel`, uuid, everything | `app`, `transport` |

> [!CAUTION]
> **Infra layer can ONLY import domain.** It cannot import app or transport layers. This prevents circular dependencies and maintains layer separation.

---

## Step 1: Define the Interface (Port)

The first step is defining the interface (port) that your adapter will implement.

### 1.1 Choose the Right Layer for Your Interface

| Adapter Type | Interface Location | Example |
|--------------|-------------------|---------|
| Data/Persistence adapters | `internal/domain/` | `UserRepository`, `Querier` |
| Auxiliary services | `internal/domain/` or `internal/shared/` | `IDGenerator`, `HTTPMetrics` |

### 1.2 Define Interface in Domain Layer

For data-related adapters (like repositories or caches), define the interface in the domain layer.

Create or edit `internal/domain/{adapter}.go`:

```go
package domain

import "context"

// CacheRepository defines the interface for caching operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
type CacheRepository interface {
    // Get retrieves a value by key. Returns ErrCacheNotFound if key doesn't exist.
    Get(ctx context.Context, key string) ([]byte, error)

    // Set stores a value with an optional TTL (0 means no expiration).
    Set(ctx context.Context, key string, value []byte, ttlSeconds int) error

    // Delete removes a key from the cache.
    Delete(ctx context.Context, key string) error
}
```

> [!TIP]
> Use `context.Context` as the first parameter for all interface methods. This enables request-scoped values like tracing and cancellation.

### 1.3 Add Domain Errors (if needed)

Add sentinel errors in `internal/domain/errors.go`:

```go
var (
    // ErrCacheNotFound is returned when a cache key doesn't exist.
    ErrCacheNotFound = errors.New("cache key not found")

    // ErrCacheFailed is returned when a cache operation fails.
    ErrCacheFailed = errors.New("cache operation failed")
)
```

### 1.4 Define Auxiliary Interfaces

For non-data adapters (like email services, SMS, external APIs), you may define interfaces in `internal/domain/` or `internal/shared/`:

```go
// EmailSender defines the interface for sending emails.
// Implementation in infra layer handles SMTP/API specifics.
type EmailSender interface {
    // Send sends an email. Returns error if delivery fails.
    Send(ctx context.Context, to, subject, body string) error
}
```

**Reference Files:**
- [internal/domain/user.go](../../internal/domain/user.go) - Repository interface
- [internal/domain/querier.go](../../internal/domain/querier.go) - Querier/TxManager interfaces

---

## Step 2: Implement the Adapter

Create the adapter implementation in the `internal/infra/` directory.

### 2.1 Create Adapter Directory

```
internal/infra/{adapter_name}/
├── {adapter}.go           # Main adapter implementation
├── {adapter}_test.go      # Integration tests
└── client.go              # Client initialization (if needed)
```

Example structure for a Redis adapter:

```
internal/infra/redis/
├── cache.go               # CacheRepository implementation
├── cache_test.go          # Integration tests
└── client.go              # Redis client initialization
```

### 2.2 Implement the Adapter Struct

Create `internal/infra/redis/cache.go`:

```go
// Package redis provides Redis-based cache implementations.
package redis

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// Cache implements domain.CacheRepository for Redis.
type Cache struct {
    client *redis.Client
}

// NewCache creates a new Cache instance.
func NewCache(client *redis.Client) *Cache {
    return &Cache{client: client}
}
```

### 2.3 Implement Interface Methods

Implement each method with proper error wrapping:

```go
// Get retrieves a value by key.
// Returns domain.ErrCacheNotFound if the key doesn't exist.
func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
    const op = "redisCache.Get"

    val, err := c.client.Get(ctx, key).Bytes()
    if errors.Is(err, redis.Nil) {
        return nil, fmt.Errorf("%s: %w", op, domain.ErrCacheNotFound)
    }
    if err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    return val, nil
}

// Set stores a value with an optional TTL.
func (c *Cache) Set(ctx context.Context, key string, value []byte, ttlSeconds int) error {
    const op = "redisCache.Set"

    expiration := time.Duration(ttlSeconds) * time.Second
    if ttlSeconds == 0 {
        expiration = 0 // No expiration
    }

    if err := c.client.Set(ctx, key, value, expiration).Err(); err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}

// Delete removes a key from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
    const op = "redisCache.Delete"

    if err := c.client.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}
```

### 2.4 Add Compile-Time Interface Check

Always add a compile-time check to ensure the adapter implements the interface:

```go
// Ensure Cache implements domain.CacheRepository at compile time.
var _ domain.CacheRepository = (*Cache)(nil)
```

> [!IMPORTANT]
> This line causes a compile error if `Cache` doesn't implement all methods of `domain.CacheRepository`. Always include this check.

### 2.5 Create Client Initialization

For adapters that need a connection/client, create a separate file:

Create `internal/infra/redis/client.go`:

```go
package redis

import (
    "context"
    "fmt"

    "github.com/redis/go-redis/v9"

    "github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

// NewClient creates a new Redis client from configuration.
func NewClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
    const op = "redis.NewClient"

    opts, err := redis.ParseURL(cfg.RedisURL)
    if err != nil {
        return nil, fmt.Errorf("%s: parse URL: %w", op, err)
    }

    client := redis.NewClient(opts)

    // Verify connection
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("%s: ping: %w", op, err)
    }

    return client, nil
}
```

### Error Wrapping Pattern

Always wrap errors with an operation string (`op`) for debugging:

```go
const op = "redisCache.Get"
// ...
return nil, fmt.Errorf("%s: %w", op, err)
```

This creates an error chain like: `redisCache.Get: cache key not found`

**Reference Files:**
- [internal/infra/postgres/user_repo.go](../../internal/infra/postgres/user_repo.go) - Repository adapter
- [internal/infra/postgres/pool.go](../../internal/infra/postgres/pool.go) - Connection pool
- [internal/infra/observability/logger.go](../../internal/infra/observability/logger.go) - Non-repository adapter

---

## Step 3: Configure via Environment Variables

All configuration is loaded from environment variables using `envconfig`.

### 3.1 Add Configuration Fields

Edit `internal/infra/config/config.go` to add new fields:

```go
type Config struct {
    // ... existing fields ...

    // Redis Configuration
    // RedisEnabled enables Redis cache. Default: false.
    RedisEnabled bool `envconfig:"REDIS_ENABLED" default:"false"`
    // RedisURL is the Redis connection URL (required if RedisEnabled=true).
    RedisURL string `envconfig:"REDIS_URL"`
    // RedisTTL is the default cache TTL in seconds. Default: 3600 (1 hour).
    RedisTTL int `envconfig:"REDIS_TTL" default:"3600"`
}
```

### 3.2 Available envconfig Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `envconfig:"VAR_NAME"` | Environment variable name | `envconfig:"REDIS_URL"` |
| `required:"true"` | Fail if not set | `required:"true"` |
| `default:"value"` | Default if not set | `default:"3600"` |

### 3.3 Add Validation

Add validation logic in the `Validate()` method:

```go
func (c *Config) Validate() error {
    // ... existing validations ...

    // Redis validation
    if c.RedisEnabled && strings.TrimSpace(c.RedisURL) == "" {
        return fmt.Errorf("REDIS_ENABLED is true but REDIS_URL is empty")
    }

    if c.RedisTTL < 0 {
        return fmt.Errorf("invalid REDIS_TTL: must be >= 0")
    }

    return nil
}
```

### 3.4 Update .env.example

Add documentation for new variables in `.env.example`:

```bash
# =============================================================================
# REDIS CONFIGURATION (optional)
# =============================================================================

# Enable Redis cache
REDIS_ENABLED=false

# Redis connection URL (required if REDIS_ENABLED=true)
# Format: redis://[:password@]host:port[/db]
REDIS_URL=redis://localhost:6379/0

# Default cache TTL in seconds (default: 3600 = 1 hour)
REDIS_TTL=3600
```

> [!WARNING]
> **No config files.** This project uses environment variables only. Never use config files like `config.yaml` or `config.json`.

### Configuration Pattern Examples

**Boolean toggles:**
```go
RedisEnabled bool `envconfig:"REDIS_ENABLED" default:"false"`
```

**Required secrets:**
```go
JWTSecret string `envconfig:"JWT_SECRET"` // Validated in Validate() when JWTEnabled=true
```

**URL/Connection strings:**
```go
DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
RedisURL    string `envconfig:"REDIS_URL"`
```

**Numeric settings:**
```go
Port         int   `envconfig:"PORT" default:"8080"`
RateLimitRPS int   `envconfig:"RATE_LIMIT_RPS" default:"100"`
MaxSize      int64 `envconfig:"MAX_REQUEST_SIZE" default:"1048576"`
```

**String enums:**
```go
LogLevel string `envconfig:"LOG_LEVEL" default:"info"` // Validated: debug, info, warn, error
```

**Reference File:** [internal/infra/config/config.go](../../internal/infra/config/config.go)

---

## Step 4: Write Tests

### 4.1 Unit Tests with Mocks

For testing code that uses your adapter, create a mock:

Create `internal/infra/redis/mock_cache.go` (or use testify/mockery):

```go
package redis

import (
    "context"

    "github.com/stretchr/testify/mock"
)

// MockCache is a mock implementation of domain.CacheRepository for testing.
type MockCache struct {
    mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
    args := m.Called(ctx, key)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttlSeconds int) error {
    args := m.Called(ctx, key, value, ttlSeconds)
    return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
    args := m.Called(ctx, key)
    return args.Error(0)
}
```

Usage in tests:

```go
func TestSomethingUsingCache(t *testing.T) {
    mockCache := new(redis.MockCache)
    mockCache.On("Get", mock.Anything, "user:123").Return([]byte(`{"name":"John"}`), nil)

    // Use mockCache in your test...

    mockCache.AssertExpectations(t)
}
```

### 4.2 Integration Tests with testcontainers-go

For testing the adapter itself against real infrastructure, use testcontainers-go.

Create `internal/infra/redis/cache_test.go`:

```go
//go:build integration

package redis_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    redisClient "github.com/redis/go-redis/v9"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/infra/redis"
)

func setupTestRedis(t *testing.T) (*redisClient.Client, func()) {
    t.Helper()
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "redis:7-alpine",
        ExposedPorts: []string{"6379/tcp"},
        WaitingFor:   wait.ForLog("Ready to accept connections"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)

    endpoint, err := container.Endpoint(ctx, "")
    require.NoError(t, err)

    client := redisClient.NewClient(&redisClient.Options{
        Addr: endpoint,
    })

    cleanup := func() {
        client.Close()
        container.Terminate(ctx)
    }

    return client, cleanup
}

func TestCache_GetSet(t *testing.T) {
    client, cleanup := setupTestRedis(t)
    defer cleanup()

    ctx := context.Background()
    cache := redis.NewCache(client)

    // Set a value
    err := cache.Set(ctx, "test-key", []byte("test-value"), 60)
    assert.NoError(t, err)

    // Get the value
    val, err := cache.Get(ctx, "test-key")
    assert.NoError(t, err)
    assert.Equal(t, []byte("test-value"), val)
}

func TestCache_Get_NotFound(t *testing.T) {
    client, cleanup := setupTestRedis(t)
    defer cleanup()

    ctx := context.Background()
    cache := redis.NewCache(client)

    val, err := cache.Get(ctx, "nonexistent-key")
    assert.Nil(t, val)
    assert.ErrorIs(t, err, domain.ErrCacheNotFound)
}

func TestCache_Delete(t *testing.T) {
    client, cleanup := setupTestRedis(t)
    defer cleanup()

    ctx := context.Background()
    cache := redis.NewCache(client)

    // Set and then delete
    err := cache.Set(ctx, "delete-key", []byte("value"), 60)
    require.NoError(t, err)

    err = cache.Delete(ctx, "delete-key")
    assert.NoError(t, err)

    // Verify deleted
    val, err := cache.Get(ctx, "delete-key")
    assert.Nil(t, val)
    assert.ErrorIs(t, err, domain.ErrCacheNotFound)
}
```

### Running Tests

```bash
# Run unit tests
make test

# Run integration tests (requires Docker)
make test-integration
# Or directly:
go test -tags=integration ./internal/infra/redis/...
```

> [!TIP]
> Use the `//go:build integration` build tag for integration tests. This prevents them from running during normal `make test`.

**Reference File:** [internal/infra/postgres/user_repo_test.go](../../internal/infra/postgres/user_repo_test.go)

---

## Step 5: Wire in main.go

### 5.1 Initialization Order

In `cmd/api/main.go`, initialize components in this order:

1. **Configuration** - Load environment variables
2. **Observability** - Logger, tracer, metrics
3. **Infrastructure Adapters** - Database, cache, external clients
4. **Repositories** - Using infrastructure adapters
5. **Use Cases** - Business logic with injected dependencies
6. **Handlers** - HTTP handlers with use cases
7. **Router** - Wire everything together
8. **Server** - Start HTTP server

### 5.2 Add Adapter Initialization

Add Redis initialization in `run()` function:

```go
func run() error {
    ctx := context.Background()

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("failed to load configuration: %w", err)
    }

    // Initialize logger
    logger := observability.NewLogger(cfg)
    slog.SetDefault(logger)

    // ... existing initialization ...

    // Initialize Redis (if enabled)
    var cacheRepo domain.CacheRepository
    var redisClient *redis.Client
    if cfg.RedisEnabled {
        redisClient, err = redisAdapter.NewClient(ctx, cfg)
        if err != nil {
            return fmt.Errorf("failed to connect to Redis: %w", err)
        }
        defer redisClient.Close()

        cacheRepo = redisAdapter.NewCache(redisClient)
        logger.Info("redis connected")
    } else {
        logger.Info("redis disabled; cache operations will be no-op")
        // Optionally: use a no-op cache implementation
    }

    // ... rest of initialization ...
}
```

### 5.3 Dependency Injection Pattern

Pass adapters to use cases through constructors:

```go
// Create use cases with injected dependencies
createUserUC := user.NewCreateUserUseCase(
    userRepo,
    auditService,
    idGen,
    txManager,
    querier,
    cacheRepo,  // Add new dependency
)
```

### 5.4 Graceful Shutdown

For adapters with resources (connections, pools), ensure proper cleanup:

```go
// In run() function
defer func() {
    // Close Redis connection
    if redisClient != nil {
        if err := redisClient.Close(); err != nil {
            logger.Error("redis close failed", slog.Any("err", err))
        }
    }

    // Close database pool
    db.Close()

    // Shutdown tracer
    if tpShutdown != nil {
        if err := tpShutdown(ctx); err != nil {
            logger.Error("tracer shutdown failed", slog.Any("err", err))
        }
    }
}()
```

### 5.5 Conditional Initialization

Use feature flags for optional adapters:

```go
if cfg.RedisEnabled {
    // Initialize Redis
} else {
    // Use no-op implementation or nil
}
```

**Reference File:** [cmd/api/main.go](../../cmd/api/main.go)

---

## Example: Redis Cache Adapter

### Complete Implementation Checklist

1. **Domain Interface** - `internal/domain/cache.go`
2. **Domain Errors** - `internal/domain/errors.go`
3. **Adapter Implementation** - `internal/infra/redis/cache.go`
4. **Client Initialization** - `internal/infra/redis/client.go`
5. **Configuration** - `internal/infra/config/config.go`
6. **Environment Variables** - `.env.example`
7. **Integration Tests** - `internal/infra/redis/cache_test.go`
8. **Wiring** - `cmd/api/main.go`

### Files to Create/Modify

| Action | File | Purpose |
|--------|------|---------|
| CREATE | `internal/domain/cache.go` | CacheRepository interface |
| MODIFY | `internal/domain/errors.go` | Add ErrCacheNotFound |
| CREATE | `internal/infra/redis/cache.go` | Redis implementation |
| CREATE | `internal/infra/redis/client.go` | Redis client setup |
| CREATE | `internal/infra/redis/cache_test.go` | Integration tests |
| MODIFY | `internal/infra/config/config.go` | Add Redis config fields |
| MODIFY | `.env.example` | Document Redis variables |
| MODIFY | `cmd/api/main.go` | Initialize and wire Redis |

### Directory Structure

```
internal/
├── domain/
│   ├── cache.go          # [NEW] CacheRepository interface
│   └── errors.go         # [MODIFY] Add cache errors
└── infra/
    └── redis/            # [NEW] Redis adapter package
        ├── cache.go
        ├── cache_test.go
        └── client.go
```

---

## Example: Email Service Adapter

### Domain Interface

```go
// internal/domain/email.go
package domain

import "context"

// EmailSender defines the interface for sending emails.
type EmailSender interface {
    // Send sends an email to the specified recipient.
    Send(ctx context.Context, to, subject, htmlBody string) error

    // SendWithTemplate sends an email using a named template.
    SendWithTemplate(ctx context.Context, to, templateName string, data map[string]any) error
}
```

### Adapter Implementation

```go
// internal/infra/email/sender.go
package email

import (
    "context"
    "fmt"
    "net/smtp"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

// SMTPSender implements domain.EmailSender using SMTP.
type SMTPSender struct {
    host     string
    port     int
    username string
    password string
    from     string
}

// NewSMTPSender creates a new SMTP email sender.
func NewSMTPSender(cfg *config.Config) *SMTPSender {
    return &SMTPSender{
        host:     cfg.SMTPHost,
        port:     cfg.SMTPPort,
        username: cfg.SMTPUsername,
        password: cfg.SMTPPassword,
        from:     cfg.SMTPFrom,
    }
}

// Send sends an email via SMTP.
func (s *SMTPSender) Send(ctx context.Context, to, subject, htmlBody string) error {
    const op = "emailSender.Send"

    addr := fmt.Sprintf("%s:%d", s.host, s.port)
    auth := smtp.PlainAuth("", s.username, s.password, s.host)

    msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
        s.from, to, subject, htmlBody)

    if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}

// SendWithTemplate sends an email using a template (simplified example).
func (s *SMTPSender) SendWithTemplate(ctx context.Context, to, templateName string, data map[string]any) error {
    const op = "emailSender.SendWithTemplate"
    // Template rendering implementation...
    return nil
}

// Ensure SMTPSender implements domain.EmailSender at compile time.
var _ domain.EmailSender = (*SMTPSender)(nil)
```

### Configuration

```go
// Add to internal/infra/config/config.go
type Config struct {
    // ... existing fields ...

    // Email/SMTP Configuration
    EmailEnabled bool   `envconfig:"EMAIL_ENABLED" default:"false"`
    SMTPHost     string `envconfig:"SMTP_HOST"`
    SMTPPort     int    `envconfig:"SMTP_PORT" default:"587"`
    SMTPUsername string `envconfig:"SMTP_USERNAME"`
    SMTPPassword string `envconfig:"SMTP_PASSWORD"`
    SMTPFrom     string `envconfig:"SMTP_FROM"`
}
```

### Environment Variables

```bash
# .env.example
# =============================================================================
# EMAIL/SMTP CONFIGURATION (optional)
# =============================================================================

# Enable email sending
EMAIL_ENABLED=false

# SMTP server configuration
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@example.com
```

---

## Quick Reference Checklist

### Files to Create/Modify for a New Adapter

| Step | File | Action | Description |
|------|------|--------|-------------|
| 1 | `internal/domain/{adapter}.go` | CREATE | Define interface (port) |
| 2 | `internal/domain/errors.go` | MODIFY | Add domain errors |
| 3 | `internal/infra/{adapter}/` | CREATE | Create adapter directory |
| 4 | `internal/infra/{adapter}/{adapter}.go` | CREATE | Implement adapter |
| 5 | `internal/infra/{adapter}/client.go` | CREATE | Client initialization (if needed) |
| 6 | `internal/infra/config/config.go` | MODIFY | Add config fields and validation |
| 7 | `.env.example` | MODIFY | Document environment variables |
| 8 | `internal/infra/{adapter}/{adapter}_test.go` | CREATE | Write tests |
| 9 | `cmd/api/main.go` | MODIFY | Initialize and wire adapter |

### Implementation Checklist

- [ ] Interface defined in domain layer with `context.Context` first parameter
- [ ] Domain errors added (if needed)
- [ ] Adapter struct implements the interface
- [ ] Constructor follows `NewXAdapter(deps...) *XAdapter` pattern
- [ ] All errors wrapped with `op` string: `fmt.Errorf("%s: %w", op, err)`
- [ ] Compile-time interface check: `var _ domain.XInterface = (*XAdapter)(nil)`
- [ ] Configuration fields added with envconfig tags
- [ ] Config validation added for required/conditional fields
- [ ] `.env.example` updated with documentation
- [ ] Integration tests written with testcontainers
- [ ] Adapter initialized in main.go with proper order
- [ ] Graceful shutdown implemented for adapters with resources

### Common Pitfalls

> [!WARNING]
> **Common Mistakes to Avoid:**

| Pitfall | Correct Approach |
|---------|-----------------|
| Importing `app` or `transport` in infra | Infra can only import `domain` |
| Not wrapping errors with `op` | Always use `fmt.Errorf("%s: %w", op, err)` |
| Missing compile-time interface check | Add `var _ Interface = (*Adapter)(nil)` |
| Hardcoding configuration values | Use environment variables via envconfig |
| Not validating required config when feature enabled | Add conditional validation in `Validate()` |
| Forgetting graceful shutdown | Always close connections/pools on shutdown |

---

## References

### Existing Adapter Files

| File | Type | Purpose |
|------|------|---------|
| [internal/domain/user.go](../../internal/domain/user.go) | Port | Repository interface example |
| [internal/domain/querier.go](../../internal/domain/querier.go) | Port | Querier/TxManager interfaces |
| [internal/infra/postgres/user_repo.go](../../internal/infra/postgres/user_repo.go) | Adapter | PostgreSQL repository |
| [internal/infra/postgres/pool.go](../../internal/infra/postgres/pool.go) | Adapter | Connection pool |
| [internal/infra/postgres/querier.go](../../internal/infra/postgres/querier.go) | Adapter | Querier implementation |
| [internal/infra/postgres/tx_manager.go](../../internal/infra/postgres/tx_manager.go) | Adapter | Transaction manager |
| [internal/infra/config/config.go](../../internal/infra/config/config.go) | Adapter | Configuration loading |
| [internal/infra/observability/logger.go](../../internal/infra/observability/logger.go) | Adapter | Structured logging |
| [internal/infra/observability/tracer.go](../../internal/infra/observability/tracer.go) | Adapter | OpenTelemetry tracing |
| [internal/infra/observability/metrics.go](../../internal/infra/observability/metrics.go) | Adapter | Prometheus metrics |
| [internal/infra/postgres/user_repo_test.go](../../internal/infra/postgres/user_repo_test.go) | Test | Integration test example |
| [cmd/api/main.go](../../cmd/api/main.go) | Wiring | Dependency injection |

### Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture.md](../architecture.md) | Hexagonal architecture details |
| [docs/project-context.md](../project-context.md) | Layer rules and conventions |
| [docs/guides/adding-module.md](./adding-module.md) | Guide for adding new modules |
