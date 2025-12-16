# Story 2.1: Implement Structured JSON Logging

Status: done

## Story

**As a** developer,
**I want** structured JSON logs with consistent fields,
**So that** I can parse and search logs easily.

## Acceptance Criteria

1. **Given** the service is running
   **When** any log entry is written
   **Then** log is emitted in JSON format to stdout
   **And** each log entry includes required fields:
   - `time` (RFC 3339 format)
   - `level` (debug/info/warn/error)
   - `msg` (log message)
   - `service` (from config SERVICE_NAME)
   - `env` (from config ENV)
   **And** request-scoped logs include `requestId`

2. **Given** `LOG_LEVEL` is set to `warn`
   **When** info-level log is written
   **Then** log is NOT emitted (filtered)

3. **Given** any log operation
   **When** log entry is written
   **Then** sensitive data is NEVER logged (Authorization header, password, token, secret fields)

4. **Given** any HTTP request is processed
   **When** request completes (success or error)
   **Then** request logging middleware emits log entry with:
   - `method` (HTTP method)
   - `route` (Chi route pattern)
   - `status` (HTTP status code)
   - `duration_ms` (request processing time)
   - `bytes` (response size)

*Covers: FR12*

## Tasks / Subtasks

- [x] Task 1: Create logger initialization (AC: #1, #2)
  - [x] Create `internal/infra/observability/logger.go`
  - [x] Configure slog with JSON handler for stdout
  - [x] Add default attributes: `service`, `env`
  - [x] Implement log level filtering from `LOG_LEVEL` config
  - [x] Create `NewLogger(cfg *config.Config) *slog.Logger` function

- [x] Task 2: Extend Config for logging (AC: #1, #2)
  - [x] Add `LogLevel` to Config struct if not present (already exists from Story 1.2)
  - [x] Validate LogLevel parsing (debug/info/warn/error)

- [x] Task 3: Create request logging middleware (AC: #4)
  - [x] Create `internal/transport/http/middleware/logging.go`
  - [x] Implement middleware that logs request completion
  - [x] Include: method, route (Chi pattern), status, duration_ms, bytes
  - [x] Extract route pattern using `chi.RouteContext(r.Context()).RoutePattern()`

- [x] Task 4: Create response writer wrapper (AC: #4)
  - [x] Create wrapper that captures status code and bytes written
  - [x] Implement `http.ResponseWriter` interface
  - [x] Track bytes written via `Write()` method

- [x] Task 5: Wire logger into main.go (AC: #1)
  - [x] Initialize logger in main.go
  - [x] Pass logger to middleware and handlers
  - [x] Set as default logger: `slog.SetDefault(logger)`

- [x] Task 6: Write unit tests (AC: #1, #2, #3, #4)
  - [x] Test JSON output format verification
  - [x] Test log level filtering
  - [x] Test required fields presence
  - [x] Test middleware captures correct request info

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

Logger package is in **Infra layer** (`internal/infra/observability/`):
- ✅ ALLOWED: domain, slog, otel, external packages
- ❌ FORBIDDEN: app, transport imports

### Technology Stack [Source: docs/project-context.md]

| Component | Package | Version |
|-----------|---------|---------|
| Logging | log/slog (stdlib) | Go 1.21+ |

### Logger Pattern [Source: docs/architecture.md]

```go
// internal/infra/observability/logger.go
package observability

import (
    "log/slog"
    "os"
    
    "github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

func NewLogger(cfg *config.Config) *slog.Logger {
    level := parseLogLevel(cfg.LogLevel)
    
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: level,
    })
    
    // Add default attributes
    logger := slog.New(handler).With(
        "service", cfg.ServiceName,
        "env", cfg.Env,
    )
    
    return logger
}

func parseLogLevel(level string) slog.Level {
    switch level {
    case "debug":
        return slog.LevelDebug
    case "info":
        return slog.LevelInfo
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
```

### Request Logging Middleware Pattern [Source: docs/architecture.md]

```go
// internal/transport/http/middleware/logging.go
package middleware

import (
    "log/slog"
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap response writer to capture status and bytes
            ww := &responseWrapper{ResponseWriter: w, status: http.StatusOK}
            
            next.ServeHTTP(ww, r)
            
            // Get route pattern from chi
            routePattern := chi.RouteContext(r.Context()).RoutePattern()
            if routePattern == "" {
                routePattern = r.URL.Path
            }
            
            logger.Info("request completed",
                "method", r.Method,
                "route", routePattern,
                "status", ww.status,
                "duration_ms", time.Since(start).Milliseconds(),
                "bytes", ww.bytes,
                "requestId", GetRequestID(r.Context()), // Will be added in Story 2.2
            )
        })
    }
}

type responseWrapper struct {
    http.ResponseWriter
    status int
    bytes  int
}

func (w *responseWrapper) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
    n, err := w.ResponseWriter.Write(b)
    w.bytes += n
    return n, err
}
```

### Required Log Fields [Source: docs/project-context.md]

| Field | Source | Always Present |
|-------|--------|----------------|
| `time` | slog default | Yes |
| `level` | slog default | Yes |
| `msg` | log message | Yes |
| `service` | Config.ServiceName | Yes |
| `env` | Config.Env | Yes |
| `requestId` | Context (Story 2.2) | Request-scoped only |
| `traceId` | Context (Story 2.3b) | When tracing enabled |

### Sensitive Data Protection [Source: docs/project-context.md]

NEVER log:
- `Authorization` header
- Passwords
- Tokens (JWT, API keys)
- Secret fields
- Full request bodies with sensitive data

### Previous Story Learnings [Source: Epic 1]

- Config system uses `kelseyhightower/envconfig` (Story 1.2)
- LOG_LEVEL validation already exists: debug/info/warn/error allowlist
- Chi router already initialized (Story 1.5)
- Use existing patterns for middleware composition

## Technical Requirements

- **Go version:** 1.23+ [Source: docs/project-context.md]
- **Logger:** log/slog (stdlib) with JSONHandler
- **Test coverage:** Unit tests for logger and middleware

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- slog is stdlib - NO external logging dependency
- JSON format to stdout for production
- Log level filterable via environment variable
- NEVER log sensitive data

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-16)

### Agent Model Used

Gemini (via Antigravity)

### Debug Log References

No issues encountered.

### Completion Notes List

- Created `logger.go` with `NewLogger(cfg)` function returning JSON slog logger with service/env default attributes
- Implemented log level filtering supporting debug/info/warn/error with info as default
- Created `response_wrapper.go` implementing `http.ResponseWriter` to capture status and bytes
- Created `logging.go` middleware using Chi's route pattern for structured request logging
- Updated `router.go` to accept logger and wire RequestLogger middleware
- Updated `main.go` to use `observability.NewLogger` and `slog.SetDefault(logger)`
- Removed duplicate `newLogger` function from main.go
- Updated integration tests to pass logger to NewRouter
- All unit tests passing (4 observability, 11 middleware, 5 handler tests)
- All lint errors fixed in new code

### File List

**New Files:**
- `internal/infra/observability/logger.go`
- `internal/infra/observability/logger_test.go`
- `internal/transport/http/middleware/logging.go`
- `internal/transport/http/middleware/logging_test.go`
- `internal/transport/http/middleware/response_wrapper.go`
- `internal/transport/http/middleware/response_wrapper_test.go`

**Modified Files:**
- `cmd/api/main.go`
- `internal/transport/http/router.go`
- `internal/transport/http/handler/integration_test.go`
- `docs/sprint-artifacts/epic-1-retro-2025-12-16.md`
- `docs/sprint-artifacts/sprint-status.yaml`
- `docs/sprint-artifacts/2-1-implement-structured-json-logging.md` (story status update)

### Change Log

- 2025-12-16: Implemented structured JSON logging with middleware (Story 2.1)
