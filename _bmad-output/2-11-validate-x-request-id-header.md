# Story 2.11: Validate X-Request-ID Header

Status: done

## Story

**As a** security engineer,
**I want** X-Request-ID to be validated and normalized,
**So that** log injection attacks are prevented.

**FR:** FR43

## Acceptance Criteria

1. **Given** X-Request-ID with invalid characters or >64 chars
   **When** the request is processed
   **Then** a new valid request_id is generated

2. **Given** the request ID
   **When** logging occurs
   **Then** only validated request_id is used

3. **Given** the implementation
   **When** unit tests are run
   **Then** validation rules are covered

## Tasks / Subtasks

- [x] Task 1: Validate X-Request-ID
  - [x] `isValidRequestID()` checks length ≤64 chars ✅
  - [x] `isSafeChar()` allows only alphanumeric, `-`, `_`, `:`, `.` ✅
  - [x] Invalid IDs trigger `generateRequestID()` ✅

- [x] Task 2: Update max length from 50 to 64
  - [x] Updated `requestid.go` line 70: `len(id) > 64`
  - [x] Added story reference comment

- [x] Task 3: Verify unit tests
  - [x] `TestRequestID_IgnoresTooLongID` - PASS (updated to 65 chars)
  - [x] `TestRequestID_IgnoresInvalidCharset` - PASS
  - [x] All 11 RequestID tests pass ✅

## Dev Notes

### Implementation

**File:** `internal/transport/http/middleware/requestid.go`

```go
// isValidRequestID checks if the ID is safe and valid (length <= 64).
// Story 2.11: Updated max length from 50 to 64 per AC.
func isValidRequestID(id string) bool {
    if id == "" || len(id) > 64 {
        return false
    }
    for _, c := range id {
        if !isSafeChar(c) {
            return false
        }
    }
    return true
}
```

### Test Results

```
--- PASS: TestRequestID_GeneratesNewID
--- PASS: TestRequestID_PassthroughExistingID
--- PASS: TestRequestID_IgnoresTooLongID (65 chars → regenerate)
--- PASS: TestRequestID_IgnoresInvalidCharset
--- PASS: TestRequestID_ResponseHeader
+ 6 more tests
PASS (11/11)
```

### Security Pattern

| Input | Action |
|-------|--------|
| Empty ID | Generate new UUID v7 |
| >64 chars | Generate new UUID v7 |
| Invalid chars ($, %, etc) | Generate new UUID v7 |
| Valid ID ≤64 chars | Passthrough |

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro

### Debug Log References
- RequestID tests: 11/11 PASS

### Completion Notes List
- Updated max length from 50 to 64 per AC
- Updated test `IgnoresTooLongID` to use 65-char ID
- All validation tests pass

### File List
- `internal/transport/http/middleware/requestid.go` - MODIFIED (64 char limit)
- `internal/transport/http/middleware/requestid_test.go` - MODIFIED (65 char test)

### Change Log
- 2024-12-24: Updated max length from 50 to 64 per AC
- 2024-12-24: Updated test to use 65-char string
- 2024-12-24: All 11 RequestID tests pass

## Senior Developer Review (AI)
- **Status:** Approved
- **Findings:**
    - Critical discrepancy found: Code was checking for > 50 chars despite AC requiring > 64. fixed.
    - Test was using 51 chars, updated to 65 chars to properly test the new limit.
- **Outcome:** Fixes applied and verified. Story is now compliant.

## Senior Developer Review - Re-verification (AI)
- **Status:** Approved
- **Findings:**
    - Implementation verified: `len(id) > 64` check is correct.
    - Test verified: Uses 65-char string.
    - Tests passing: All 11 tests passed.
- **Outcome:** Fixes are solid. No regressions.
