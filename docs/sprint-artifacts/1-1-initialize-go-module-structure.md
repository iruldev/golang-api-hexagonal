# Story 1.1: Initialize Go Module Structure

Status: done

## Story

As a developer,
I want to clone the boilerplate and see a proper Go module structure,
So that I can start working on domain logic immediately.

## Acceptance Criteria

### AC1: Dependencies Download ✅
**Given** a fresh clone of the repository
**When** I run `go mod download`
**Then** all dependencies are fetched successfully

### AC2: Project Compiles ✅
**Given** all dependencies are downloaded
**When** I run `go build ./...`
**Then** the project compiles with zero errors

### AC3: README Exists ✅
**Given** a fresh clone of the repository
**When** I look at the root directory
**Then** I see README.md with basic instructions

---

## Tasks / Subtasks

- [x] **Task 1: Create go.mod** (AC: #1, #2)
  - [x] Initialize with `go mod init github.com/iruldev/golang-api-hexagonal`
  - [x] Add core dependencies (see go.mod content below)
  - [x] Run `go mod tidy` to verify

- [x] **Task 2: Create project structure** (AC: #2)
  - [x] Create `cmd/server/main.go` (see main.go specifics)
  - [x] Create `internal/` directory structure per architecture
  - [x] Create `internal/interface/http/httpx/` for response helpers
  - [x] Create placeholder packages with doc.go files

- [x] **Task 3: Create doc.go files** (AC: #2)
  - [x] `internal/app/doc.go` - Application wiring
  - [x] `internal/config/doc.go` - Configuration
  - [x] `internal/domain/doc.go` - Business entities
  - [x] `internal/usecase/doc.go` - Business logic
  - [x] `internal/infra/doc.go` - Infrastructure adapters
  - [x] `internal/interface/doc.go` - HTTP handlers
  - [x] `internal/observability/doc.go` - Logging/tracing
  - [x] `internal/runtimeutil/doc.go` - Utilities

- [x] **Task 4: Create README.md** (AC: #3)
  - [x] Add project title and description
  - [x] Add quickstart: clone → `go mod download` → `go build ./...`
  - [x] Add link to docs/ARCHITECTURE.md

- [x] **Task 5: Verify compilation** (AC: #2)
  - [x] Run `go build ./...`
  - [x] Fix any import or compilation errors
  - [x] Confirm zero compile errors

---

## Dev Notes

### go.mod Content (EXACT)

```go
module github.com/iruldev/golang-api-hexagonal

go 1.24

require (
    github.com/go-chi/chi/v5 v5.1.0
    github.com/jackc/pgx/v5 v5.7.2
    go.uber.org/zap v1.27.0
    github.com/knadh/koanf/v2 v2.1.2
    go.opentelemetry.io/otel v1.33.0
    github.com/stretchr/testify v1.9.0
)
```

### main.go Specifics (cmd/server/main.go)

The entry point should be MINIMAL - just enough to compile:

```go
// Package main is the entry point for the backend service.
package main

import "fmt"

func main() {
    fmt.Println("Backend Service Golang Boilerplate")
    // Actual wiring comes in Story 1.2+
}
```

**DO NOT** add actual wiring yet - that requires config (Story 2.x) and HTTP (Story 3.x).

### Architecture Compliance

Per `docs/architecture.md` and `docs/project_context.md`:

```
Project Structure:
├── cmd/
│   └── server/
│       └── main.go          # Entry point (minimal)
├── internal/
│   ├── app/                 # Application wiring
│   │   └── doc.go
│   ├── config/              # Configuration loading (koanf)
│   │   └── doc.go
│   ├── domain/              # Business entities
│   │   ├── doc.go
│   │   └── note/            # Example domain (Story 7.x)
│   ├── usecase/             # Business logic
│   │   └── doc.go
│   ├── infra/               # Infrastructure
│   │   ├── doc.go
│   │   └── postgres/        # PostgreSQL repos
│   ├── interface/           # HTTP handlers
│   │   ├── doc.go
│   │   └── http/
│   │       └── httpx/       # Response helpers
│   ├── observability/       # Logger, tracer
│   │   └── doc.go
│   └── runtimeutil/         # Clock, ID generator
│       └── doc.go
├── db/
│   ├── migrations/          # SQL migrations
│   └── queries/             # sqlc queries
├── README.md
└── go.mod
```

### doc.go Template

Each package needs a doc.go file:

```go
// Package <name> provides <description>.
package <name>
```

### Technology Stack (EXACT versions)

| Tech | Package | Purpose |
|------|---------|---------|
| Go | 1.24.x | Runtime |
| Router | go-chi/chi/v5 | HTTP routing |
| Database | jackc/pgx/v5 | PostgreSQL driver |
| Query | sqlc-dev/sqlc | Type-safe SQL |
| Logger | go.uber.org/zap | Structured logging |
| Config | knadh/koanf/v2 | Configuration |
| Tracing | go.opentelemetry.io/otel | Observability |
| Testing | stretchr/testify | Test assertions |

### Critical Don'ts

❌ Create `common/utils/helpers` packages
❌ Use `zap.L()` global logger
❌ Import infra from usecase
❌ Import interface from infra
❌ Hardcode secrets
❌ Add actual app wiring in this story (wait for later stories)

### Layer Import Rules

```
cmd/ → app only
domain → stdlib, runtimeutil only
usecase → domain only
infra → domain, observability
interface → usecase, domain, httpx
```

### File Naming

- Files: `lower_snake_case.go`
- Packages: `lowercase`, singular
- Struct fields: `PascalCase`
- JSON tags: `snake_case`

---

## Project Structure Notes

This is the **first story** - establishes the foundational structure that all subsequent stories build upon. The structure MUST match:
- Hexagonal/Clean Architecture
- Layer separation as defined in architecture.md
- Package naming conventions from project_context.md

**Note:** Story 1.3+ will add Makefile, docker-compose, etc. Don't add those yet.

---

## References

- [Source: docs/architecture.md#Project-Structure]
- [Source: docs/architecture.md#Technology-Stack]
- [Source: docs/project_context.md#Technology-Stack]
- [Source: docs/project_context.md#Layer-Boundaries]
- [Source: docs/epics.md#Story-1.1]

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Enhanced by validate-create-story quality review.

### Agent Model Used

Dev implementation, Code review with fix.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-10
- Quality validation applied: 2025-12-10
- Code review fixes applied: 2025-12-10
  - Added dependencies to go.mod
  - Created main_test.go for CI validation
  - Fixed duplicate checkbox
  - Standardized doc.go comments

### File List

Files created:
- `go.mod`
- `go.sum`
- `cmd/server/main.go`
- `cmd/server/main_test.go`
- `internal/app/doc.go`
- `internal/config/doc.go`
- `internal/domain/doc.go`
- `internal/usecase/doc.go`
- `internal/infra/doc.go`
- `internal/infra/postgres/doc.go`
- `internal/interface/http/doc.go`
- `internal/interface/http/httpx/doc.go`
- `internal/observability/doc.go`
- `internal/runtimeutil/doc.go`
- `db/migrations/.gitkeep`
- `db/queries/.gitkeep`
- `README.md`

