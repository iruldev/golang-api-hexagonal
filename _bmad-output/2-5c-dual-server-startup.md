# Story 2.5c: Dual Server Startup with Graceful Shutdown

Status: done

## Story

As a **platform engineer**,
I want both public and internal servers to run simultaneously,
so that /metrics is available on internal port while API is on public port.

**Dependencies:** Story 2.5a (Config), Story 2.5b (Internal Router)

## Acceptance Criteria

1. **Given** application starts
   **When** both PORT and INTERNAL_PORT are configured
   **Then** two HTTP servers run concurrently

2. **Given** shutdown signal (SIGINT/SIGTERM)
   **When** received
   **Then** both servers gracefully shutdown with same timeout

3. **And** startup logs show both ports

## Tasks / Subtasks

- [x] Task 1: Create Internal Server in main.go
  - [x] Create second `http.Server` for internal router
  - [x] Bind to INTERNAL_PORT

- [x] Task 2: Start Both Servers Concurrently
  - [x] Use goroutines for both servers
  - [x] Use errgroup or WaitGroup for coordination

- [x] Task 3: Implement Dual Graceful Shutdown
  - [x] Shutdown both servers on signal
  - [x] Use same timeout (e.g., 15s) for both
  - [x] Log shutdown progress for each

- [x] Task 4: Add Startup Logging
  - [x] Log "public server starting on :PORT"
  - [x] Log "internal server starting on :INTERNAL_PORT"

## Dev Notes

### Implementation Details
The `run()` function in `cmd/api/main.go` was updated to:
1. Create `publicServer` and `internalServer`.
2. Start both in separate goroutines using error channels.
3. Wait for shutdown signal.
4. Use `sync.WaitGroup` to shutdown both servers concurrently with a shared context timeout.

```go
// Start servers in goroutines
go func() {
    logger.Info("public server listening", slog.String("addr", publicAddr))
    serverErrors <- publicSrv.ListenAndServe()
}()

go func() {
    logger.Info("internal server listening", slog.String("addr", internalAddr))
    serverErrors <- internalSrv.ListenAndServe()
}()
```

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro (via Workflow)

### Debug Log References
- Implementation verified in `main.go`.
- Regression tests (16 packages) PASS.

### Completion Notes List
- Confirmed concurrent startup of Public (:PORT) and Internal (:INTERNAL_PORT) servers.
- Verified graceful shutdown logic handles both servers.
- Verified startup logging includes both addresses.

### File List
- `cmd/api/main.go` - MODIFIED (Implemented dual server startup and shutdown, refactored timeouts, added no-op DB pool for smoke tests)
- `internal/infra/config/config.go` - MODIFIED (Added timeout configurations and hidden smoke test flags)
- `cmd/api/smoke_test.go` - NEW (Added binary startup smoke test)

### Change Log
- 2024-12-24: Implemented dual server startup and graceful shutdown in `main.go` (completed during 2.5b follow-up)
- 2024-12-24: Documented completion and verified regression tests.
- 2024-12-24: [Code Review] Refactored hardcoded timeouts to configuration and added smoke test for verification.
- 2024-12-24: [Re-Review] Fixed smoke test implementation and added `IGNORE_DB_STARTUP_ERROR` logic to `main.go` to support DB-less verify.
