# Story 11.2: Implement `bplat init service` Command

Status: Done

## Story

As a developer,
I want to initialize a new service from template,
So that I can start new projects quickly.

## Acceptance Criteria

1. **Given** I run `bplat init service myservice`
   **When** command completes
   **Then** new directory with complete structure is created

2. **Given** the init command runs successfully
   **When** I check the generated directory
   **Then** go.mod is updated with correct module name

3. **Given** the init command runs successfully
   **When** I check the generated directory
   **Then** README is customized with service name

## Tasks / Subtasks

- [x] Task 1: Create Init Command Structure (AC: #1)
  - [x] Create `cmd/bplat/cmd/init.go` with `init` subcommand
  - [x] Create `cmd/bplat/cmd/init_service.go` with `init service <name>` subcommand
  - [x] Register init command in `root.go` via `init()` function
  - [x] Add flags: `--module` (optional module path override), `--dir` (output directory)
  
- [x] Task 2: Create Template Embed System (AC: #1, #2, #3)
  - [x] Use `embed.FS` with `//go:embed all:templates/service` directive
  - [x] Use `text/template` for template processing
  - [x] Create template file structure mirroring project layout in `cmd/bplat/cmd/templates/service/`
  
- [x] Task 3: Implement Template Processing (AC: #1, #2, #3)
  - [x] Implement template variable replacement (service name, module path)
  - [x] Handle go.mod generation with correct module name
  - [x] Process README.md with service name customization
  
- [x] Task 4: Implement Directory Scaffolding (AC: #1)
  - [x] Walk embedded templates and create directory structure
  - [x] Copy non-template files directly
  - [x] Process `.tmpl` files with text/template
  - [x] Handle file permissions correctly
  
- [x] Task 5: Add Validation and Error Handling (AC: #1, #2, #3)
  - [x] Validate service name (alphanumeric, no spaces, valid Go identifier)
  - [x] Check if target directory already exists
  - [x] Provide clear error messages on failure
  - [x] Add `--force` flag to overwrite existing directory
  
- [x] Task 6: Write Tests (AC: #1, #2, #3)
  - [x] Create `cmd/bplat/cmd/init_test.go` with table-driven tests
  - [x] Test directory creation and file generation
  - [x] Test go.mod module name replacement
  - [x] Test README customization
  - [x] Test error cases (invalid name, existing dir)
  
- [x] Task 7: Update Documentation
  - [x] Update AGENTS.md with init command patterns
  - [x] Add usage examples to README.md CLI section

## Dev Notes

### Architecture Requirements

- **Location:** `cmd/bplat/cmd/` following cobra convention from Story 11.1
- **Templates:** Use Go's `embed.FS` for compile-time embedding
- **Processing:** Use `text/template` for variable replacement

### File Structure

```
cmd/bplat/
├── main.go
├── cmd/
│   ├── root.go
│   ├── root_test.go
│   ├── version.go
│   ├── version_test.go
│   ├── init.go          # NEW - init parent command
│   ├── init_service.go  # NEW - init service subcommand
│   └── init_test.go     # NEW - init tests
└── templates/           # NEW - embedded templates
    └── service/
        ├── cmd/
        │   └── server/
        │       └── main.go.tmpl
        ├── internal/
        │   └── app/
        │       └── app.go.tmpl
        ├── go.mod.tmpl
        ├── README.md.tmpl
        └── .env.example
```

### Implementation Guidelines

#### Init Command Structure

```go
package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize new components",
    Long:  `Initialize new service or module from templates.`,
}

func init() {
    rootCmd.AddCommand(initCmd)
}
```

#### Init Service Subcommand Pattern

```go
package cmd

import (
    "embed"
    "fmt"
    "os"
    "path/filepath"
    "text/template"
    
    "github.com/spf13/cobra"
)

//go:embed templates/service/*
var serviceTemplates embed.FS

var (
    modulePath string
    outputDir  string
    force      bool
)

var initServiceCmd = &cobra.Command{
    Use:   "service <name>",
    Short: "Initialize a new service",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        serviceName := args[0]
        if err := validateServiceName(serviceName); err != nil {
            return err
        }
        return scaffoldService(serviceName, modulePath, outputDir)
    },
}

func init() {
    initCmd.AddCommand(initServiceCmd)
    initServiceCmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (default: github.com/<user>/<service>)")
    initServiceCmd.Flags().StringVarP(&outputDir, "dir", "d", ".", "Output directory")
    initServiceCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing directory")
}
```

#### Template Processing Pattern

```go
type TemplateData struct {
    ServiceName string
    ModulePath  string
    GoVersion   string
}

func processTemplate(tmplContent string, data TemplateData) (string, error) {
    tmpl, err := template.New("").Parse(tmplContent)
    if err != nil {
        return "", fmt.Errorf("parse template: %w", err)
    }
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("execute template: %w", err)
    }
    return buf.String(), nil
}
```

#### Template File Examples

**go.mod.tmpl:**
```
module {{.ModulePath}}

go {{.GoVersion}}
```

**README.md.tmpl:**
```markdown
# {{.ServiceName}}

Service generated with bplat CLI.

## Quick Start

```bash
make dev
```
```

### Previous Story Learnings (Story 11.1)

- Use factory functions for testable commands (`newTestRootCmd()` pattern)
- Use `cmd.OutOrStdout()` for testable output
- Follow table-driven test pattern with AAA structure
- Avoid global state mutation in tests - removed `t.Parallel()` for tests mutating global vars
- Handle errors from fmt.Fprintf properly
- Register commands in `init()` function via `rootCmd.AddCommand()`

### Testing Guidelines

#### Test Factory Pattern (from 11.1)
```go
// Create fresh command for each test to avoid global state issues
func newTestInitCmd() *cobra.Command {
    init := &cobra.Command{Use: "init"}
    init.AddCommand(newInitServiceCmd())
    return init
}
```

#### Test Cases to Cover
- Valid service name creation
- Invalid service name (spaces, special chars)
- Existing directory without force flag → error
- Existing directory with force flag → overwrite
- go.mod generated with correct module path
- README.md contains service name
- Template variable replacement works

### Project Structure Notes

- Follows cobra convention with nested subcommands (`bplat init service`)
- Uses same module path `github.com/iruldev/golang-api-hexagonal`
- Templates are embedded at compile time for single binary distribution

### Validation Rules

| Rule | Pattern | Error Message |
|------|---------|---------------|
| Name empty | `len(name) == 0` | "service name is required" |
| Invalid chars | `regexp.MustCompile("^[a-z][a-z0-9_-]*$")` | "service name must start with letter and contain only lowercase letters, numbers, hyphens, underscores" |
| Dir exists | `os.Stat(dir)` | "directory already exists, use --force to overwrite" |

### References

- [Source: docs/epics.md#Story-11.2]
- [Source: cmd/bplat/cmd/root.go - existing CLI patterns from Story 11.1]
- [Source: cmd/bplat/cmd/version.go - command implementation pattern]
- [Go embed documentation: https://pkg.go.dev/embed]
- [Cobra subcommands: https://github.com/spf13/cobra#subcommands]

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created
- Created `cmd/bplat/cmd/init.go` with init parent command following cobra convention
- Created `cmd/bplat/cmd/init_service.go` with service subcommand, validation, scaffolding
- Implemented TemplateData struct with ServiceName, ModulePath, GoVersion
- Implemented validateServiceName() with regex pattern `^[a-z][a-z0-9_-]*$`
- Implemented scaffoldService() with directory creation and template processing using `fs.WalkDir`
- Added flags: `--module`, `--dir`, `--force` for flexibility
- Created comprehensive init_test.go with 22 test cases (all passing)
- Updated AGENTS.md with init command documentation, flags, and examples
- Updated README.md CLI section with init service examples
- Updated root_test.go factory function to include init command
- **[Code Review Fix]** Refactored from hardcoded string constants to proper `embed.FS` with `//go:embed all:templates/service` directive
- **[Code Review Fix]** Created `templates/service/` directory with `.tmpl` files for better maintainability

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | SM Agent | Story created with comprehensive context from 11.1 learnings |
| 2025-12-14 | Dev Agent | Implemented init command structure, templates, validation, and tests |
| 2025-12-14 | Code Review | Refactored to `embed.FS` pattern per architecture requirements |

### File List

- `cmd/bplat/cmd/init.go` - NEW - Init parent command
- `cmd/bplat/cmd/init_service.go` - NEW/MODIFIED - Init service subcommand using embed.FS
- `cmd/bplat/cmd/init_test.go` - NEW - Comprehensive test suite (22 tests)
- `cmd/bplat/cmd/root_test.go` - MODIFIED - Added newInitCmd() to test factory
- `cmd/bplat/cmd/templates/service/go.mod.tmpl` - NEW - Go module template
- `cmd/bplat/cmd/templates/service/README.md.tmpl` - NEW - README template
- `cmd/bplat/cmd/templates/service/.env.example.tmpl` - NEW - Environment template
- `cmd/bplat/cmd/templates/service/cmd/server/main.go.tmpl` - NEW - Main entry template
- `cmd/bplat/cmd/templates/service/internal/app/app.go.tmpl` - NEW - App struct template
- `cmd/bplat/cmd/templates/service/db/migrations/.gitkeep` - NEW - Directory placeholder
- `cmd/bplat/cmd/templates/service/db/queries/.gitkeep` - NEW - Directory placeholder
- `AGENTS.md` - MODIFIED - Added init command documentation with flags and examples
- `README.md` - MODIFIED - Added init service examples to CLI section

