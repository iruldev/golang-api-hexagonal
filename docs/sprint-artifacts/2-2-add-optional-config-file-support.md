# Story 2.2: Add Optional Config File Support

Status: done

## Story

As a developer,
I want to optionally load configuration from a YAML/JSON file,
So that I can use config files in certain deployment scenarios.

## Acceptance Criteria

### AC1: Config file loaded when APP_CONFIG_FILE is set ✅
**Given** `APP_CONFIG_FILE` environment variable is set
**When** the application starts
**Then** configuration is loaded from the specified file
**And** environment variables override file values

### AC2: Only env vars used when no config file specified ✅
**Given** no config file is specified
**When** the application starts
**Then** only environment variables are used

---

## Tasks / Subtasks

- [x] **Task 1: Add koanf file providers** (AC: #1)
  - [x] Add `github.com/knadh/koanf/parsers/yaml` dependency
  - [x] Add `github.com/knadh/koanf/parsers/json` dependency
  - [x] Add `github.com/knadh/koanf/providers/file` dependency

- [x] **Task 2: Update loader.go for file support** (AC: #1, #2)
  - [x] Check for `APP_CONFIG_FILE` env var
  - [x] Load config file first if specified
  - [x] Detect file type from extension (.yaml, .yml, .json)
  - [x] Env vars loaded second to override file values
  - [x] Return error if file specified but not found

- [x] **Task 3: Create config tests for file loading** (AC: #1, #2)
  - [x] Create `internal/config/loader_file_test.go`
  - [x] Test loading from YAML file
  - [x] Test loading from JSON file
  - [x] Test env vars override file values
  - [x] Test no config file (uses only env vars)
  - [x] Test file not found error
  - [x] Test unsupported format error (bonus)

- [x] **Task 4: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all tests pass (90% coverage)
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Updated Loader Design

```go
// internal/config/loader.go (updated)
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Load loads configuration from optional file and environment variables.
// If APP_CONFIG_FILE is set, loads from that file first.
// Environment variables always override file values.
func Load() (*Config, error) {
	k := koanf.New(".")

	// Step 1: Load from config file if specified
	if configFile := os.Getenv("APP_CONFIG_FILE"); configFile != "" {
		if err := loadFromFile(k, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile, err)
		}
	}

	// Step 2: Load env vars (overrides file values)
	for prefix, path := range envPrefixes {
		if err := loadEnvPrefix(k, prefix, path); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadFromFile loads configuration from a YAML or JSON file.
func loadFromFile(k *koanf.Koanf, path string) error {
	ext := filepath.Ext(path)
	var parser koanf.Parser

	switch ext {
	case ".yaml", ".yml":
		parser = yaml.Parser()
	case ".json":
		parser = json.Parser()
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	return k.Load(file.Provider(path), parser)
}
```

### New Dependencies

```bash
go get github.com/knadh/koanf/parsers/yaml
go get github.com/knadh/koanf/parsers/json
go get github.com/knadh/koanf/providers/file
```

### Config File Example (config.yaml)

```yaml
app:
  name: my-service
  env: development
  http_port: 8080

db:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  name: mydb
  ssl_mode: disable

otel:
  exporter_otlp_endpoint: http://localhost:4317
  service_name: my-service

log:
  level: info
  format: console
```

### Override Behavior

| Priority | Source | Example |
|----------|--------|---------|
| 1 (highest) | Environment variable | `APP_HTTP_PORT=9090` |
| 2 | Config file | `http_port: 8080` in YAML |
| 3 (lowest) | Default (zero value) | 0 |

### Testing Strategy

```go
// internal/config/loader_file_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTempConfigFile creates a temporary config file for testing.
func createTempConfigFile(t *testing.T, ext, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config."+ext)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	return filePath
}

func TestLoad_FromYAMLFile(t *testing.T) {
	tmpFile := createTempConfigFile(t, "yaml", `
app:
  http_port: 8080
  name: test-from-file
`)
	t.Setenv("APP_CONFIG_FILE", tmpFile)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.App.HTTPPort)
	assert.Equal(t, "test-from-file", cfg.App.Name)
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	tmpFile := createTempConfigFile(t, "yaml", `
app:
  http_port: 8080
`)
	t.Setenv("APP_CONFIG_FILE", tmpFile)
	t.Setenv("APP_HTTP_PORT", "9090")  // Override!

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.App.HTTPPort)  // Env wins
}
```

### References

- [Source: docs/epics.md#Story-2.2]
- [Story 2.1 - Base loader implementation]
- [koanf file provider](https://github.com/knadh/koanf)
- [koanf yaml parser](https://github.com/knadh/koanf)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Depends on Story 2.1 (loader.go, envPrefixes map).

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11
- Validation applied: 2025-12-11
- Implementation completed: 2025-12-11
- Code review fixes applied: 2025-12-11
  - HIGH: Fixed test file permissions (0644→0600)
  - MEDIUM: Added os.Stat check for clearer errors
  - Coverage improved: 90% → 90.9%
  - loader.go now 80 lines

### File List

Files modified:
- `internal/config/loader.go` - File loading with os.Stat check (80 lines)
- `internal/config/loader_file_test.go` - Secure 0600 permissions
- `go.mod` - Added yaml/json parser dependencies
- `go.sum` - Updated checksums
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

Files created:
- `internal/config/loader_file_test.go` - File loading tests (110 lines)
