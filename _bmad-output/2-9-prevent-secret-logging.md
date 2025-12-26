# Story 2.9: Prevent Secret Logging

Status: done

## Story

**As a** security engineer,
**I want** JWT_SECRET and Authorization headers never logged,
**So that** secrets don't leak to log aggregators.

**FR:** FR41

## Acceptance Criteria

1. **Given** any log statement in the codebase
   **When** it could log request headers or config
   **Then** JWT_SECRET and Authorization are redacted

2. **Given** the codebase
   **When** a code review is performed
   **Then** the checklist includes secret logging check

3. **Given** the codebase
   **When** grep search is performed
   **Then** no secret logging patterns are found

## Tasks / Subtasks

- [x] Task 1: Verify JWT_SECRET is never logged
  - [x] Grep found only `fmt.Errorf` (error messages, not logs) ✅
  - [x] Config loading does not log secrets ✅
  - [x] Existing redactor covers auth tokens ✅

- [x] Task 2: Verify Authorization header is redacted
  - [x] `fieldAuthorization` in `redactor.go` exists ✅
  - [x] **FIXED:** Request logger was logging full JWT in `auth.go:106`
  - [x] Removed `authHeader` from log statement

- [x] Task 3: Verify config struct is NOT logged
  - [x] `main.go` only logs: port, log_level, otel_enabled (non-sensitive) ✅
  - [x] No config struct dump with secrets ✅

- [x] Task 4: Add code review checklist entry
  - [x] Created `CONTRIBUTING.md` with security section
  - [x] Added "No JWT_SECRET logging" check
  - [x] Added "No Authorization header logging" check

- [x] Task 5: Run verification grep commands
  - [x] `grep -r "JWT_SECRET" --include="*.go" | grep -i "log\|print"` - Only error messages ✅
  - [x] `grep -r 'Header' middleware/*.go | grep -i log` - No issues after fix ✅

## Dev Notes

### Critical Fix (auth.go)

**Before (LINE 106 - UNSAFE):**
```go
logger.WarnContext(r.Context(), "auth failed: invalid header format", "header", authHeader)
```

**After (SAFE):**
```go
// Story 2.9: Never log Authorization header - it contains sensitive tokens
logger.WarnContext(r.Context(), "auth failed: invalid header format")
```

### Verification Results

| Check | Result |
|-------|--------|
| JWT_SECRET in logs | ✅ None (only fmt.Errorf) |
| Authorization in logs | ✅ **FIXED** |
| Config struct logging | ✅ Safe (only non-sensitive fields) |
| PIIRedactor patterns | ✅ `fieldAuthorization` exists |

### New File: CONTRIBUTING.md

Created security checklist for code reviews:
- No JWT_SECRET logging
- No Authorization header logging
- No config dump
- Use PIIRedactor for user data

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro

### Debug Log References
- Middleware tests: PASS
- Full regression: 16 packages PASS

### Completion Notes List
- Found and fixed JWT token leak in `auth.go:106`
- Verified config logging is safe (only non-sensitive fields)
- Created `CONTRIBUTING.md` with security checklist
- Existing PIIRedactor covers `authorization` field

### File List
- `internal/transport/http/middleware/auth.go` - MODIFIED (removed authHeader from log)
- `CONTRIBUTING.md` - NEW (security checklist)

### Change Log
- 2024-12-24: Fixed JWT token leak in auth.go:106
- 2024-12-24: Created CONTRIBUTING.md with security checklist
- 2024-12-24: Verified all grep checks pass
- 2024-12-24: Full regression passes (16 packages)
