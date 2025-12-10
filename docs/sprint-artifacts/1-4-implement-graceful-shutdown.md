# Story 1.4: Implement Graceful Shutdown

Status: done

## Story

As a SRE,
I want the service to shutdown gracefully,
So that in-flight requests complete before termination.

## Acceptance Criteria

### AC1: Signal handling stops new connections ✅
**Given** the application is running with active requests
**When** SIGTERM or SIGINT is received
**Then** the server stops accepting new connections

### AC2: In-flight requests complete with timeout ✅
**Given** the server received shutdown signal
**When** existing requests are still processing
**Then** existing requests have up to 30 seconds to complete

### AC3: Clean exit on successful shutdown ✅
**Given** all in-flight requests completed (or timeout reached)
**When** shutdown sequence finishes
**Then** the application exits with code 0

---

## Tasks / Subtasks

- [x] **Task 1: Create shutdown handler** (AC: #1, #2, #3)
  - [x] Create `internal/app/shutdown.go` with graceful shutdown logic
  - [x] Register signal handlers for SIGTERM and SIGINT
  - [x] Implement 30-second timeout context
  - [x] Call `http.Server.Shutdown(ctx)` for graceful stop

- [x] **Task 2: Update main.go** (AC: #1, #2, #3)
  - [x] Create HTTP server with configurable address
  - [x] Start server in goroutine
  - [x] Wait for shutdown signal or server error
  - [x] Log shutdown events (starting, completed, timeout)
  - [x] Update outdated comment (Story 1.2+ → Story 1.4)

- [x] **Task 3: Create shutdown test** (AC: #1, #2, #3)
  - [x] Create `internal/app/shutdown_test.go`
  - [x] Test signal handling triggers shutdown
  - [x] Test graceful wait for in-flight requests
  - [x] Test timeout behavior after 30 seconds

- [x] **Task 4: Verify shutdown behavior** (AC: #1, #2, #3)
  - [x] Run `make test` - all tests pass (77.8% coverage)
  - [x] Run `make lint` - 0 issues
  - [x] Verified signal handling works

---

## Dev Notes

### Go Graceful Shutdown Pattern

```go
// internal/app/shutdown.go
package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const ShutdownTimeout = 30 * time.Second

// GracefulShutdown handles OS signals and shuts down the server gracefully.
func GracefulShutdown(server *http.Server, done chan<- error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	<-quit // Block until signal received
	
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		done <- err
		return
	}
	done <- nil
}
```

### Main.go Pattern

```go
// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
)

func main() {
	// TODO: Load config from Story 2.1
	port := os.Getenv("APP_HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil, // TODO: Add router in Story 3.x
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	done := make(chan error, 1)
	go app.GracefulShutdown(server, done)

	if err := <-done; err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
	os.Exit(0)
}
```

### Testing Strategy

**Unit Tests (internal/app/shutdown_test.go):**
1. `TestGracefulShutdown_SignalHandling` - Send signal, verify channel receives
2. `TestGracefulShutdown_Timeout` - Mock slow handler, verify timeout enforced
3. `TestGracefulShutdown_CleanExit` - Verify no error on clean shutdown

Per project_context.md:
- Table-driven tests with `t.Run`
- Use testify (require/assert)
- `t.Parallel()` when safe
- Naming: `Test<Thing>_<Behavior>`

### Project Structure Notes

Files to create:
- `internal/app/shutdown.go` - Graceful shutdown logic
- `internal/app/shutdown_test.go` - Unit tests

Files to modify:
- `cmd/server/main.go` - Add server and shutdown integration

**Layer:** `internal/app` is the application layer (wiring).
**Import Rules:** Can import all internal packages. cmd/ imports app.

### Dependencies

No new dependencies required. Uses Go stdlib:
- `os/signal`
- `syscall`
- `net/http` (server.Shutdown)
- `context` (timeout)

### NFR Reference

From epics.md NFR5:
- **Graceful shutdown < 30 seconds** (matches ShutdownTimeout constant)

From NFR7:
- **Shutdown completes all in-flight requests** (server.Shutdown behavior)

### References

- [Source: docs/epics.md#Story-1.4]
- [Source: docs/epics.md#NFR5-Graceful-shutdown]
- [Source: docs/project_context.md#Testing-Rules]
- [Source: Story 1.3 - APP_HTTP_PORT in .env.example]

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.

### Agent Model Used

dev-story workflow execution.

### Debug Log References

None.

### Completion Notes List

- Story created: 2025-12-11
- Validation applied: 2025-12-11
- Implementation completed: 2025-12-11
- Code review fixes applied: 2025-12-11
  - CRITICAL: Added `defer signal.Stop(quit)` to prevent handler leak
  - Coverage improved: 77.8% → 80%
  - Lint: 0 issues

### File List

Files created:
- `internal/app/shutdown.go` - Graceful shutdown handler (35 lines)
- `internal/app/shutdown_test.go` - Unit tests (54 lines)

Files modified:
- `cmd/server/main.go` - HTTP server with shutdown integration
- `go.mod` - Updated dependencies
- `go.sum` - Updated checksums
