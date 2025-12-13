# Story 11.1: Create CLI Tool Structure (bplat)

Status: Done

## Story

As a developer,
I want a CLI tool for boilerplate operations,
So that I can scaffold code quickly.

## Acceptance Criteria

1. **Given** `cmd/bplat/main.go` exists
   **When** I run `bplat --help`
   **Then** available commands are listed

2. **Given** the CLI is properly configured
   **When** I run `bplat version` or `bplat --version`
   **Then** current version is displayed

3. **Given** the CLI follows cobra convention
   **When** I review the code structure
   **Then** commands are organized in `cmd/bplat/cmd/` directory

## Tasks / Subtasks

- [x] Task 1: Create CLI Entry Point (AC: #1, #2, #3)
  - [x] Create `cmd/bplat/main.go` with cobra root command
  - [x] Create `cmd/bplat/cmd/root.go` with root command and help
  - [x] Add version flag/command showing semantic version
  
- [x] Task 2: Add Cobra Dependency (AC: #3)
  - [x] Add `github.com/spf13/cobra` to go.mod
  - [x] Run `go mod tidy` to update dependencies

- [x] Task 3: Implement Root Command (AC: #1)
  - [x] Define root command with description
  - [x] Add `--version` flag
  - [x] Add placeholder for subcommands

- [x] Task 4: Implement Version Command (AC: #2)
  - [x] Create `cmd/bplat/cmd/version.go`
  - [x] Display version, build date, and Go version
  - [x] Use ldflags for build-time injection

- [x] Task 5: Write Tests (AC: #1, #2, #3)
  - [x] Create `cmd/bplat/cmd/root_test.go`
  - [x] Create `cmd/bplat/cmd/version_test.go`
  - [x] Test help output format
  - [x] Test version output format

- [x] Task 6: Update Makefile (AC: #1, #2)
  - [x] Add `build-bplat` target
  - [x] Add ldflags for version injection
  - [x] Add `install-bplat` target

- [x] Task 7: Update Documentation
  - [x] Update AGENTS.md with CLI patterns
  - [x] Add CLI section to README.md

## Dev Notes

### Architecture Requirements

- **Location:** `cmd/bplat/` following existing pattern (`cmd/server/`, `cmd/worker/`)
- **Package structure:** Cobra convention with `cmd/` subdirectory for commands
- **Dependencies:** Use `github.com/spf13/cobra` (industry standard)

### File Structure

```
cmd/bplat/
├── main.go           # Entry point, calls cmd.Execute()
└── cmd/
    ├── root.go       # Root command with Execute() function
    ├── root_test.go  # Root command tests
    ├── version.go    # Version command
    └── version_test.go
```

### Implementation Guidelines

#### main.go Pattern
```go
package main

import "github.com/iruldev/golang-api-hexagonal/cmd/bplat/cmd"

func main() {
    cmd.Execute()
}
```

#### Root Command Pattern
```go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "bplat",
    Short: "Boilerplate CLI tool for code scaffolding",
    Long: `bplat is a CLI tool for boilerplate operations.
It helps you scaffold new services and modules quickly.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

#### Version Pattern with ldflags
```go
// Variables set via ldflags at build time
var (
    Version   = "dev"
    BuildDate = "unknown"
    GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("bplat version %s\n", Version)
        fmt.Printf("Build date: %s\n", BuildDate)
        fmt.Printf("Git commit: %s\n", GitCommit)
        fmt.Printf("Go version: %s\n", runtime.Version())
    },
}
```

#### Makefile ldflags Pattern
```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD)

LDFLAGS := -ldflags "\
    -X 'github.com/iruldev/golang-api-hexagonal/cmd/bplat/cmd.Version=$(VERSION)' \
    -X 'github.com/iruldev/golang-api-hexagonal/cmd/bplat/cmd.BuildDate=$(BUILD_DATE)' \
    -X 'github.com/iruldev/golang-api-hexagonal/cmd/bplat/cmd.GitCommit=$(GIT_COMMIT)'"

.PHONY: build-bplat
build-bplat:
	go build $(LDFLAGS) -o bin/bplat ./cmd/bplat

.PHONY: install-bplat
install-bplat:
	go install $(LDFLAGS) ./cmd/bplat
```

### Testing Guidelines

#### Table-Driven Test Pattern
```go
func TestRootCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantExit int
        wantOut  string
    }{
        {
            name:     "help flag",
            args:     []string{"--help"},
            wantExit: 0,
            wantOut:  "bplat is a CLI tool",
        },
        {
            name:     "unknown command",
            args:     []string{"unknown"},
            wantExit: 1,
            wantOut:  "unknown command",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Previous Story Learnings (Epic 10)

- Use functional options pattern for configuration
- Always include example tests for API documentation
- Follow testify assertions for consistency
- Update AGENTS.md with new patterns

### Project Structure Notes

- Follows existing `cmd/` layout with separate subdirectories
- Uses same module path `github.com/iruldev/golang-api-hexagonal`
- Cobra is not yet in go.mod - needs to be added

### References

- [Source: docs/epics.md#Story-11.1]
- [Source: docs/architecture.md#Makefile]
- [Source: cmd/server/main.go - main entry point pattern]
- [Cobra documentation: https://github.com/spf13/cobra]

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created
- Created CLI entry point `cmd/bplat/main.go` with cobra root command
- Created `cmd/bplat/cmd/root.go` with Execute() function and version subcommand registration
- Created `cmd/bplat/cmd/version.go` with ldflags injection for Version, BuildDate, GitCommit
- Added comprehensive table-driven tests for root and version commands (6 tests passing)
- Added `build-bplat` and `install-bplat` Makefile targets with ldflags
- Updated AGENTS.md with CLI patterns section
- Updated README.md with CLI tool section and project structure
- **[Code Review]** Fixed test isolation by adding `newTestRootCmd()` factory function
- **[Code Review]** Added `newVersionCmd()` factory function for testable version commands
- **[Code Review]** Removed unsafe `os.Stdout` hijacking in tests
- **[Code Review]** Fixed data race by removing `t.Parallel()` from tests that mutate global state

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | SM Agent | Story created with comprehensive context |
| 2025-12-14 | Dev Agent | Implemented CLI structure with cobra, tests, Makefile, and docs |
| 2025-12-14 | Code Review | Fixed test isolation and data race issues

### File List

- `cmd/bplat/main.go` - NEW - CLI entry point
- `cmd/bplat/cmd/root.go` - NEW - Root command with Execute()
- `cmd/bplat/cmd/root_test.go` - NEW - Root command tests
- `cmd/bplat/cmd/version.go` - NEW - Version command with ldflags
- `cmd/bplat/cmd/version_test.go` - NEW - Version command tests
- `Makefile` - MODIFIED - Added build-bplat, install-bplat targets
- `AGENTS.md` - MODIFIED - Added CLI Tool section with patterns
- `README.md` - MODIFIED - Added CLI Tool section and updated project structure
- `go.mod` - MODIFIED - Added github.com/spf13/cobra dependency
