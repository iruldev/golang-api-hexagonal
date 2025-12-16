# Story 1.1: Initialize Hexagonal Folder Structure

Status: done

## Story

**As a** developer,
**I want** a pre-configured hexagonal folder structure with clear layer separation,
**So that** I can immediately understand where to place code.

## Acceptance Criteria

1. **Given** I clone the repository
   **When** I view the project structure
   **Then** the following directories exist and are preserved in git via `.keep`:
   - `cmd/api/`
   - `internal/domain/`
   - `internal/app/`
   - `internal/transport/http/handler/`
   - `internal/transport/http/contract/`
   - `internal/transport/http/middleware/`
   - `internal/infra/config/`
   - `internal/infra/postgres/`
   - `internal/infra/observability/`
   - `internal/shared/`
   - `migrations/`
   - `.github/workflows/`
   - `docs/`

2. **Given** all directories exist
   **When** I view `cmd/api/main.go`
   **Then** entry point file exists with minimal boilerplate

3. **Given** all source files present
   **When** I run `go build ./...`
   **Then** build succeeds without errors

## Tasks / Subtasks

- [x] Task 1: Create directory structure (AC: #1)
  - [x] Create `cmd/api/` with `main.go` placeholder
  - [x] Create `internal/domain/` with `.keep`
  - [x] Create `internal/app/` with `.keep`
  - [x] Create transport subdirectories with `.keep`
  - [x] Create infra subdirectories with `.keep`
  - [x] Create `internal/shared/` with `.keep`
  - [x] Create `migrations/` with `.keep`
  - [x] Create `.github/workflows/` with `.keep`
  - [x] Create `docs/` with `.keep`

- [x] Task 2: Create minimal main.go (AC: #2)
  - [x] Implement basic `main()` function with placeholder comment
  - [x] Add package declaration and basic imports
  - [x] Ensure file compiles standalone

- [x] Task 3: Initialize Go module (AC: #3)
  - [x] Run `go mod init github.com/iruldev/golang-api-hexagonal`
  - [x] Verify `go build ./...` succeeds

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

This story establishes the hexagonal architecture foundation. Key layer rules:

| Layer | Location | Allowed Imports | Forbidden |
|-------|----------|-----------------|-----------|
| Domain | `internal/domain/` | stdlib only | slog, uuid, pgx, chi, otel |
| App | `internal/app/` | domain only | net/http, pgx, slog, otel |
| Transport | `internal/transport/http/` | domain, app, chi, uuid | pgx, direct infra |
| Infra | `internal/infra/` | domain, pgx, slog, otel | app, transport |

### Directory Purpose Reference [Source: docs/architecture.md]

```
cmd/api/                          # Application entry point
├── main.go
internal/
├── domain/                       # Business entities, interfaces (stdlib only)
├── app/                          # Use cases, application logic
│   └── user/                     # Example: user use cases
├── transport/
│   └── http/
│       ├── handler/              # HTTP handlers
│       ├── contract/             # DTOs, request/response types
│       └── middleware/           # HTTP middleware
├── infra/
│   ├── config/                   # Configuration loading
│   ├── postgres/                 # Database repositories
│   └── observability/            # Logging, tracing, metrics
└── shared/                       # Cross-cutting utilities
migrations/                       # Database migrations (goose)
.github/workflows/                # CI/CD pipelines
docs/                            # Documentation
```

### .keep File Convention

Empty directories are preserved using `.keep` files (or `.gitkeep`). These are empty files that allow git to track otherwise empty directories.

```bash
# Create .keep file
touch internal/domain/.keep
```

### Minimal main.go Template

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	// TODO: Wire up dependencies and start server
	fmt.Println("golang-api-hexagonal starting...")
	os.Exit(0)
}
```

### Go Module Name

**Module path:** `github.com/iruldev/golang-api-hexagonal`

### Testing Verification

After implementation, verify:
```bash
go build ./...          # Should succeed
go mod tidy             # Should not add dependencies
```

## Technical Requirements

- **Go version:** 1.23+ [Source: docs/project-context.md]
- **Module path:** github.com/iruldev/golang-api-hexagonal
- **No external dependencies** for this story (stdlib only in main.go)

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- Domain layer: stdlib ONLY (no slog, uuid, pgx, chi, otel)
- Use `type ID string` for identifiers (not uuid.UUID in domain)
- NO logging in domain layer - ever
- Authorization checks happen in app layer
- depguard enforcement in CI — violations = build failure

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-16)

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- `go mod init github.com/iruldev/golang-api-hexagonal` - SUCCESS
- `go build ./...` - SUCCESS (no errors)
- `go mod tidy` - SUCCESS (no dependencies added)

### Completion Notes List

- [x] All directories created with .keep files
- [x] main.go compiles successfully
- [x] go build ./... passes
- [x] No external dependencies added

### File List

Files created:
- `cmd/api/main.go` (NEW)
- `cmd/api/.keep` (NEW)
- `.gitignore` (NEW)
- `internal/domain/.keep` (NEW)
- `internal/app/.keep` (NEW)
- `internal/transport/http/handler/.keep` (NEW)
- `internal/transport/http/contract/.keep` (NEW)
- `internal/transport/http/middleware/.keep` (NEW)
- `internal/infra/config/.keep` (NEW)
- `internal/infra/postgres/.keep` (NEW)
- `internal/infra/observability/.keep` (NEW)
- `internal/shared/.keep` (NEW)
- `migrations/.keep` (NEW)
- `.github/workflows/.keep` (NEW)
- `docs/.keep` (NEW)
- `docs/analysis/product-brief-golang-api-hexagonal-2025-12-16.md` (NEW)
- `docs/architecture.md` (NEW)
- `docs/bmm-workflow-status.yaml` (NEW)
- `docs/epics.md` (NEW)
- `docs/implementation-readiness-report-2025-12-16.md` (NEW)
- `docs/prd.md` (NEW)
- `docs/project-context.md` (NEW)
- `docs/sprint-artifacts/sprint-status.yaml` (NEW)
- `docs/test-design-system.md` (NEW)
- `docs/sprint-artifacts/1-1-initialize-hexagonal-folder-structure.md` (NEW)
- `go.mod` (NEW)

### Senior Developer Review (AI)

Review menemukan masalah utama di *dokumen story* dan sedikit gap pada AC:
- Dev Agent Record → File List tidak merefleksikan perubahan aktual yang terlihat di git (banyak file staged tapi tidak didokumentasikan).
- AC #1 mengimplikasikan semua direktori “dipreservasi via `.keep`”, namun `cmd/api/` sebelumnya tidak punya `.keep` (meski tidak kosong karena ada `main.go`).
- Ada artefak lokal `./api` (binary) yang sebaiknya tidak ada di workspace (meski sudah di-ignore).

Perbaikan yang diaplikasikan (auto-fix):
- Tambah `cmd/api/.keep` untuk menghilangkan ambiguitas AC #1.
- Update File List agar sesuai dengan perubahan yang ada di git.
- Verifikasi `go build ./...` sukses.
- Bersihkan artefak lokal `./api`.

### Change Log

- 2025-12-16: Story 1.1 implemented - Hexagonal folder structure created with all directories, .keep files, main.go entry point, and go.mod initialized
- 2025-12-16: Senior Dev Review (AI) - Reconciled story record with git reality, added cmd/api/.keep, verified build, cleaned local artifacts
