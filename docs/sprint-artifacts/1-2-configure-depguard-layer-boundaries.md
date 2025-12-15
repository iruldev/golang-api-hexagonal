# Story 1.2: Configure depguard Layer Boundaries

Status: done

## Story

As a developer,
I want depguard rules enforcing hexagonal layer boundaries,
So that I cannot accidentally import infra from usecase.

## Acceptance Criteria

1. **Given** code in `internal/usecase/` that imports `internal/infra/`
   **When** I run `make lint`
   **Then** depguard reports a boundary violation error
   **And** the error message explains which layer rule was violated

2. **Given** code in `internal/domain/` that imports any other internal layer
   **When** I run `make lint`
   **Then** depguard reports a boundary violation error

3. **Given** valid code respecting layer boundaries
   **When** I run `make lint`
   **Then** no depguard violations are reported

## Tasks / Subtasks

- [x] Task 1: Enable depguard in golangci-lint (AC: #1, #2, #3)
  - [x] 1.1 Add `depguard` to enabled linters in `policy/golangci.yml`
  - [x] 1.2 Configure depguard settings in `policy/golangci.yml` (inline rules)
  - [x] 1.3 Verify depguard is recognized and runs with `golangci-lint linters`

- [x] Task 2: Configure layer boundary rules (AC: #1, #2)
  - [x] 2.1 Add domain layer rules (cannot import usecase, interface, infra)
  - [x] 2.2 Add usecase layer rules (cannot import interface, infra)
  - [x] 2.3 Add interface layer rules (cannot import infra)
  - [x] 2.4 Infra layer rules (no restrictions needed per architecture)
  - [x] 2.5 Configure descriptive error messages for each violation type
  - [x] 2.6 (BONUS) Fixed existing violation - created ctxutil package

- [x] Task 3: Verify implementation (AC: #1, #2, #3)
  - [x] 3.1 Run `make lint` and verify no violations in existing code
  - [x] 3.2 Temporarily add invalid import in usecase to verify depguard catches it
  - [x] 3.3 Revert test import and confirm clean lint output
  - [x] 3.4 Verify error messages are descriptive and actionable

- [x] Task 4: Update documentation
  - [x] 4.1 Update `policy/README.md` with depguard configuration details
  - [x] 4.2 Review `AGENTS.md` - layer boundaries already documented

- [x] Review Follow-ups (AI)
  - [x] [AI-Review][High] Enabled linting for test files (AC: #1)
  - [x] [AI-Review][Medium] Added unit tests for ctxutil package
  - [x] [AI-Review][Medium] Fixed depguard subpackage matching pattern
  - [x] [AI-Review] Fixed lint issues in existing test files

## Dev Notes

### Architectural Requirements (MUST FOLLOW)

**Layer Dependencies from `docs/architecture.md` and `project_context.md`:**

```
domain → (nothing)
usecase → domain only
interface → usecase, domain
infra → domain only
```

These rules are documented in:
- [Source: docs/architecture.md#Layer Dependencies](file:///docs/architecture.md) - Table showing allowed dependencies
- [Source: project_context.md#Layer Boundaries](file:///project_context.md) - Critical rules summary
- [Source: docs/architecture-decisions.md#Architectural Boundaries](file:///docs/architecture-decisions.md) - Boundary enforcement table

### depguard Configuration Pattern

**From technical research (`docs/analysis/research/technical-go-golden-template-2025-12-15.md`):**

```yaml
# policy/golangci.yml - depguard settings
linters-settings:
  depguard:
    rules:
      domain:
        files:
          - "**/internal/domain/**"
        deny:
          - pkg: "github.com/*/internal/usecase"
            desc: "domain cannot import usecase"
          - pkg: "github.com/*/internal/interface"
            desc: "domain cannot import interface"
          - pkg: "github.com/*/internal/infra"
            desc: "domain cannot import infra"
            
      usecase:
        files:
          - "**/internal/usecase/**"
        deny:
          - pkg: "github.com/*/internal/interface"
            desc: "usecase cannot import interface"
          - pkg: "github.com/*/internal/infra"
            desc: "usecase cannot import infra"
```

### golangci-lint v2 Configuration

**Current `policy/golangci.yml` structure (from Story 1.1):**
- version: "2"
- Uses `linters.settings` section for linter-specific config
- Enabled linters: errcheck, govet, staticcheck, unused, ineffassign, cyclop
- **depguard needs to be added to the `enable` list**

**Recommended approach:** Add depguard inline in `policy/golangci.yml` rather than external file reference, as golangci-lint v2 expects all linter settings in main config.

### Implementation Pattern (RECOMMENDED)

**policy/golangci.yml addition:**

```yaml
linters:
  enable:
    # ... existing linters ...
    - depguard

  settings:
    # ... existing settings ...
    depguard:
      rules:
        domain-layer:
          files:
            - "**/internal/domain/**"
          deny:
            - pkg: "github.com/iruldev/golang-api-hexagonal/internal/usecase"
              desc: "domain layer cannot import usecase - violates hexagonal architecture"
            - pkg: "github.com/iruldev/golang-api-hexagonal/internal/interface"
              desc: "domain layer cannot import interface - violates hexagonal architecture"
            - pkg: "github.com/iruldev/golang-api-hexagonal/internal/infra"
              desc: "domain layer cannot import infra - violates hexagonal architecture"
        
        usecase-layer:
          files:
            - "**/internal/usecase/**"
          deny:
            - pkg: "github.com/iruldev/golang-api-hexagonal/internal/interface"
              desc: "usecase layer cannot import interface - violates hexagonal architecture"
            - pkg: "github.com/iruldev/golang-api-hexagonal/internal/infra"
              desc: "usecase layer cannot import infra - use dependency injection"
        
        interface-layer:
          files:
            - "**/internal/interface/**"
          deny:
            - pkg: "github.com/iruldev/golang-api-hexagonal/internal/infra"
              desc: "interface layer cannot import infra - use dependency injection"
```

### Important Considerations

1. **Package Path:** Use full module path `github.com/iruldev/golang-api-hexagonal/...` for deny rules (not wildcards)
2. **Files Pattern:** Use `**/internal/{layer}/**` to match all files in layer
3. **Observability Exception:** The `internal/observability/` package is cross-cutting and may be used by any layer - do NOT restrict it
4. **Testing Exception:** Test files may need to import test utilities - consider excluding `*_test.go` if issues arise
5. **Config/App Exception:** `internal/config/` and `internal/app/` are infrastructure-level but used for wiring - may need special handling

### Known Constraints

1. **Module Path:** `github.com/iruldev/golang-api-hexagonal` (verify in `go.mod`)
2. **Existing Code:** Run full lint to check for any pre-existing violations before adding rules
3. **Cross-cutting Concerns:** `internal/observability/`, `internal/testing/`, `internal/runtimeutil/` may need exceptions

### Previous Story Learnings (from Story 1.1)

- Policy directory structure is already created at `policy/`
- `policy/depguard.yml` exists as placeholder with example structure
- `make lint` uses `--config policy/golangci.yml`
- golangci-lint v2 format is already in use
- Makefile lint target works correctly

### Testing Strategy

1. **Positive Test:** Existing code should pass lint (no violations expected)
2. **Negative Test:** Temporarily add invalid import to verify detection:
   ```go
   // Temporarily add in any usecase file:
   import _ "github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
   ```
3. **Error Message Test:** Verify error message includes layer name and helpful description

### Files to Modify

| Action | File | Notes |
|--------|------|-------|
| MODIFY | `policy/golangci.yml` | Add depguard to linters, add depguard settings |
| MODIFY | `policy/depguard.yml` | Update with actual configuration (or note if inline) |
| MODIFY | `policy/README.md` | Document depguard rules and layer boundaries |
| MODIFY | `AGENTS.md` | Add layer boundary enforcement section if not present |

### Project Structure Notes

- Current internal/ structure has 10 subdirectories: app, config, domain, infra, interface, observability, runtimeutil, testing, usecase, worker
- Layer boundaries apply to: domain, usecase, interface, infra
- Cross-cutting packages (observability, runtimeutil, testing, config, app, worker) have special handling

### References

- [Source: docs/architecture-decisions.md#Decision 2](file:///docs/architecture-decisions.md) - Policy Pack Directory decision
- [Source: docs/architecture-decisions.md#Architectural Boundaries](file:///docs/architecture-decisions.md) - Layer boundary table
- [Source: docs/architecture.md#Layer Dependencies](file:///docs/architecture.md) - Layer dependency rules
- [Source: project_context.md#Layer Boundaries](file:///project_context.md) - Critical layer rules
- [Source: docs/analysis/research/technical-go-golden-template-2025-12-15.md#depguard](file:///docs/analysis/research/technical-go-golden-template-2025-12-15.md) - depguard configuration examples
- [Source: docs/sprint-artifacts/1-1-setup-policy-pack-directory.md](file:///docs/sprint-artifacts/1-1-setup-policy-pack-directory.md) - Previous story implementation and learnings

## Dev Agent Record

### Context Reference

- Story generated by create-story workflow on 2025-12-15
- Epic 1: Foundation & Quality Gates (MVP) - in-progress

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

None required.

### Completion Notes List

- ✅ Added depguard to `policy/golangci.yml` with 3 layer boundary rules (domain, usecase, interface)
- ✅ depguard correctly detected existing violation in `usecase/note/usecase.go` importing middleware
- ✅ Created `internal/ctxutil/ctxutil.go` - new cross-cutting package for Claims and RequestID context utilities
- ✅ Refactored `middleware/auth.go` and `middleware/requestid.go` to use ctxutil (backwards-compatible wrappers)
- ✅ Fixed `usecase/note/usecase.go` to use ctxutil instead of middleware import
- ✅ Updated `middleware/auth_test.go` to use public API instead of internal claimsKey
- ✅ All tests pass (31+ package tests)
- ✅ `make lint` returns 0 issues
- ✅ Tested violation detection with temporary invalid import - works correctly
- ✅ Updated `policy/README.md` with depguard documentation
- ✅ [AI-Review] Enabled test linting and fixed 17+ issues in test files
- ✅ [AI-Review] Added `ctxutil` unit tests (100% coverage)
- ✅ [AI-Review] Removed unused `policy/depguard.yml`

### File List

| Action | File |
|--------|------|
| MODIFY | `policy/golangci.yml` |
| DELETE | `policy/depguard.yml` |
| CREATE | `internal/ctxutil/ctxutil.go` |
| CREATE | `internal/ctxutil/ctxutil_test.go` |
| MODIFY | `internal/interface/http/middleware/requestid.go` |
| MODIFY | `internal/interface/http/middleware/auth.go` |
| MODIFY | `internal/interface/http/middleware/auth_test.go` |
| MODIFY | `internal/interface/http/middleware/apikey_example_test.go` |
| MODIFY | `internal/observability/audit_logger_test.go` |
| MODIFY | `internal/interface/grpc/interceptor/interceptor_test.go` |
| MODIFY | `cmd/bplat/cmd/generate_test.go` |
| MODIFY | `internal/usecase/note/usecase.go` |
| MODIFY | `policy/README.md` |
