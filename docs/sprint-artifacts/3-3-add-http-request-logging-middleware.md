# Story 3.3: Add HTTP Request Logging Middleware

Status: done

## Story

As a SRE,
I want all HTTP requests logged with structured fields,
So that I can monitor and debug traffic.

## Acceptance Criteria

### AC1: Log contains required fields
**Given** any HTTP request is made
**When** the request completes
**Then** log entry contains: method, path, status, latency_ms, request_id

### AC2: JSON format in production
**Given** `APP_ENV=production`
**When** request is logged
**Then** output is JSON format

---

## Tasks / Subtasks

- [x] **Task 1: Add zap dependency** (AC: #1, #2)
  - [x] Run `go get go.uber.org/zap` ✅ v1.27.1
  - [x] Verify zap is added to go.mod ✅

- [x] **Task 2: Create logger initialization** (AC: #2)
  - [x] Create `internal/observability/logger.go` ✅
  - [x] Implement `NewLogger(cfg *config.LogConfig) (*zap.Logger, error)` ✅
  - [x] Use JSON encoder when APP_ENV=production ✅
  - [x] Use console encoder for development ✅
  - [x] Configure log level from cfg.Log.Level ✅

- [x] **Task 3: Create logging middleware** (AC: #1, #2)
  - [x] Create `internal/interface/http/middleware/logging.go` ✅
  - [x] Implement `Logging(logger *zap.Logger) func(next http.Handler) http.Handler` ✅
  - [x] Log: method, path, status, latency_ms, request_id ✅
  - [x] Use middleware.GetRequestID for request_id field ✅
  - [x] Calculate latency using time.Since ✅

- [x] **Task 4: Wire middleware into router** (AC: #1, #2)
  - [x] Update `NewRouter` to initialize logger ✅
  - [x] Add `r.Use(middleware.Logging(logger))` after RequestID ✅
  - [x] Wire cfg.Log and cfg.App.Env into router ✅
  - [x] Add fallback to NopLogger on error ✅

- [x] **Task 5: Create middleware tests** (AC: #1, #2)
  - [x] Create `internal/interface/http/middleware/logging_test.go` ✅
  - [x] Test: log contains all required fields ✅
  - [x] Test: latency is measured correctly ✅
  - [x] Test: request_id is included ✅
  - [x] Test: captures non-200 status ✅
  - [x] Test: works with nop logger ✅

- [x] **Task 6: Create logger tests** (AC: #2)
  - [x] Create `internal/observability/logger_test.go` ✅
  - [x] Test: production config uses JSON encoder ✅
  - [x] Test: development config uses console encoder ✅
  - [x] Test: staging uses production config ✅
  - [x] Test: invalid level defaults to info ✅

- [x] **Task 7: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass ✅ (100% coverage)
  - [x] Run `make lint` - 0 issues ✅
  - [x] Fixed router_test.go to use testConfig() ✅

---

## Dev Notes

### Logger Initialization Pattern

```go
// internal/observability/logger.go
package observability

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "github.com/iruldev/golang-api-hexagonal/internal/config"
)

// NewLogger creates a new zap logger based on configuration.
func NewLogger(cfg *config.LogConfig, appEnv string) (*zap.Logger, error) {
    var zapConfig zap.Config

    if appEnv == "production" || appEnv == "staging" {
        zapConfig = zap.NewProductionConfig()
    } else {
        zapConfig = zap.NewDevelopmentConfig()
    }

    // Override format if specified
    if cfg.Format == "json" {
        zapConfig.Encoding = "json"
    } else if cfg.Format == "console" {
        zapConfig.Encoding = "console"
    }

    // Set log level
    level, err := zapcore.ParseLevel(cfg.Level)
    if err != nil {
        level = zapcore.InfoLevel
    }
    zapConfig.Level = zap.NewAtomicLevelAt(level)

    return zapConfig.Build()
}
```

### Logging Middleware Pattern

```go
// internal/interface/http/middleware/logging.go
package middleware

import (
    "net/http"
    "time"

    "go.uber.org/zap"
)

// Logging middleware logs HTTP requests with structured fields.
func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Wrap response writer to capture status code
            ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

            next.ServeHTTP(ww, r)

            latency := time.Since(start)

            logger.Info("request",
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.Int("status", ww.statusCode),
                zap.Duration("latency_ms", latency),
                zap.String("request_id", GetRequestID(r.Context())),
            )
        })
    }
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### Router Integration

```go
// internal/interface/http/router.go
func NewRouter(cfg *config.Config) chi.Router {
    logger, _ := observability.NewLogger(&cfg.Log, cfg.App.Env)

    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.Logging(logger))

    // API v1 routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/health", handlers.HealthHandler)
    })

    return r
}
```

### Architecture Compliance

**Layer:** `internal/observability` and `internal/interface/http/middleware`
**Pattern:** Structured logging with zap
**Dependency:** zap is widely used in production Go projects

### Previous Story Learnings

From **Story 3.2** code review:
- ✅ Perfect implementation with 0 issues
- ✅ Maintain 100% coverage pattern
- ✅ Use GetRequestID from context

### Dependencies

**New:** `go.uber.org/zap`
**Existing:** middleware.GetRequestID (from Story 3.2)

### File Structure After Implementation

```
internal/
├── observability/
│   ├── doc.go
│   ├── logger.go           # Logger initialization (NEW)
│   └── logger_test.go      # Logger tests (NEW)
└── interface/http/
    ├── middleware/
    │   ├── requestid.go
    │   ├── requestid_test.go
    │   ├── logging.go      # Logging middleware (NEW)
    │   └── logging_test.go # Logging tests (NEW)
    └── router.go           # Updated
```

### Configuration Reference

From `internal/config/config.go`:
```go
type LogConfig struct {
    Level  string `koanf:"level"`   // debug, info, warn, error
    Format string `koanf:"format"`  // json, console
}
```

From `.env.example`:
```
LOG_LEVEL=info
LOG_FORMAT=console
APP_ENV=development
```

### References

- [Source: docs/epics.md#Story-3.3]
- [zap documentation](https://github.com/uber-go/zap)
- [Story 3.2 - Request ID Middleware](file:///docs/sprint-artifacts/3-2-implement-request-id-middleware.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Third story in Epic 3: HTTP API Core.
Adds structured logging to HTTP requests.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11
- Implementation completed: 2025-12-11
  - Added zap v1.27.1 and multierr v1.10.0 dependencies
  - Created logger.go (44 lines) with NewLogger and NewNopLogger
  - Created logger_test.go (62 lines) with 5 test cases
  - Created logging.go (45 lines) with responseWriter wrapper
  - Created logging_test.go (118 lines) with 5 test cases
  - Wired middleware into router.go with fallback to NopLogger on error
  - Fixed router_test.go to use testConfig() instead of nil
  - Fixed lint issue: use switch instead of if-else chain
  - Coverage: 100% (middleware, observability), Lint: 0 issues

### File List

Files created:
- `internal/observability/logger.go` - Logger initialization (44 lines)
- `internal/observability/logger_test.go` - Logger tests (62 lines)
- `internal/interface/http/middleware/logging.go` - Logging middleware (45 lines)
- `internal/interface/http/middleware/logging_test.go` - Middleware tests (118 lines)

Files modified:
- `go.mod` - Added zap and multierr dependencies
- `go.sum` - Updated with zap dependencies
- `internal/interface/http/router.go` - Wired logging middleware (40 lines)
- `internal/interface/http/router_test.go` - Added testConfig helper (77 lines)
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
