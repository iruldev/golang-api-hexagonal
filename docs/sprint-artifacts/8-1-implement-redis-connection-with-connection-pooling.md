# Story 8.1: Implement Redis Connection with Connection Pooling

Status: done

## Story

As a developer,
I want the system to connect to Redis with proper connection pooling,
So that I can use Redis for caching and job queues.

## Acceptance Criteria

### AC1: Redis connection established
**Given** valid `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` environment variables
**When** the application starts
**Then** Redis connection pool is established
**And** connection is validated at startup
**And** `/readyz` includes Redis health check

---

## Tasks / Subtasks

- [x] **Task 1: Add Redis config** (AC: #1)
  - [x] RedisConfig struct added to config.go
  - [x] REDIS_* env vars added to .env.example
  - [x] REDIS_ prefix added to loader
  - [x] Redis config validation (validateRedis in validate.go)

- [x] **Task 2: Create Redis connection** (AC: #1)
  - [x] Created `internal/infra/redis/redis.go`
  - [x] Connection pool with go-redis v9
  - [x] Ping validation at connection time
  - [x] go-redis v9 dependency added to go.mod

- [x] **Task 3: Add Redis health check** (AC: #1)
  - [x] Client implements DBHealthChecker interface
  - [x] ReadyzHandler.WithRedis() method added
  - [x] Returns 503 when Redis unavailable
  - [x] Redis wired to RouterDeps in main.go
  - [x] Graceful shutdown with defer Close()

- [x] **Task 4: Add docker-compose service** (AC: #1)
  - [x] Redis 7-alpine service added
  - [x] Port 6379 exposed
  - [x] Health check configured

- [x] **Task 5: Unit tests** (AC: #1)
  - [x] Created `internal/infra/redis/redis_test.go`

---

## Dev Notes

### go-redis Setup

```go
import "github.com/redis/go-redis/v9"

type Client struct {
    rdb *redis.Client
}

func NewClient(cfg Config) (*Client, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
        PoolSize: cfg.PoolSize,
    })
    
    // Validate connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := rdb.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis connection failed: %w", err)
    }
    
    return &Client{rdb: rdb}, nil
}
```

### Config

```go
type RedisConfig struct {
    Host     string `koanf:"host" default:"localhost"`
    Port     int    `koanf:"port" default:"6379"`
    Password string `koanf:"password"`
    DB       int    `koanf:"db" default:"0"`
    PoolSize int    `koanf:"pool_size" default:"10"`
}
```

### File List

Files created:
- `internal/infra/redis/redis.go` - Redis client with connection pooling
- `internal/infra/redis/redis_test.go` - Unit tests for Redis client

Files modified:
- `go.mod` - Added go-redis v9 dependency
- `internal/config/config.go` - Add RedisConfig struct
- `internal/config/validate.go` - Add validateRedis() function
- `internal/config/loader.go` - Add REDIS_ prefix mapping
- `.env.example` - Add Redis env vars
- `docker-compose.yaml` - Add Redis service
- `internal/interface/http/handlers/health.go` - Add WithRedis() method
- `internal/interface/http/router.go` - Add RedisChecker to RouterDeps, call WithRedis()
- `cmd/server/main.go` - Create Redis client, wire to router, add graceful shutdown

---

## Code Review Fixes Applied

Fixes from adversarial code review (2025-12-12):

1. ✅ **go-redis dependency** - Added to go.mod via `go get`
2. ✅ **Redis client wired in main.go** - Creates and validates connection at startup
3. ✅ **Graceful shutdown** - `defer redisClient.Close()` for proper cleanup
4. ✅ **RouterDeps extended** - Added RedisChecker field
5. ✅ **WithRedis() called** - /readyz now checks Redis health
6. ✅ **Config validation** - validateRedis() added for fail-fast on bad config
7. ✅ **Unit tests created** - redis_test.go with coverage for defaults
8. ✅ **Duplicate interface removed** - Uses handlers.DBHealthChecker

