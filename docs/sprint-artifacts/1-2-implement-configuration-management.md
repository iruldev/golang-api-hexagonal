# Story 1.2: Implement Configuration Management

Status: done

## Story

**As a** developer,
**I want** environment-based configuration with validation and sensible defaults,
**So that** the service fails fast with clear errors if misconfigured.

## Acceptance Criteria

1. **Given** I start the service without required configuration (e.g., DATABASE_URL missing)
   **When** the service attempts to load configuration
   **Then** the service exits with non-zero exit code
   **And** the error message clearly indicates which configuration is missing (e.g., "required key DATABASE_URL missing value")

2. **Given** I start the service with valid configuration
   **When** the service loads configuration
   **Then** defaults are applied for optional values:
   - `PORT` defaults to `8080`
   - `LOG_LEVEL` defaults to `info`
   - `ENV` defaults to `development`
   - `SERVICE_NAME` defaults to `golang-api-hexagonal`
   **And** `.env.example` exists with all configurable options documented
   **And** configuration loader uses `kelseyhightower/envconfig` package

## Tasks / Subtasks

- [x] Task 1: Create Config struct (AC: #2)
  - [x] Create `internal/infra/config/config.go`
  - [x] Define Config struct with all fields (DatabaseURL, Port, LogLevel, Env, ServiceName)
  - [x] Add struct tags for envconfig (`envconfig:"PORT" default:"8080"`)
  - [x] Add struct tags for required fields (`required:"true"` for DATABASE_URL)

- [x] Task 2: Implement config loading function (AC: #1, #2)
  - [x] Create `Load()` function that returns `(*Config, error)`
  - [x] Use `envconfig.Process()` to parse environment variables
  - [x] Return wrapped error with clear message on failure

- [x] Task 3: Create .env.example file (AC: #2)
  - [x] Create `.env.example` in project root
  - [x] Document all configuration options with descriptions
  - [x] Include example values and defaults

- [x] Task 4: Integrate config into main.go (AC: #1)
  - [x] Import config package
  - [x] Call `config.Load()` at startup
  - [x] Exit with non-zero code on config error
  - [x] Print loaded config summary (without sensitive values)

- [x] Task 5: Write unit tests for config (AC: #1, #2)
  - [x] Test successful config loading with all required fields
  - [x] Test default values are applied
  - [x] Test missing required field returns error
  - [x] Test error message contains field name

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

Config package is in **Infra layer** (`internal/infra/config/`):
- ✅ ALLOWED: domain imports, external packages (envconfig)
- ❌ FORBIDDEN: app, transport imports

### Technology Stack [Source: docs/project-context.md]

| Component | Package | Version |
|-----------|---------|---------|
| Config | github.com/kelseyhightower/envconfig | v1.4.0 |

### Config Struct Pattern [Source: docs/architecture.md]

```go
// internal/infra/config/config.go
package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
    // Required fields
    DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
    
    // Optional with defaults
    Port        int    `envconfig:"PORT" default:"8080"`
    LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
    Env         string `envconfig:"ENV" default:"development"`
    ServiceName string `envconfig:"SERVICE_NAME" default:"golang-api-hexagonal"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, fmt.Errorf("config.Load: %w", err)
    }
    return &cfg, nil
}
```

### Error Wrapping Pattern [Source: docs/project-context.md]

Infra layer uses `op` string pattern:
```go
return nil, fmt.Errorf("config.Load: %w", err)
```

### .env.example Template

```bash
# Required
DATABASE_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable

# Optional (defaults shown)
PORT=8080
LOG_LEVEL=info
ENV=development
SERVICE_NAME=golang-api-hexagonal
```

### Testing Pattern [Source: docs/project-context.md]

Table-driven tests with testify:
```go
func TestLoad(t *testing.T) {
    tests := []struct {
        name    string
        envVars map[string]string
        want    *Config
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Set env vars, call Load(), assert
        })
    }
}
```

### Previous Story Learnings [Source: Story 1.1]

- Project structure is established with hexagonal layers
- `go.mod` initialized with module path `github.com/iruldev/golang-api-hexagonal`
- Use `.keep` files for empty directories
- No conflicts with the hexagonal architecture

## Technical Requirements

- **Go version:** 1.23+ [Source: docs/project-context.md]
- **Package:** github.com/kelseyhightower/envconfig (must add to go.mod)
- **Test coverage:** Unit tests for all config scenarios

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- Infra layer wraps errors with `op` string pattern
- NO logging in domain/app layers
- Config is loaded in main.go, passed to dependencies

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-16)

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- `go mod tidy` - SUCCESS (added envconfig v1.4.0, testify v1.11.1)
- `go test -v ./internal/infra/config/...` - SUCCESS (7/7 tests pass)
- `go build ./...` - SUCCESS

### Completion Notes List

- [x] Config struct created with envconfig tags
- [x] Load() function implemented with error wrapping
- [x] Config validation added (ENV/LOG_LEVEL allowlist, PORT range, non-empty SERVICE_NAME)
- [x] .env.example created with all options documented
- [x] main.go updated to load config and exit on error
- [x] Unit tests pass (7 tests, 100% scenarios covered)
- [x] go build ./... passes

### File List

Files created/modified:
- `internal/infra/config/config.go` (NEW)
- `internal/infra/config/config_test.go` (NEW)
- `.env.example` (NEW)
- `cmd/api/main.go` (MODIFIED)
- `docs/sprint-artifacts/sprint-status.yaml` (MODIFIED - story tracking)
- `docs/sprint-artifacts/1-2-implement-configuration-management.md` (MODIFIED - review + status updates)
- `go.mod` (MODIFIED - added envconfig, testify)
- `go.sum` (NEW)

### Change Log

- 2025-12-16: Story 1.2 implemented - Configuration management with envconfig, .env.example, main.go integration, and comprehensive unit tests
- 2025-12-16: Senior code review performed - validation tightened, tests improved, status synced
- 2025-12-16: Review follow-up - .env.example clarified (no auto `.env` loading), ENV docs synced, SERVICE_NAME validation test added, `go test ./...` PASS

## Senior Developer Review (AI)

Reviewer: Chat
Date: 2025-12-16

### Summary

- Outcome: Approved (semua High/Medium sudah dibereskan)
- Verification: `go test ./...` PASS (2025-12-16)

### Fixes Applied

- Added explicit validation beyond env parsing: `PORT` range, `ENV` allowlist, `LOG_LEVEL` allowlist, `SERVICE_NAME` not empty
- Updated startup output to use `SERVICE_NAME` (no hardcoded service name)
- Hardened config tests using `t.Setenv` and added coverage for invalid ENV/LOG_LEVEL/PORT range
- Perbaikan dokumentasi & test: `.env.example` diperjelas (service tidak auto-load `.env`), ENV docs diselaraskan, test untuk SERVICE_NAME kosong/whitespace ditambahkan
