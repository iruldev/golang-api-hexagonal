# Story 1.6: Implement Fail-Fast Config Validation

Status: Done

## Story

As a developer,
I want the server to fail immediately if required config is missing,
So that I don't discover config issues at runtime.

## Acceptance Criteria

1. **Given** a required environment variable is missing
   **When** the server starts
   **Then** it exits immediately with error code 1
   **And** the error message identifies missing config without leaking secrets

2. **Given** a required environment variable has an invalid format
   **When** the server starts
   **Then** it exits immediately with error code 1
   **And** the error message explains what format is expected

3. **Given** all required environment variables are correctly set
   **When** the server starts
   **Then** it initializes successfully
   **And** no validation errors are logged

4. **Given** multiple required config values are missing
   **When** the server starts
   **Then** ALL missing values are reported in a single error (not just the first)
   **And** the developer can fix all issues in one attempt

5. **Given** a config error occurs
   **When** the error message is generated
   **Then** it does NOT contain sensitive values (passwords, secrets, tokens)
   **And** it identifies the missing/invalid config by name only

## Tasks / Subtasks

- [x] Task 1: Audit and enhance existing config validation (AC: #1, #2, #4)
  - [x] 1.1 Audit `internal/config/validate.go` for completeness against FR40
  - [x] 1.2 Verify all critical fields are marked as required (DB_HOST, DB_USER, DB_NAME, APP_HTTP_PORT)
  - [x] 1.3 Add validation for any missing critical fields discovered in audit
  - [x] 1.4 Ensure `ValidationError` collects ALL errors, not just first one

- [x] Task 2: Verify fail-fast behavior in entry points (AC: #1)
  - [x] 2.1 Verify `cmd/server/main.go` exits with code 1 on config error
  - [x] 2.2 Verify `cmd/worker/main.go` has same fail-fast pattern
  - [x] 2.3 Verify any other entry points (`cmd/scheduler/`, `cmd/bplat/`) have fail-fast

- [x] Task 3: Ensure secure error messaging (AC: #5)
  - [x] 3.1 Audit `ValidationError.Error()` for secret leaks
  - [x] 3.2 Verify password/secret fields are NEVER included in error messages
  - [x] 3.3 Add test cases for secret-safe error messages

- [x] Task 4: Add comprehensive test coverage (AC: All)
  - [x] 4.1 Test missing required fields produces correct errors
  - [x] 4.2 Test invalid format produces correct errors
  - [x] 4.3 Test multiple errors are collected together
  - [x] 4.4 Test valid config passes validation

- [x] Task 5: Documentation and verification (AC: All)
  - [x] 5.1 Document required environment variables in README.md or docs
  - [x] 5.2 Add example `.env.example` file if not exists
  - [x] 5.3 Verify `make up` works with sample config
  - [x] 5.4 Verify missing config causes immediate exit with clear error

## Dev Notes

### Existing Implementation Analysis

**✅ Already Implemented:**
The project already has substantial config validation infrastructure in place:

1. **Config Loader** (`internal/config/loader.go`):
   - `Load()` function validates config after loading
   - Returns error with "config validation failed: ..." prefix
   - Already calls `cfg.Validate()` and fails if error

2. **Validation Logic** (`internal/config/validate.go`):
   - `ValidationError` type collects multiple errors ✓
   - `Validate()` method checks all sections (DB, App, Redis, Asynq) ✓
   - Individual validators: `validateDatabase()`, `validateApp()`, `validateRedis()`, `validateAsynq()`

3. **Entry Point** (`cmd/server/main.go`):
   - Uses `log.Fatalf("Configuration error: %v", err)` for fail-fast ✓
   - This exits with code 1 as required

**⚠️ Gaps to Address:**
- Verify worker entry points have same pattern
- Ensure all critical fields documented in `.env.example` or README
- Add explicit tests for secret-safe error messages

### Required Environment Variables (Current)

From `internal/config/validate.go`:
| Variable | Required | Validation |
|----------|----------|------------|
| `DB_HOST` | ✓ | Non-empty |
| `DB_PORT` | ✓ | 1-65535 |
| `DB_USER` | ✓ | Non-empty |
| `DB_NAME` | ✓ | Non-empty |
| `DB_PASSWORD` | ✗ | Optional (trust-based local dev) |
| `APP_HTTP_PORT` | ✓ | 1-65535 |
| `APP_ENV` | ✗ | Optional, if set must be: development, staging, production |
| `REDIS_*` | ✗ | Optional (only if REDIS_HOST set) |
| `ASYNQ_*` | ✗ | Optional |

### Security Considerations

**Secret-Safe Error Messages:**
- Error messages must identify the config KEY, not the VALUE
- Example: "DB_HOST is required" ✓
- Never: "DB_PASSWORD 'actual_password' is invalid" ✗

The current implementation is already secret-safe because:
1. `ValidationError.Error()` only contains field names
2. No secret values are ever appended to error strings

### File Structure

```
project-root/
├── cmd/
│   ├── server/main.go      # [VERIFY] fail-fast pattern
│   └── worker/main.go      # [VERIFY] fail-fast pattern
├── internal/
│   └── config/
│       ├── config.go       # [REFERENCE] Config struct
│       ├── loader.go       # [REFERENCE] Load() with validation
│       ├── loader_test.go  # [ENHANCE] Add tests
│       └── validate.go     # [REFERENCE] Validation logic
│       └── validate_test.go # [ENHANCE] Add secret-safe tests
├── .env.example            # [CREATE/UPDATE] Document required vars
└── README.md               # [UPDATE] Document config requirements
```

### Testing Strategy

1. **Unit Tests** (`internal/config/*_test.go`):
   - Test each validation function individually
   - Test error collection (multiple errors)
   - Test boundary values (port 0, 65535, 65536)

2. **Integration Tests**:
   - Start server with missing config → verify exit code 1
   - Verify error message format

### NFR Targets

| NFR | Requirement | Verification |
|-----|-------------|--------------|
| NFR-S3 | Secrets via env/secret manager only | Error messages don't leak secrets |
| FR40 | Fail-fast if config required missing | Exit code 1 with clear message |

### Previous Story Learnings (from Story 1.5)

- CI validates lint + test, so any new code must pass `make verify`
- Use existing patterns from Story 1.4/1.5 for consistency
- Exit code 1 blocks operations (same pattern for config fail-fast)

### Critical Points

1. **ValidationError Pattern:** Use existing `ValidationError` type to collect all errors
2. **Fail-Fast:** `log.Fatalf` already used in main.go - verify all entry points
3. **No Secret Leaks:** Never include actual secret values in error messages
4. **Test Coverage:** Existing tests in `loader_test.go` and `validate_test.go` - expand as needed
5. **Documentation:** Make required config discoverable for new developers

### References

- [Source: docs/epics.md#Story 1.6](file:///docs/epics.md) - FR40: System fail-fast jika config required missing
- [Source: docs/prd.md](file:///docs/prd.md) - NFR-S3: Secrets via env/secret manager only
- [Source: docs/architecture.md](file:///docs/architecture.md) - Layer structure
- [Source: internal/config/validate.go](file:///internal/config/validate.go) - Existing validation logic
- [Source: internal/config/loader.go](file:///internal/config/loader.go) - Load() with fail-fast validation
- [Source: cmd/server/main.go](file:///cmd/server/main.go) - Entry point with log.Fatalf fail-fast

## Dev Agent Record

### Context Reference

- Story generated by create-story workflow on 2025-12-15
- Epic 1: Foundation & Quality Gates (MVP) - in-progress
- Previous stories: 1.1 (done), 1.2 (done), 1.3 (done), 1.4 (done), 1.5 (done)

### Agent Model Used

Gemini Claude 3.5 Sonnet

### Debug Log References

None required.

### Completion Notes List

- **Task 1**: Audited `internal/config/validate.go` - all critical fields already validated. `ValidationError` correctly collects ALL errors.
- **Task 2**: Verified fail-fast in all entry points:
  - `cmd/server/main.go` uses `log.Fatalf("Configuration error: %v", err)` (exit code 1)
  - `cmd/worker/main.go` uses same pattern
  - `cmd/scheduler/main.go` uses same pattern
  - `cmd/bplat/` is CLI tool (cobra-based), config validation handled per-command
- **Task 3**: `ValidationError.Error()` is secret-safe - only includes field names, never values
- **Task 4**: Added 2 new test functions to `validate_test.go`:
  - `TestValidationError_SecretSafe` - verifies DB_PASSWORD and REDIS_PASSWORD never leak
  - `TestValidationError_MultipleErrorCollection` - verifies FR40 all-errors-at-once behavior
- **Task 5**: `.env.example` already exists with comprehensive documentation including required fields notation at line 28

### File List

| File | Action |
|------|--------|
| `internal/config/validate_test.go` | Modified - added secret-safe and FR40 tests |
| `docs/sprint-artifacts/sprint-status.yaml` | Modified - story status |
| `docs/sprint-artifacts/1-6-implement-fail-fast-config-validation.md` | Modified - task checkboxes and completion notes |
| `internal/config/validate.go` | Modified - improved error message formatting (newlines) |

### Review Follow-ups (Auto-Fixed)

- [x] **Missing Test Coverage**: Added `TestValidate_Asynq` and `TestValidate_Redis` to cover all validation logic branches.
- [x] **Error Formatting**: Changed `ValidationError.Error()` separator from semicolon to newline for better readability of multiple errors (FR40).
- [x] **Hardcoded Strings**: Accepted as low priority technical debt for now (strings are consistent across tests/code).

### Change Log

| Date | Change |
|------|--------|
| 2025-12-15 | Story created with comprehensive context from architecture, PRD, and existing implementation analysis |
| 2025-12-15 | Story implemented - verified existing fail-fast behavior, added secret-safe and FR40 tests, all ACs satisfied |
| 2025-12-15 | Code Review - Auto-fixed missing tests for Asynq/Redis and improved error message formatting |
