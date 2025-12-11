# Story 2.1: Implement Environment Variable Loading

Status: done

## Story

As a developer,
I want the system to load configuration from environment variables,
So that I can configure the service without modifying files.

## Acceptance Criteria

### AC1: Environment variables loaded into Config struct ✅
**Given** environment variables `APP_HTTP_PORT`, `DB_HOST`, `DB_PORT` are set
**When** the application starts
**Then** configuration is loaded from environment variables
**And** the Config struct is populated correctly

---

## Tasks / Subtasks

- [x] **Task 1: Create config package structure** (AC: #1)
  - [x] Note: `internal/config/doc.go` already exists from Story 1.1
  - [x] Create `internal/config/config.go` with Config struct
  - [x] Define nested structs: App, Database, Observability, Log
  - [x] Add struct tags for koanf binding (`koanf:"field_name"`)

- [x] **Task 2: Implement koanf loader** (AC: #1)
  - [x] Create `internal/config/loader.go`
  - [x] Initialize koanf with environment provider
  - [x] Use prefix-based loading (APP_, DB_, OTEL_, LOG_)
  - [x] Implement `Load() (*Config, error)` function

- [x] **Task 3: Create config tests** (AC: #1)
  - [x] Create `internal/config/loader_test.go`
  - [x] Test loading from environment variables
  - [x] Test partial env vars (zero values for unset)

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all tests pass (70.6% coverage)
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Config Struct Design

```go
// internal/config/config.go
package config

import "time"

// Config holds all application configuration.
type Config struct {
	App         AppConfig         `koanf:"app"`
	Database    DatabaseConfig    `koanf:"db"`
	Observability ObservabilityConfig `koanf:"otel"`
	Log         LogConfig         `koanf:"log"`
}

// AppConfig holds application settings.
type AppConfig struct {
	Name     string `koanf:"name"`
	Env      string `koanf:"env"`      // development, staging, production
	HTTPPort int    `koanf:"http_port"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string        `koanf:"host"`
	Port            int           `koanf:"port"`
	User            string        `koanf:"user"`
	Password        string        `koanf:"password"`
	Name            string        `koanf:"name"`
	SSLMode         string        `koanf:"ssl_mode"`
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
}

// ObservabilityConfig holds OpenTelemetry settings.
type ObservabilityConfig struct {
	ExporterEndpoint string `koanf:"exporter_otlp_endpoint"`
	ServiceName      string `koanf:"service_name"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"` // json, console
}
```

### Koanf Loader Implementation

```go
// internal/config/loader.go
package config

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/v2/providers/env"
)

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	k := koanf.New(".")

	// Load environment variables with prefix mapping
	// APP_HTTP_PORT -> app.http_port
	// DB_HOST -> db.host
	if err := k.Load(env.Provider("", "_", func(s string) string {
		// Map uppercase env vars to lowercase dotted paths
		s = strings.ToLower(s)
		// Map prefixes: APP_ -> app., DB_ -> db., OTEL_ -> otel., LOG_ -> log.
		for _, prefix := range []string{"app", "db", "otel", "log"} {
			if strings.HasPrefix(s, prefix+"_") {
				return prefix + "." + strings.TrimPrefix(s, prefix+"_")
			}
		}
		return s
	}), nil); err != nil {
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
```

### Environment Variable Mapping

| Env Var | Config Path | Type |
|---------|-------------|------|
| `APP_NAME` | `app.name` | string |
| `APP_ENV` | `app.env` | string |
| `APP_HTTP_PORT` | `app.http_port` | int |
| `DB_HOST` | `db.host` | string |
| `DB_PORT` | `db.port` | int |
| `DB_USER` | `db.user` | string |
| `DB_PASSWORD` | `db.password` | string |
| `DB_NAME` | `db.name` | string |
| `DB_SSL_MODE` | `db.ssl_mode` | string |
| `DB_MAX_OPEN_CONNS` | `db.max_open_conns` | int |
| `DB_MAX_IDLE_CONNS` | `db.max_idle_conns` | int |
| `DB_CONN_MAX_LIFETIME` | `db.conn_max_lifetime` | duration |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel.exporter_otlp_endpoint` | string |
| `OTEL_SERVICE_NAME` | `otel.service_name` | string |
| `LOG_LEVEL` | `log.level` | string |
| `LOG_FORMAT` | `log.format` | string |

### Dependencies to Add

```bash
go get github.com/knadh/koanf/v2
go get github.com/knadh/koanf/v2/providers/env
```

> **Note:** This story loads env vars but does NOT set defaults.
> Default values and validation are handled in Story 2.3 (fail-fast).

### Testing Strategy

Per project_context.md:
- Table-driven tests with `t.Run`
- Use testify (require/assert)
- `t.Parallel()` when safe
- Naming: `Test<Thing>_<Behavior>`

```go
// internal/config/loader_test.go
func TestLoad_FromEnvVars(t *testing.T) {
	t.Setenv("APP_HTTP_PORT", "9090")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.App.HTTPPort)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
}
```

### References

- [Source: docs/epics.md#Story-2.1]
- [Source: docs/project_context.md - knadh/koanf/v2]
- [Source: Story 1.3 - .env.example variables]
- [koanf GitHub](https://github.com/knadh/koanf)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11
- Validation applied: 2025-12-11
- Implementation completed: 2025-12-11
- Code review fixes applied: 2025-12-11
  - HIGH: Refactored loader.go with envPrefixes map (DRY fix)
  - MEDIUM: Added TestLoad_EmptyEnv test
  - Coverage improved: 70.6% → 80%
  - Lines reduced: 54 → 46 (-15%)

### File List

Files created:
- `internal/config/config.go` - Typed Config struct (44 lines)
- `internal/config/loader.go` - Koanf loader with prefix map (46 lines)
- `internal/config/loader_test.go` - Unit tests (94 lines, 3 tests)

Files modified:
- `go.mod` - Add koanf dependencies
- `go.sum` - Updated checksums
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

Files existing (no changes):
- `internal/config/doc.go` - Package doc (from Story 1.1)
