# Story 11.3: Implement `bplat generate module` Command

Status: Done

## Story

As a developer,
I want to generate new domain module,
So that I can add features following patterns.

## Acceptance Criteria

1. **Given** I run `bplat generate module payment`
   **When** command completes
   **Then** domain/usecase/infra/interface layers are created

2. **Given** the generate command runs successfully
   **When** I check the generated files
   **Then** migration template is created

3. **Given** the generate command runs successfully
   **When** I check the generated files
   **Then** sqlc query file is created

4. **Given** the generate command runs successfully
   **When** I check the generated files
   **Then** test files with templates are created

## Tasks / Subtasks

- [x] Task 1: Create Generate Command Structure (AC: #1)
  - [x] Create `cmd/bplat/cmd/generate.go` with `generate` parent command
  - [x] Create `cmd/bplat/cmd/generate_module.go` with `generate module <name>` subcommand
  - [x] Register generate command in `root.go` via `init()` function
  - [x] Add flags: `--entity` (entity name, default: singular of module name)
  
- [x] Task 2: Create Module Template Embed System (AC: #1, #2, #3, #4)
  - [x] Use `embed.FS` with `//go:embed all:templates/module` directive
  - [x] Create template file structure mirroring hexagonal architecture layers
  - [x] Follow existing `templates/service/` pattern from Story 11.2
  
- [x] Task 3: Implement Domain Layer Generation (AC: #1)
  - [x] Create `internal/domain/{name}/entity.go.tmpl` - Entity with Validate()
  - [x] Create `internal/domain/{name}/errors.go.tmpl` - Domain-specific errors
  - [x] Create `internal/domain/{name}/repository.go.tmpl` - Repository interface
  - [x] Create `internal/domain/{name}/entity_test.go.tmpl` - Entity tests

- [x] Task 4: Implement Use Case Layer Generation (AC: #1)
  - [x] Create `internal/usecase/{name}/usecase.go.tmpl` - Business logic
  - [x] Create `internal/usecase/{name}/usecase_test.go.tmpl` - Unit tests with mocks

- [x] Task 5: Implement Interface Layer Generation (AC: #1, #4)
  - [x] Create `internal/interface/http/{name}/handler.go.tmpl` - HTTP handlers
  - [x] Create `internal/interface/http/{name}/dto.go.tmpl` - Request/Response DTOs
  - [x] Create `internal/interface/http/{name}/handler_test.go.tmpl` - Handler tests

- [x] Task 6: Implement Infrastructure Layer Generation (AC: #2, #3)
  - [x] Create `db/migrations/{timestamp}_{name}.up.sql.tmpl` - Up migration template
  - [x] Create `db/migrations/{timestamp}_{name}.down.sql.tmpl` - Down migration template  
  - [x] Create `db/queries/{name}.sql.tmpl` - sqlc query template

- [x] Task 7: Implement Scaffolding Logic (AC: #1, #2, #3, #4)
  - [x] Validate module name (same rules as service name from 11.2)
  - [x] Check for existing module and provide warning
  - [x] Generate timestamp for migration files
  - [x] Walk embedded templates and create files
  - [x] Apply text/template processing with TemplateData
  
- [x] Task 8: Write Tests (AC: #1, #2, #3, #4)
  - [x] Create `cmd/bplat/cmd/generate_test.go` with table-driven tests
  - [x] Test directory structure creation
  - [x] Test all layer files are generated
  - [x] Test migration timestamp format
  - [x] Test error cases (invalid name, existing module)

- [x] Task 9: Update Documentation
  - [x] Update AGENTS.md with generate command patterns
  - [x] Add usage examples to README.md CLI section

## Dev Notes

### Architecture Requirements

- **Location:** `cmd/bplat/cmd/` following cobra convention from Story 11.1, 11.2
- **Templates:** Use Go's `embed.FS` for compile-time embedding (proven pattern from 11.2)
- **Processing:** Use `text/template` for variable replacement
- **Output:** Files are generated in the CURRENT project, not a new directory

### File Structure

```
cmd/bplat/cmd/
├── generate.go           # NEW - generate parent command
├── generate_module.go    # NEW - generate module subcommand
├── generate_test.go      # NEW - generate tests
└── templates/
    ├── service/          # Existing from 11.2
    └── module/           # NEW - module templates
        ├── domain/
        │   └── {name}/
        │       ├── entity.go.tmpl
        │       ├── errors.go.tmpl
        │       ├── repository.go.tmpl
        │       └── entity_test.go.tmpl
        ├── usecase/
        │   └── {name}/
        │       ├── usecase.go.tmpl
        │       └── usecase_test.go.tmpl
        ├── interface/
        │   └── http/
        │       └── {name}/
        │           ├── handler.go.tmpl
        │           ├── dto.go.tmpl
        │           └── handler_test.go.tmpl
        └── db/
            ├── migrations/
            │   ├── up.sql.tmpl
            │   └── down.sql.tmpl
            └── queries/
                └── queries.sql.tmpl
```

### Implementation Guidelines

#### Generate Command Structure

```go
package cmd

import "github.com/spf13/cobra"

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate new components",
    Long:  `Generate new domain modules, handlers, or other components.`,
}

func init() {
    rootCmd.AddCommand(generateCmd)
}
```

#### Generate Module Subcommand Pattern

```go
package cmd

import (
    "embed"
    "fmt"
    "time"
    
    "github.com/spf13/cobra"
)

//go:embed all:templates/module
var moduleTemplates embed.FS

var (
    entityName string
)

var generateModuleCmd = &cobra.Command{
    Use:   "module <name>",
    Short: "Generate a new domain module",
    Long:  `Generate domain/usecase/infra/interface layers for a new module.`,
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        moduleName := args[0]
        if err := validateModuleName(moduleName); err != nil {
            return err
        }
        return scaffoldModule(cmd, moduleName, entityName)
    },
}

func init() {
    generateCmd.AddCommand(generateModuleCmd)
    generateModuleCmd.Flags().StringVarP(&entityName, "entity", "e", "", "Entity name (default: singularized module name)")
}
```

#### TemplateData for Module Generation

```go
type ModuleTemplateData struct {
    ModuleName   string // lowercase, e.g., "payment"
    EntityName   string // PascalCase, e.g., "Payment"
    TableName    string // snake_case plural, e.g., "payments"
    Timestamp    string // migration timestamp, e.g., "20251214021630"
    ModulePath   string // full module path, e.g., "github.com/iruldev/golang-api-hexagonal"
}
```

#### Template File Examples

**entity.go.tmpl:**
```go
package {{.ModuleName}}

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

// {{.EntityName}} represents a {{.ModuleName}} entity
type {{.EntityName}} struct {
    ID        uuid.UUID
    // TODO: Add fields
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Validate validates the {{.EntityName}} entity
func (e *{{.EntityName}}) Validate() error {
    if e.ID == uuid.Nil {
        return errors.New("id is required")
    }
    // TODO: Add validation
    return nil
}
```

**repository.go.tmpl:**
```go
package {{.ModuleName}}

import (
    "context"

    "github.com/google/uuid"
)

// Repository defines the {{.EntityName}} repository interface
type Repository interface {
    Create(ctx context.Context, entity *{{.EntityName}}) error
    GetByID(ctx context.Context, id uuid.UUID) (*{{.EntityName}}, error)
    Update(ctx context.Context, entity *{{.EntityName}}) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*{{.EntityName}}, error)
}
```

**up.sql.tmpl:**
```sql
-- Migration: Create {{.TableName}} table
CREATE TABLE {{.TableName}} (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- TODO: Add columns
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add updated_at trigger
CREATE TRIGGER set_{{.TableName}}_updated_at
    BEFORE UPDATE ON {{.TableName}}
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**queries.sql.tmpl:**
```sql
-- name: Create{{.EntityName}} :one
INSERT INTO {{.TableName}} (
    -- TODO: Add columns
) VALUES (
    -- TODO: Add values
) RETURNING *;

-- name: Get{{.EntityName}}ByID :one
SELECT * FROM {{.TableName}} WHERE id = $1;

-- name: List{{.EntityName}}s :many
SELECT * FROM {{.TableName}} ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: Update{{.EntityName}} :one
UPDATE {{.TableName}} SET
    updated_at = NOW()
    -- TODO: Add columns
WHERE id = $1
RETURNING *;

-- name: Delete{{.EntityName}} :exec
DELETE FROM {{.TableName}} WHERE id = $1;
```

### Previous Story Learnings (Story 11.1 & 11.2)

- Use `embed.FS` with `//go:embed all:templates/...` directive (NOT `*` pattern!)
- Templates directory must be inside package directory (e.g., `cmd/bplat/cmd/templates/`)
- Use factory functions for testable commands (`newTestRootCmd()` pattern)
- Use `cmd.OutOrStdout()` for testable output
- Follow table-driven test pattern with AAA structure
- Avoid global state mutation in tests - removed `t.Parallel()` for tests mutating global vars
- Use `fs.WalkDir` for walking embedded templates
- Process `.tmpl` files with text/template, copy others directly
- Skip `.gitkeep` files during scaffolding

### Testing Guidelines

#### Test Factory Pattern (from 11.1, 11.2)
```go
func newTestGenerateCmd() *cobra.Command {
    gen := &cobra.Command{Use: "generate"}
    gen.AddCommand(newGenerateModuleCmd())
    return gen
}
```

#### Test Cases to Cover
- Valid module name creation
- Invalid module name (same validation as service name)
- All layer files are generated correctly
- Migration files have correct timestamp format
- sqlc queries file is generated
- Test templates are generated
- Entity name customization via --entity flag
- Template variable replacement works

### Project Structure Notes

- Follows cobra convention with nested subcommands (`bplat generate module`)
- Uses same module path `github.com/iruldev/golang-api-hexagonal`
- Templates are embedded at compile time for single binary distribution
- Generated files go into the CURRENT project directory structure

### Validation Rules

| Rule | Pattern | Error Message |
|------|---------|---------------|
| Name empty | `len(name) == 0` | "module name is required" |
| Invalid chars | `regexp.MustCompile("^[a-z][a-z0-9_-]*$")` | "module name must start with letter and contain only lowercase letters, numbers, hyphens, underscores" |

### References

- [Source: docs/epics.md#Story-11.3]
- [Source: cmd/bplat/cmd/init_service.go - template scaffolding pattern from Story 11.2]
- [Source: AGENTS.md#Adding-a-New-Domain - existing manual module creation guide]
- [Source: internal/domain/note/ - example domain layer structure]
- [Source: db/queries/note.sql - example sqlc queries]
- [Go embed documentation: https://pkg.go.dev/embed]

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created
- Created `generate.go` parent command following cobra pattern
- Created `generate_module.go` with scaffolding logic, validation, helper functions
- Created 12 module template files covering all hexagonal architecture layers
- Comprehensive test suite with 15+ test cases (all pass)
- Updated `root_test.go` factory function to include generate command
- Added comprehensive generate module documentation to AGENTS.md
- All 21 test packages pass, 0 regressions

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | SM Agent | Story created with comprehensive context from 11.1 and 11.2 learnings |
| 2025-12-14 | Dev Agent | Implemented all 9 tasks, tests pass, documentation complete |
| 2025-12-14 | Code Review | Enhanced tests with content verification, marked done |

### File List

Files created:
- `cmd/bplat/cmd/generate.go` - Generate parent command
- `cmd/bplat/cmd/generate_module.go` - Generate module subcommand with scaffolding
- `cmd/bplat/cmd/generate_test.go` - Comprehensive test suite (15+ tests)
- `cmd/bplat/cmd/templates/module/internal/domain/{name}/entity.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/domain/{name}/errors.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/domain/{name}/repository.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/domain/{name}/entity_test.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/usecase/{name}/usecase.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/usecase/{name}/usecase_test.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/interface/http/{name}/handler.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/interface/http/{name}/dto.go.tmpl`
- `cmd/bplat/cmd/templates/module/internal/interface/http/{name}/handler_test.go.tmpl`
- `cmd/bplat/cmd/templates/module/db/migrations/{timestamp}_{name}.up.sql.tmpl`
- `cmd/bplat/cmd/templates/module/db/migrations/{timestamp}_{name}.down.sql.tmpl`
- `cmd/bplat/cmd/templates/module/db/queries/{name}.sql.tmpl`

Files modified:
- `cmd/bplat/cmd/root_test.go` - Added newGenerateCmd() to test factory
- `AGENTS.md` - Added generate module command documentation
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
