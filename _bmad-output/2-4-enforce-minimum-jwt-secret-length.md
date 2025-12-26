# Story 2.4: Enforce Minimum JWT Secret Length

Status: done

## Story

As a **security engineer**,
I want JWT secrets shorter than 32 bytes to be rejected,
so that weak secrets cannot be used in production.

## Acceptance Criteria

1. **Given** `JWT_SECRET` is 20 bytes
   **When** the application starts
   **Then** startup fails with "JWT_SECRET must be >= 32 bytes"

2. **And** unit test verifies length validation

## Tasks / Subtasks

- [x] Task 1: Verify Existing Validation (AC: #1)
  - [x] Confirmed `config.go:131-132` has length check
  - [x] Confirmed error message: "JWT_SECRET must be at least 32 bytes when JWT_ENABLED is true"

- [x] Task 2: Add/Verify Unit Test (AC: #2)
  - [x] Added `TestLoad_JWTSecretTooShort` - 20 bytes → error
  - [x] Added `TestLoad_JWTSecretExactly32Bytes` - 32 bytes → passes
  - [x] Added `TestLoad_JWTSecretOver32Bytes` - 52 bytes → passes

## Dev Notes

### Verified Implementation ✅

```go
// config.go - Normalized and Validated
c.JWTSecret = strings.TrimSpace(c.JWTSecret)
// ...
if c.JWTEnabled {
    if c.JWTSecret == "" {
        return fmt.Errorf("JWT_ENABLED is true but JWT_SECRET is empty")
    }
    if len(c.JWTSecret) < 32 {
        return fmt.Errorf("JWT_SECRET must be at least 32 bytes when JWT_ENABLED is true")
    }
}
```

### New Test Cases

| Test | Secret Length | Expected |
|------|---------------|----------|
| `TestLoad_JWTSecretTooShort` | 20 bytes | Error |
| `TestLoad_JWTSecretExactly32Bytes` | 32 bytes | Pass |
| `TestLoad_JWTSecretOver32Bytes` | 52 bytes | Pass |
| `TestLoad_JWTSecretNormalization` | 32 bytes (padded) | Pass (Trimmed) |
| `TestConfig_Redacted` | N/A | Pass (Redacted) |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro (Initial), Antigravity (Code Review & Fix)

### Debug Log References

- Verification: `config.go` validation logic confirmed
- New Tests: 5 PASS (3 initial + 2 review fixes)
- Regression: 16 packages ALL PASS

### Completion Notes List

- Verified existing JWT secret length validation in `config.go`
- Added 3 boundary-testing unit tests covering <32, =32, >32 scenarios
- **Review Fixes**:
    - Fixed normalization bug where `JWTSecret` wasn't trimmed in struct
    - Added `Redacted()` method to prevent secret leakage in logs
    - Optimized string trimming in validation logic
    - Added 2 new tests for normalization and redaction

### File List

- `internal/infra/config/config.go` - MODIFIED (normalization & redaction)
- `internal/infra/config/config_test.go` - MODIFIED (added 5 tests total)

### Change Log

- 2024-12-24: Verified existing length validation at config.go:131-132
- 2024-12-24: Added 3 boundary-testing unit tests for JWT secret length
- 2024-12-24: [Review] Fixed normalization bug and added Redacted() method
- 2024-12-24: [Review] Added tests for normalization and redaction
