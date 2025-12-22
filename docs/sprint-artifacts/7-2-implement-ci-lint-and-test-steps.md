Status: Done

## Story

As a **developer**,
I want **lint and test steps in CI with coverage enforcement**,
so that **code quality and correctness are enforced automatically**.

## Acceptance Criteria

1. **Given** CI workflow runs, **When** lint step executes, **Then** `golangci-lint run` is executed with project config, **And** boundary violations (depguard) cause step to fail.

2. **Given** CI workflow runs, **When** test step executes, **Then** `go test -race -coverprofile=coverage.out ./...` runs, **And** test failures cause step to fail, **And** coverage report is uploaded as artifact.

3. **Given** coverage for `internal/domain/...` and `internal/app/...` is below 80%, **When** coverage check runs, **Then** step fails with clear message indicating coverage gap.

*Covers: FR46*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 7.2".
- CI workflow established in Story 7.1: `.github/workflows/ci.yml`.
- Coverage target and commands established in `Makefile` (`make coverage`, 80% threshold).
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Enhance lint step with explicit golangci-lint execution (AC: #1)
  - [x] 1.1 Ensure `make lint` uses `.golangci.yml` configuration
  - [x] 1.2 Verify depguard rules detect boundary violations (already working)
  - [x] 1.3 Add comment in workflow documenting lint behavior

- [x] Task 2: Enhance test step with coverage profile (AC: #2)
  - [x] 2.1 Update test step to use `make test` (already generates coverage.out)
  - [x] 2.2 Add step to upload coverage.out as artifact using `actions/upload-artifact@v4`
  - [x] 2.3 Configure artifact retention (30 days)

- [x] Task 3: Add coverage enforcement step (AC: #3)
  - [x] 3.1 Add new step after tests: "Check coverage threshold"
  - [x] 3.2 Run `make coverage` to enforce 80% threshold on domain+app
  - [x] 3.3 Verify step fails if coverage < 80%

- [x] Task 4: Verify CI workflow integration
  - [x] 4.1 Validate updated YAML syntax
  - [x] 4.2 Test locally: `make lint` and `make coverage` work
  - [x] 4.3 Document final workflow structure

## Dependencies & Blockers

- **Depends on:** Story 7.1 (completed) - Provides the base `.github/workflows/ci.yml`
- **Uses:** Existing Makefile targets: `make lint`, `make test`, `make coverage`
- **Uses:** `.golangci.yml` with depguard rules for boundary enforcement

## Assumptions & Open Questions

- `make coverage` already implements 80% threshold enforcement (verified in Makefile)
- Coverage artifact will be accessible from GitHub Actions UI for debugging
- Existing test suite already passes with >80% coverage (verified: 90.3%)

## Definition of Done

- [x] CI lint step executes `golangci-lint run` via `make lint`
- [x] CI test step generates `coverage.out` profile
- [x] Coverage artifact uploaded for each CI run
- [x] Coverage enforcement step runs `make coverage` after tests
- [x] Coverage < 80% causes CI failure with clear message
- [x] depguard boundary violations cause lint step to fail
- [x] All existing tests continue to pass

## Non-Functional Requirements

- Workflow execution time should remain reasonable (<5 minutes)
- Coverage artifact size should be minimal (text file)
- Clear step names for debugging CI failures

## Testing & Coverage

- Verify lint step: Introduce a boundary violation temporarily → CI fails
- Verify test step: Coverage artifact visible in GitHub Actions
- Verify coverage enforcement: Temporarily lower coverage → CI fails
- Verify normal flow: All steps pass on clean code

## Dev Notes

### ⚠️ CRITICAL: Workflow Enhancement Patterns

Follow these conventions when enhancing GitHub Actions workflows:

```
✅ Use existing Makefile targets for consistency
✅ Upload artifacts for debugging (coverage reports)
✅ Fail fast with clear error messages
✅ Use latest action versions (@v4, @v5)
❌ Don't duplicate logic from Makefile in workflow
❌ Don't use hardcoded coverage thresholds (use Makefile)
```

### Existing Code Context

**From Story 7.1 (Completed):**
| File | Description |
|------|-------------|
| `.github/workflows/ci.yml` | Base CI workflow with lint, test, build steps |
| `Makefile` | Contains `make lint`, `make test`, `make coverage` targets |
| `.golangci.yml` | Linting configuration with depguard boundary rules |

**This story MODIFIES:**
| File | Modification |
|------|-------------|
| `.github/workflows/ci.yml` | Add coverage artifact upload, add coverage check step |

### Current CI Workflow (from 7.1)

```yaml
name: CI

on:
  push:
    branches:
      - '**'
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  ci:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Install dependencies
        run: go mod download

      - name: Run linter
        run: make lint

      - name: Run tests
        run: make test

      - name: Build
        run: make build
```

### Required Changes

Add after "Run tests" step:

```yaml
      - name: Upload coverage report
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.out
          retention-days: 30

      - name: Check coverage threshold
        run: make coverage
```

### Makefile Coverage Target Reference

```makefile
## coverage: Check test coverage meets 80% threshold for domain+app
.PHONY: coverage
coverage:
	@echo "📊 Running tests with coverage (domain+app)..."
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic \
		./internal/domain/... \
		./internal/app/...
	@echo ""
	@echo "📈 Coverage report:"
	@go tool cover -func=coverage.out | tail -1
	@THRESHOLD="$(COVERAGE_THRESHOLD)"; \
	COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{gsub(/%/,"",$$3); print $$3}'); \
	if awk 'BEGIN {exit !('"$$COVERAGE"' < '"$$THRESHOLD"')}'; then \
		echo ""; \
		echo "❌ Coverage $$COVERAGE% is below $$THRESHOLD% threshold"; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ Coverage $$COVERAGE% meets $$THRESHOLD% threshold"; \
	fi
```

### References

- [Source: docs/epics.md#Story 7.2] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#CI/CD Pipeline] - CI pipeline design
- [Source: Makefile#coverage] - Coverage enforcement logic
- [Source: .golangci.yml] - Linting and depguard configuration
- [Source: docs/sprint-artifacts/7-1-setup-github-actions-workflow.md] - Previous story patterns

### Learnings from Story 7.1

**Critical Patterns to Follow:**
1. **Use Makefile targets:** All logic lives in Makefile, workflow just calls targets
2. **Coverage via make coverage:** Already implements 80% threshold enforcement
3. **Sequential steps:** Fail-fast behavior is automatic
4. **Clear step names:** Essential for debugging CI failures

### Coverage Artifact Considerations

- `coverage.out` generated by `make test` (all packages)
- `make coverage` generates coverage for domain+app only (threshold check)
- Upload the `coverage.out` from `make test` for complete picture
- Artifact retention: 30 days (configurable)

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-22)

- `docs/epics.md` - Story 7.2 acceptance criteria
- `docs/architecture.md` - CI/CD pipeline design
- `docs/project-context.md` - Project conventions
- `Makefile` - Coverage enforcement (`make coverage`)
- `.github/workflows/ci.yml` - Base workflow from Story 7.1
- `docs/sprint-artifacts/7-1-setup-github-actions-workflow.md` - Story format reference

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- ✅ Task 1: Added lint behavior comments in ci.yml documenting `.golangci.yml` usage and depguard boundary enforcement
- ✅ Task 2: Added coverage artifact upload step using `actions/upload-artifact@v4` with 30-day retention
- ✅ Task 3: Added coverage threshold check step using `make coverage` (enforces 80% on domain+app)
- ✅ Task 4: Verified YAML syntax, tested `make lint` (0 issues), `make test` (90.3% coverage), `make coverage` (99.0% domain+app)
- All acceptance criteria verified: lint uses .golangci.yml, test generates coverage.out, coverage artifact uploaded, 80% threshold enforced

### Change Log

- 2025-12-22: Story file created (ready-for-dev)
- 2025-12-22: Implementation complete - CI workflow enhanced with coverage artifact and threshold check

### File List

**Modified Files:**
- `.github/workflows/ci.yml` - Added lint comments, coverage artifact upload, coverage threshold check step
