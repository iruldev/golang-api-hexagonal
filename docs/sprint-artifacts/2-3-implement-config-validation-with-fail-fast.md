# Story 2.3: Implement Config Validation with Fail-Fast

Status: done

## Story

As a SRE,
I want the system to fail fast on invalid configuration,
So that misconfigurations are caught at startup, not runtime.

## Acceptance Criteria

### AC1: Required field validation ✅
**Given** required config `DB_HOST` is missing
**When** the application starts
**Then** startup fails with exit code 1
**And** error message indicates `DB_HOST is required`

### AC2: Type mismatch validation ✅
**Given** `APP_HTTP_PORT` is set to "invalid"
**When** the application starts
**Then** startup fails with validation error
**And** error message indicates invalid value

> [!NOTE]
> **Implementation Reality:** koanf silently converts "invalid" → 0 for int fields.
> Our validation catches `port <= 0` and reports: "must be between 1 and 65535".
> This satisfies "fail-fast" intent even though message differs from original AC wording.

---

## Tasks / Subtasks

- [x] **Task 1: Create validation module** (AC: #1, #2)
  - [x] Create `internal/config/validate.go`
  - [x] Implement `Validate(cfg *Config) error` function
  - [x] Define required fields for each config section
  - [x] Return error with field name and reason

- [x] **Task 2: Add required field checks** (AC: #1)
  - [x] Validate `DB.Host` is not empty
  - [x] Validate `DB.Port` > 0
  - [x] Validate `DB.User` is not empty
  - [x] Validate `DB.Name` is not empty
  - [x] Validate `App.HTTPPort` > 0
  - [x] Collect ALL errors, don't stop at first

- [x] **Task 3: Add range/format validations** (AC: #2)
  - [x] Validate `App.Env` is one of: development, staging, production
  - [x] Validate port ranges (1-65535)
  - [x] Validate `DB.MaxOpenConns` >= 0
  - [x] Validate `DB.MaxIdleConns` >= 0

- [x] **Task 4: Integrate validation into loader** (AC: #1, #2)
  - [x] Call `Validate()` after unmarshal in `Load()`
  - [x] Return wrapped error if validation fails

- [x] **Task 5: Create comprehensive tests** (AC: #1, #2)
  - [x] Create `internal/config/validate_test.go`
  - [x] Test missing required field (DB_HOST)
  - [x] Test invalid port (negative)
  - [x] Test invalid env value (APP_ENV="invalid")
  - [x] Test multiple errors collected
  - [x] Test valid config passes
  - [x] Test Load() with invalid env type (APP_HTTP_PORT="invalid" → 0 → error)

- [x] **Task 6: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass, coverage ≥90% ✅ (98.2%)
  - [x] Run `make lint` - 0 issues ✅

---

## Dev Notes

### Architecture Compliance

**Layer:** `internal/config` (allowed: stdlib, external config libs)
**Pattern:** Validation as pure function, no side effects
**Error Style:** Collect all errors, wrap with context

### Existing Code Analysis

**Current `loader.go` (80 lines):**
```go
// After unmarshal, add validation call:
var cfg Config
if err := k.Unmarshal("", &cfg); err != nil {
    return nil, err
}

// NEW: Validate before returning
if err := cfg.Validate(); err != nil {
    return nil, fmt.Errorf("config validation failed: %w", err)
}

return &cfg, nil
```

**Current `config.go` (44 lines):**
- `Config` struct with nested: App, Database, Observability, Log
- No validation tags currently - we'll add method-based validation

### Required Fields Analysis

| Field | Section | Required | Notes |
|-------|---------|----------|-------|
| Host | Database | ✅ | Cannot connect without |
| Port | Database | ✅ | Default 5432 típico |
| User | Database | ✅ | Auth required |
| Name | Database | ✅ | DB name required |
| HTTPPort | App | ✅ | Server needs port |
| Password | Database | ❓ | Could be empty for local |

**Decision:** Password NOT required (trust-based local dev).

### Validation Implementation

```go
// internal/config/validate.go
package config

import (
    "fmt"
    "strings"
)

// ValidationError holds multiple validation errors.
type ValidationError struct {
    Errors []string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("config validation failed: %s", strings.Join(e.Errors, "; "))
}

// Is supports errors.Is() pattern for type checking.
func (e *ValidationError) Is(target error) bool {
    _, ok := target.(*ValidationError)
    return ok
}

// Validate checks configuration for required fields and valid ranges.
func (c *Config) Validate() error {
    var errs []string

    // Database required fields
    if c.Database.Host == "" {
        errs = append(errs, "DB_HOST is required")
    }
    if c.Database.Port <= 0 || c.Database.Port > 65535 {
        errs = append(errs, "DB_PORT must be between 1 and 65535")
    }
    if c.Database.User == "" {
        errs = append(errs, "DB_USER is required")
    }
    if c.Database.Name == "" {
        errs = append(errs, "DB_NAME is required")
    }

    // App required fields
    if c.App.HTTPPort <= 0 || c.App.HTTPPort > 65535 {
        errs = append(errs, "APP_HTTP_PORT must be between 1 and 65535")
    }

    // App.Env validation (optional but if set, must be valid)
    if c.App.Env != "" {
        validEnvs := map[string]bool{
            "development": true,
            "staging":     true,
            "production":  true,
        }
        if !validEnvs[c.App.Env] {
            errs = append(errs, "APP_ENV must be one of: development, staging, production")
        }
    }

    // Connection pool validation
    if c.Database.MaxOpenConns < 0 {
        errs = append(errs, "DB_MAX_OPEN_CONNS must be >= 0")
    }
    if c.Database.MaxIdleConns < 0 {
        errs = append(errs, "DB_MAX_IDLE_CONNS must be >= 0")
    }

    if len(errs) > 0 {
        return &ValidationError{Errors: errs}
    }
    return nil
}
```

### Testing Pattern

```go
// internal/config/validate_test.go
package config

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestValidate_MissingDBHost(t *testing.T) {
    cfg := &Config{
        App:      AppConfig{HTTPPort: 8080},
        Database: DatabaseConfig{Port: 5432, User: "test", Name: "testdb"},
    }

    err := cfg.Validate()
    require.Error(t, err)
    assert.Contains(t, err.Error(), "DB_HOST is required")
}

func TestValidate_InvalidPort(t *testing.T) {
    cfg := &Config{
        App:      AppConfig{HTTPPort: -1},
        Database: DatabaseConfig{Host: "localhost", Port: 5432, User: "test", Name: "testdb"},
    }

    err := cfg.Validate()
    require.Error(t, err)
    assert.Contains(t, err.Error(), "APP_HTTP_PORT must be between 1 and 65535")
}

func TestValidate_MultipleErrors(t *testing.T) {
    cfg := &Config{} // All empty/zero

    err := cfg.Validate()
    require.Error(t, err)

    // Should collect ALL errors
    validErr, ok := err.(*ValidationError)
    require.True(t, ok)
    assert.GreaterOrEqual(t, len(validErr.Errors), 4) // At least 4 required fields
}

func TestValidate_ValidConfig(t *testing.T) {
    cfg := &Config{
        App: AppConfig{
            Name:     "test-app",
            Env:      "development",
            HTTPPort: 8080,
        },
        Database: DatabaseConfig{
            Host: "localhost",
            Port: 5432,
            User: "postgres",
            Name: "testdb",
        },
    }

    err := cfg.Validate()
    assert.NoError(t, err)
}
```

### Previous Story Learnings

From **Story 2.2** code review:
- ✅ Use `0600` permissions for any file operations
- ✅ Add `os.Stat` check for clearer file errors
- ✅ Keep coverage high (currently 90.9%)

From **Story 2.1** code review:
- ✅ Use map for prefixes to avoid DRY violation
- ✅ Test zero values for unset environment variables

### Integration Points

1. **loader.go modification:**
   - Add `Validate()` call after unmarshal
   - Wrap validation error with context

2. **Error propagation:**
   - ValidationError implements `error` interface
   - Contains all errors for developer visibility

### Quality Requirements

- Cyclomatic complexity ≤ 15 (Validate function simple)
- Test coverage ≥ 90%
- Lint: 0 issues
- Error messages match AC format exactly

### References

- [Source: docs/epics.md#Story-2.3]
- [Source: docs/architecture.md#Error-Handling]
- [Story 2.1 - Config struct fields](file:///docs/sprint-artifacts/2-1-implement-environment-variable-loading.md)
- [Story 2.2 - File loading patterns](file:///docs/sprint-artifacts/2-2-add-optional-config-file-support.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Builds on loader.go from Stories 2.1 and 2.2.
Validation follows NFR9: Config validation fails fast.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11
- Ultimate context engine analysis completed
- Validation improvements applied: 2025-12-11
- Implementation completed: 2025-12-11
- Code review fixes applied: 2025-12-11
  - MEDIUM: Moved validAppEnvs to package level for efficiency
  - MEDIUM: Added documentation comment explaining Password not required
  - Coverage: 98.1% (maintained)
  - Lint: 0 issues

### File List

Files created:
- `internal/config/validate.go` - Config validation logic (86 lines)
- `internal/config/validate_test.go` - Validation tests (275 lines)

Files modified:
- `internal/config/loader.go` - Added Validate() call after unmarshal (85 lines)
- `internal/config/loader_test.go` - Updated tests for validation requirements
- `internal/config/loader_file_test.go` - Updated tests for validation requirements
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
