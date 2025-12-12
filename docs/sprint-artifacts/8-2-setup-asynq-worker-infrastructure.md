# Story 8.2: Setup Asynq Worker Infrastructure

Status: done

## Story

As a developer,
I want asynq worker infrastructure ready to use,
So that I can process background jobs reliably.

## Acceptance Criteria

### AC1: Worker Package Created
**Given** `internal/worker/` package exists
**When** I review the code
**Then** asynq client and server are configured
**And** task handler registration pattern is implemented
**And** package follows Hexagonal Architecture

### AC2: Worker Starts with Graceful Shutdown
**Given** worker is running
**When** SIGTERM or SIGINT is received
**Then** worker stops accepting new tasks
**And** existing tasks have up to 30 seconds to complete
**And** worker exits gracefully

### AC3: Configuration Support
**Given** valid `ASYNQ_*` environment variables
**When** the worker starts
**Then** concurrency, queue priorities, and retry settings are applied
**And** Redis connection reuses existing Redis config

---

## Tasks / Subtasks

- [x] **Task 1: Add Asynq config** (AC: #3)
  - [x] Add `AsynqConfig` struct to `config.go`
  - [x] Add `Asynq` field to main `Config` struct
  - [x] Add `"ASYNQ_": "asynq"` to `envPrefixes` map in `loader.go`
  - [x] Add `validateAsynq()` to `validate.go`
  - [x] Add ASYNQ_* vars to `.env.example`

- [x] **Task 2: Install asynq dependency** (AC: #1)
  - [x] `go get github.com/hibiken/asynq@latest` (v0.25.1)

- [x] **Task 3: Create worker server** (AC: #1, #2)
  - [x] Create `internal/worker/server.go`
  - [x] Priority queues: `critical:6`, `default:3`, `low:1`
  - [x] Graceful shutdown with configurable timeout

- [x] **Task 4: Create asynq client** (AC: #1)
  - [x] Create `internal/worker/client.go`
  - [x] Queue helpers: `EnqueueCritical()`, `EnqueueDefault()`, `EnqueueLow()`

- [x] **Task 5: Create middleware** (AC: #1)
  - [x] Create `internal/worker/middleware.go`
  - [x] Logging middleware with zap (not log package)
  - [x] Recovery middleware
  - [x] OTEL tracing middleware

- [x] **Task 6: Create worker entry point** (AC: #2)
  - [x] Create `cmd/worker/main.go`
  - [x] Add Makefile target: `make worker`

- [x] **Task 7: Unit tests** (AC: #1, #2, #3)
  - [x] Create `internal/worker/server_test.go`
  - [x] Create `internal/worker/client_test.go`
  - [x] Create `internal/worker/middleware_test.go`

---

## Dev Notes

### Architecture Placement

```
cmd/worker/main.go           # Entry point
internal/worker/
├── server.go                # Asynq server wrapper
├── client.go                # Asynq client wrapper  
├── middleware.go            # Logging, recovery, tracing
└── tasks/                   # Task handlers (Story 8.3)
```

---

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro

### Completion Notes
- Worker is separate process from HTTP server
- Middleware uses zap logger for consistency with HTTP middleware
- OTEL tracing enabled for distributed tracing across HTTP and worker
- Asynq manages its own Redis connection pool (reuses same config)
- All 14 worker tests pass
- All 29 config tests pass (no regressions from Asynq validation)
- Build compiles successfully

### File List

**Created:**
- `internal/worker/server.go`
- `internal/worker/client.go`
- `internal/worker/middleware.go`
- `internal/worker/server_test.go`
- `internal/worker/client_test.go`
- `internal/worker/middleware_test.go`
- `cmd/worker/main.go`

**Modified:**
- `internal/config/config.go` - Added AsynqConfig struct and Asynq field
- `internal/config/loader.go` - Added ASYNQ_ prefix to envPrefixes
- `internal/config/validate.go` - Added validateAsynq() function
- `.env.example` - Added Asynq configuration section
- `Makefile` - Added worker target
- `go.mod` - Added asynq v0.25.1 dependency

### Change Log
- 2025-12-12: Implemented asynq worker infrastructure with server, client, middleware, entry point, and unit tests
- 2025-12-12: Code review fixes - replaced magic number with time.Second, added OTEL span status reporting

---

## Senior Developer Review (AI)

**Review Date:** 2025-12-12
**Outcome:** Approved with fixes applied

### Issues Found: 1 High, 3 Medium, 2 Low

### Action Items
- [x] [HIGH] server.go:31 - Replace magic number `1e9` with `time.Second`
- [x] [MEDIUM] middleware.go:82-85 - Add `span.SetStatus()` for OTEL error classification
- [ ] [MEDIUM] Config defaults in infrastructure layer - Acceptable for now (pattern consistent)
- [ ] [MEDIUM] No integration tests - Deferred to Story 8.5 (testcontainers)
- [ ] [LOW] Unused ctx parameter in client Enqueue methods - Asynq API design
- [ ] [LOW] Server tests don't verify queue weights - Minor coverage gap
