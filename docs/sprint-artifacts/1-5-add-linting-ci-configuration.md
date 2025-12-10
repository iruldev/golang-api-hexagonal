# Story 1.5: Add Linting & CI Configuration

Status: done

## Story

As a tech lead,
I want consistent linting rules and CI example,
So that all code follows the same quality standards.

## Acceptance Criteria

### AC1: Linting with project configuration ✅
**Given** `.golangci.yml` exists in project root
**When** I run `make lint`
**Then** golangci-lint uses the project configuration
**And** no lint errors on clean codebase
**Note:** Already implemented in Story 1.2

### AC2: CI runs tests and linting on push ✅
**Given** `.github/workflows/ci.yml` exists
**When** I push to main branch
**Then** CI runs tests and linting

---

## Tasks / Subtasks

- [x] **Task 1: Verify existing lint setup** (AC: #1)
  - [x] Confirm `.golangci.yml` exists and is correct
  - [x] Run `make lint` to verify no errors (0 issues)
  - [x] No additions needed to linter config

- [x] **Task 2: Create GitHub Actions CI workflow** (AC: #2)
  - [x] Create `.github/workflows/ci.yml`
  - [x] Configure Go 1.24.x setup
  - [x] Add test job with coverage
  - [x] Add lint job with golangci-lint
  - [x] Trigger on push to main and pull requests

- [x] **Task 3: Verify CI configuration** (AC: #2)
  - [x] Run `make test` locally - all tests pass (80% coverage)
  - [x] Run `make lint` locally - 0 issues
  - [x] YAML syntax validated
  - [x] Workflow triggers on push to main ✓
  - [x] Jobs (test, lint) run in parallel ✓

---

## Dev Notes

### GitHub Actions CI Workflow Template

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true  # Enable Go module caching
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          fail_ci_if_error: false
        # Note: For private repos, set CODECOV_TOKEN in repo secrets

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true  # Enable Go module caching
      
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

### Existing .golangci.yml (from Story 1.2)

The linter config is already set up with:
- `version: "2"` (golangci-lint v2 format)
- Linters: errcheck, govet, staticcheck, unused, ineffassign, cyclop
- cyclop max-complexity: 15
- 5m timeout

### Story 1.2 Overlap

**What's already done:**
- `.golangci.yml` created and working
- `make lint` target in Makefile
- `make test` target with -race flag

**What's new in this story:**
- `.github/workflows/ci.yml` for GitHub Actions
- CI trigger on push to main and PRs

### Best Practices for CI

1. **Cache Go modules** for faster builds
2. **Run tests first** → lint second (fail fast on logic errors)
3. **Use matrix** if supporting multiple Go versions (optional for boilerplate)
4. **Upload coverage** to track code quality trends

### References

- [Source: docs/epics.md#Story-1.5]
- [Source: Story 1.2 - .golangci.yml created]
- [GitHub Actions Go](https://github.com/actions/setup-go)
- [golangci-lint-action](https://github.com/golangci/golangci-lint-action)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.

### Agent Model Used

dev-story workflow execution.

### Debug Log References

None.

### Completion Notes List

- Story created: 2025-12-11
- Validation applied: 2025-12-11
- Implementation completed: 2025-12-11
- Code review fixes applied: 2025-12-11
  - Fixed Go version: 1.24 → 1.23 (valid stable version)
  - Removed redundant `go mod download` step
  - Added `timeout-minutes: 10` to both jobs
  - Pinned golangci-lint to v1.61.0 for reproducibility

### File List

Files created:
- `.github/workflows/ci.yml` - GitHub Actions CI workflow (49 lines)

Files modified:
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

Files verified (no changes):
- `.golangci.yml` - Already exists from Story 1.2
- `Makefile` - lint target already exists
