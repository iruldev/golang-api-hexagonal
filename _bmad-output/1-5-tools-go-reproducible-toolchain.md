# Story 1.5: tools.go for Reproducible Toolchain

Status: done

## Story

As a **developer**,
I want pinned tool versions,
so that everyone uses same versions.

## Acceptance Criteria

1. **AC1:** `tools/tools.go` with blank imports for: mockgen, sqlc, goose, golangci-lint
2. **AC2:** `make bootstrap` runs `go install` for all tools
3. **AC3:** Tool versions pinned via go.mod
4. **AC4:** README documents `make bootstrap` as first step

## Tasks / Subtasks

- [x] Task 1: Update tools/tools.go (AC: #1, #3)
  - [x] Add blank import for sqlc
  - [x] Add blank import for goose
  - [x] Add blank import for golangci-lint
  - [x] Run `go mod tidy` to pin versions
- [x] Task 2: Add make bootstrap target (AC: #2)
  - [x] Add `bootstrap` target to Makefile
  - [x] Install all tools with pinned versions from go.mod
  - [x] Print helpful success message
- [x] Task 3: Update README (AC: #4)
  - [x] Add `make bootstrap` as first step in Getting Started
  - [x] Document what tools are installed

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-006:** Reproducible Toolchain requirements
- **Pattern 7:** Tool version pinning

### Current tools/tools.go

```go
// Current state (from Story 1.2)
import (
    _ "go.uber.org/mock/gomock"
)
```

### Updated tools/tools.go

```go
//go:build tools
// +build tools

// Package tools pins tool dependencies for reproducible builds.
// Install all tools via: make bootstrap
package tools

import (
    _ "go.uber.org/mock/mockgen"
    _ "github.com/sqlc-dev/sqlc/cmd/sqlc"
    _ "github.com/pressly/goose/v3/cmd/goose"
    _ "github.com/golangci-lint/golangci-lint/cmd/golangci-lint"
)
```

### make bootstrap Target

```makefile
## bootstrap: Install all development tools (run once after clone)
.PHONY: bootstrap
bootstrap:
	@echo "🔧 Installing development tools..."
	go install go.uber.org/mock/mockgen@$(shell go list -m -f '{{.Version}}' go.uber.org/mock)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(shell go list -m -f '{{.Version}}' github.com/sqlc-dev/sqlc)
	go install github.com/pressly/goose/v3/cmd/goose@$(shell go list -m -f '{{.Version}}' github.com/pressly/goose/v3)
	go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@$(shell go list -m -f '{{.Version}}' github.com/golangci-lint/golangci-lint)
	@echo "✅ All tools installed"
```

### Pinned Tool Versions (from go.mod)

| Tool | Package | Current Version |
|------|---------|-----------------|
| mockgen | go.uber.org/mock | v0.6.0 |
| sqlc | github.com/sqlc-dev/sqlc | v1.28.0 |
| goose | github.com/pressly/goose/v3 | v3.26.0 |
| golangci-lint | github.com/golangci-lint/golangci-lint | v1.64.2 |

### Previous Story Learnings (Story 1.1-1.4)

- tools/tools.go already exists (Story 1.2)
- setup target exists but installs with @latest (not pinned)
- Use consistent emoji pattern in Makefile

### References

- [Source: _bmad-output/architecture.md#AD-006 Reproducible Toolchain]
- [Source: _bmad-output/epics.md#Story 1.5]
- [Source: _bmad-output/prd.md#FR1]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### File List

_Files created/modified during implementation:_
- [ ] `tools/tools.go` (add imports)
- [ ] `Makefile` (add bootstrap target)
- [ ] `README.md` (document make bootstrap)
