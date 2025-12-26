# Story 1.3: Add ActorID and RequestID to Audit Events

Status: done

## Story

As a **security auditor**,
I want audit events to include ActorID and RequestID metadata,
so that I can trace who performed actions and correlate with request logs.

## Acceptance Criteria

1. **Given** a user performs an auditable action
   **When** the audit event is recorded
   **Then** the event includes `actor_id` (user ID or empty for system)

2. **And** the event includes `request_id` from request context

3. **And** unit tests verify both fields are populated correctly

## Tasks / Subtasks

- [x] Task 1: Update User Handler (AC: #1, #2)
  - [x] Edit `internal/transport/http/handler/user.go` lines 82-88
  - [x] Add imports: `middleware` package
  - [x] Extract RequestID: `middleware.GetRequestID(r.Context())`
  - [x] Extract ActorID: `app.GetAuthContext(r.Context())` → `domain.ID(authCtx.SubjectID)`
  - [x] Populate `user.CreateUserRequest.RequestID` and `.ActorID`

- [x] Task 2: Verify End-to-End Flow (AC: #1, #2)
  - [x] Trace: Handler → UseCase → AuditService → Repo
  - [x] Confirm `create_user.go:112-118` passes fields to `AuditEventInput`
  - [x] Confirm `audit_event_repo.go:45-48` persists both columns

- [x] Task 3: Add Handler Test (AC: #3)
  - [x] Edit `internal/transport/http/handler/user_test.go`
  - [x] Added `middleware.SetRequestID()` helper for test context setup
  - [x] Test: PropagatesRequestIDAndActorID - PASS
  - [x] Test: EmptyActorIDWhenNoAuthContext - PASS

## Dev Notes

### Architecture

| Layer | File | Role |
|-------|------|------|
| Transport | `handler/user.go` | Extract from context |
| App | `user/create_user.go` | Pass to AuditService |
| App | `audit/service.go` | Build AuditEvent |
| Infra | `postgres/audit_event_repo.go` | Persist to DB |

### Key Helpers

- **RequestID:** `middleware.GetRequestID(ctx)` → `string`
- **SetRequestID:** `middleware.SetRequestID(ctx, id)` → `context.Context` (for tests)
- **AuthContext:** `app.GetAuthContext(ctx)` → `*app.AuthContext` (nil if unauthenticated)
- **System Actor:** Empty `domain.ID` = system/unauthenticated

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- Build: SUCCESS
- Tests: `go test ./internal/transport/http/handler/... -count=1` - ALL PASS (20 tests)
- Regression: `go test ./... -count=1` - ALL PASS (15 packages)

### Completion Notes List

- Updated `handler/user.go` to extract RequestID and ActorID from context
- Added `middleware.SetRequestID()` helper for test context setup
- Added 2 new tests: PropagatesRequestIDAndActorID, EmptyActorIDWhenNoAuthContext
- Verified end-to-end flow: Handler → UseCase → AuditService → Repo
- All 15 packages pass tests with no regressions

### File List

- `internal/transport/http/handler/user.go` - MODIFIED (extract RequestID/ActorID)
- `internal/transport/http/handler/user_test.go` - MODIFIED (add 2 propagation tests)
- `internal/transport/http/middleware/requestid.go` - MODIFIED (SetRequestID helper, UUIDv7 fix)
- `internal/transport/http/middleware/requestid_test.go` - MODIFIED (UUIDv7 tests)
- `internal/transport/http/middleware/logging_test.go` - MODIFIED (UUID format fix)
- `internal/infra/postgres/audit_event_repo.go` - MODIFIED (ActorID robustness)

### Change Log

- 2024-12-24: Updated handler to extract and propagate RequestID and ActorID
- 2024-12-24: Added SetRequestID test helper to middleware
- 2024-12-24: Added unit tests verifying ID propagation
- 2024-12-24: [Code Review] Fixed RequestID middleware to use strictly consistent UUIDv7 format
- 2024-12-24: [Code Review] Enhanced AuditEventRepo to robustly handle invalid ActorID formats
- 2024-12-24: [Code Review Round 2] Added DoS protection for RequestID (max length 50)
- 2024-12-24: [Code Review Round 2] Added error logging for dropped invalid ActorIDs
- 2024-12-24: [Code Review Round 3] Enhanced DoS protection with strict charset validation
- 2024-12-24: [Code Review Round 3] Migrated AuditEventRepo unstructured logging to standard log/slog



