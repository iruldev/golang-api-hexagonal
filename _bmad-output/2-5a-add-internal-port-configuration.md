# Story 2.5a: Add INTERNAL_PORT Configuration

Status: done

## Story

As a **platform engineer**,
I want to configure an internal port for administrative endpoints,
so that internal services like /metrics can be isolated from public traffic.

## Acceptance Criteria

1. **Given** `INTERNAL_PORT` env var is set
   **When** application starts
   **Then** config has InternalPort field with that value

2. **Given** `INTERNAL_PORT` equals `PORT`
   **When** application starts
   **Then** startup fails with clear error message

3. **And** default value is 8081

4. **Given** `INTERNAL_BIND_ADDRESS` is set (or default)
   **When** application starts
   **Then** config has InternalBindAddress field with that value (default 127.0.0.1)

5. **Given** Port or InternalPort is 0
   **When** application starts
   **Then** validation passes (dynamic allocation support)

## Tasks / Subtasks

- [x] Task 1: Add InternalPort to Config struct
  - [x] Added `InternalPort int` field with default 8081
  - [x] Added envconfig tag `INTERNAL_PORT`

- [x] Task 2: Add Validation
  - [x] Validated InternalPort is in valid port range (0-65535)
  - [x] Validated InternalPort != Port (prevent collision)
  - [x] Error message: "INTERNAL_PORT must differ from PORT"

- [x] Task 3: Add Unit Tests
  - [x] `TestLoad_InternalPortDefault` - default is 8081
  - [x] `TestLoad_InternalPortCustom` - custom value works
  - [x] `TestLoad_InternalPortCollision` - collision fails
  - [x] `TestLoad_InternalPortInvalidRange` - invalid range fails

- [x] Task 4: Fix Code Review Findings (AI-Auto-Fix)
  - [x] Add `InternalBindAddress` field + validation
  - [x] Allow port 0 (dynamic allocation) for both Port and InternalPort
  - [x] Add `INTERNAL_PORT` and `INTERNAL_BIND_ADDRESS` to `.env.example`
  - [x] Update tests for new fields and validation rules

## Dev Notes

### Implementation

```go
// Internal Server (Story 2.5a)
InternalPort int `envconfig:"INTERNAL_PORT" default:"8081"`
// InternalBindAddress is the bind address for the internal server.
InternalBindAddress string `envconfig:"INTERNAL_BIND_ADDRESS" default:"127.0.0.1"`
```

### Validation

```go
// Allow 0 for dynamic port allocation
if c.Port < 0 || c.Port > 65535 {
    return fmt.Errorf("invalid PORT: must be between 0 and 65535")
}
if c.InternalPort < 0 || c.InternalPort > 65535 {
    return fmt.Errorf("invalid INTERNAL_PORT: must be between 0 and 65535")
}
if c.InternalBindAddress == "" {
    return fmt.Errorf("INTERNAL_BIND_ADDRESS cannot be empty")
}
```

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro (Dev), Gemini 2.5 Pro (Reviewer)

### Debug Log References

- Build: SUCCESS
- New Tests: 6 PASS (Total in package)
- Regression: 16 packages ALL PASS

### Completion Notes List

- Added `InternalPort` and `InternalBindAddress` fields
- Added validation for port range (allowing 0) and collision
- Added unit tests for all scenarios including defaults and bounds
- Updated documentation in `.env.example`

### File List

- `internal/infra/config/config.go` - MODIFIED (added InternalPort/BindAddress + validation)
- `internal/infra/config/config_test.go` - MODIFIED (added tests)
- `.env.example` - MODIFIED (added internal config)

### Change Log

- 2024-12-24: Added InternalPort field with default 8081
- 2024-12-24: Added validation for port range and collision with PORT
- 2024-12-24: Added 4 unit tests
- 2024-12-24: (Review) Added InternalBindAddress, relaxed port validation, updated docs

## Senior Developer Review (AI)

- **Date**: 2024-12-24
- **Reviewer**: AI (Adversarial Reviewer)
- **Outcome**: Approved (with fixes)

### Findings & Fixes

1.  **Security/Isolation**: Identified missing `INTERNAL_BIND_ADDRESS`. Added defaulting to `127.0.0.1` to ensure internal metrics aren't exposed publicly by default.
2.  **Flexibility**: Relaxed port validation to allow port `0` for dynamic allocation (useful in integration tests).
3.  **Documentation**: Added missing `INTERNAL_PORT` configuration to `.env.example`.
4.  **Verification**: All tests passed, including new cases for the fixes.
