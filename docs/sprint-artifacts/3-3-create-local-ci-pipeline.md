# Story 3.3: Configure Local CI Pipeline

Status: ready-for-dev

## Story

As a **developer**,
I want **to run the full CI pipeline locally**,
So that **I can verify everything passes before pushing**.

## Acceptance Criteria

1. **Given** I want to verify my changes and my working tree is clean (changes staged/committed)
   **When** I run `make ci`
   **Then** the following steps execute in order:
   1. `go mod tidy` check
   2. `gofmt` check
   3. `make lint`
   4. `make test`
   **And** after `go mod tidy`, `git diff --exit-code go.mod go.sum` passes (no changes)
   **And** after `gofmt`, `git diff --exit-code` passes (no formatting changes)
   **And** pipeline fails fast on first error
   **And** exit code is non-zero if any step fails

*Covers: FR49, FR50*

## Tasks / Subtasks

- [ ] Task 1: Create `make ci` target in Makefile (AC: #1)
  - [ ] Add `ci` target that runs steps in sequence
  - [ ] Implement `go mod tidy` check with `git diff --exit-code go.mod go.sum`
  - [ ] Implement `gofmt` check by running `gofmt` and verifying `git diff --exit-code` passes (no formatting changes)
  - [ ] Call existing `make lint` target
  - [ ] Call existing `make test` target
  - [ ] Ensure fail-fast behavior (exit on first error)
  - [ ] Add clear step indicators with emoji for readability

- [ ] Task 2: Create helper targets for individual checks (AC: #1)
  - [ ] Add `check-mod-tidy` target for go.mod verification
  - [ ] Add `check-fmt` target for gofmt verification
  - [ ] Ensure both targets fail with clear error messages

- [ ] Task 3: Update Makefile help documentation (AC: #1)
  - [ ] Add `ci` target with description to help output
  - [ ] Add `check-mod-tidy` and `check-fmt` to help output
  - [ ] Verify `make help` shows new targets

- [ ] Task 4: Verify local CI pipeline works correctly (AC: #1)
  - [ ] Run `make ci` on clean repo - should pass
  - [ ] Test fail path: Modify go.mod without `go mod tidy` - verify failure
  - [ ] Test fail path: Add unformatted code - verify gofmt failure
  - [ ] Test fail path: Add lint violation - verify lint failure
  - [ ] Test fail path: Break a test - verify test failure
  - [ ] Restore clean state after testing

- [ ] Task 5: Update story documentation (N/A)
  - [ ] Add completion notes
  - [ ] Update sprint-status.yaml to `review`

## Dev Notes

### Architecture Context

This story is part of **Epic 3: Local Quality Gates** which ensures developers can verify code quality locally before pushing. The local CI pipeline provides a single command that mimics the actual CI/CD pipeline.

From `docs/epics.md`:
- **FR49**: Developer can run linting checks locally
- **FR50**: Developer can run full CI pipeline locally before pushing

### Current State Analysis

**Existing Makefile Targets (from Story 3.1 & 3.2):**
- `make lint` - golangci-lint with hexagonal boundary rules ✅
- `make test` - Tests with `-race -coverprofile=coverage.out` ✅
- `make coverage` - Coverage check with 80% threshold ✅
- `make build` - Build the application ✅

**Missing:**
- `make ci` - Combined local CI pipeline (this story)
- `check-mod-tidy` - go.mod verification
- `check-fmt` - gofmt verification

### Makefile Implementation

The `ci` target should run steps sequentially with fail-fast behavior:

```makefile
## ci: Run full CI pipeline locally (mod-tidy, fmt, lint, test)
.NOTPARALLEL: ci
.PHONY: ci
ci:
	@$(MAKE) check-mod-tidy
	@$(MAKE) check-fmt
	@$(MAKE) lint
	@$(MAKE) test
	@echo ""
	@echo "✅ All CI checks passed!"

## check-mod-tidy: Verify go.mod and go.sum are tidy
.PHONY: check-mod-tidy
check-mod-tidy:
	@echo "📦 Checking go.mod is tidy..."
	@$(GOMOD) tidy
	@if ! git diff --exit-code go.mod go.sum > /dev/null 2>&1; then \
		echo ""; \
		echo "❌ go.mod or go.sum is not tidy"; \
		echo "   Run 'go mod tidy' and commit the changes"; \
		exit 1; \
	fi
	@echo "✅ go.mod is tidy"

## check-fmt: Verify code is formatted with gofmt
.PHONY: check-fmt
check-fmt:
	@echo "📐 Checking code formatting (gofmt)..."
	@FILES=$$(git ls-files '*.go'); \
	if [ -n "$$FILES" ]; then \
		gofmt -w $$FILES; \
	fi; \
	if ! git diff --exit-code > /dev/null 2>&1; then \
		echo ""; \
		echo "❌ gofmt would change files (working tree is not clean after formatting)"; \
		echo "   Run 'gofmt -w .' and commit the changes"; \
		echo ""; \
		git --no-pager diff --name-only; \
		exit 1; \
	fi
	@echo "✅ All files are formatted"
```

### Key Implementation Details

1. **Fail-Fast + Order Guarantees**: Use a `ci` recipe that calls each step via `$(MAKE)` to guarantee order even if someone runs `make -j ci`. Make stops on the first failing sub-make target.

2. **go mod tidy Check**: 
   - Run `go mod tidy` first
   - Use `git diff --exit-code go.mod go.sum` to check for changes
   - If changes detected, fail with clear message

3. **gofmt Check**:
   - Run `gofmt -w` over git-tracked Go files (`git ls-files '*.go'`)
   - Use `git diff --exit-code` to ensure formatting did not introduce changes
   - Print changed file list on failure

4. **Existing Targets**: Reuse existing `lint` and `test` targets rather than duplicating commands

### Previous Story Learnings (From Story 3.1 & 3.2)

1. **depguard rules active** - All CI checks must pass existing lint rules
2. **Coverage gate exists** - `make coverage` enforces 80% threshold (not included in basic `make ci` to keep it fast; can be added as `make ci-full`)
3. **Race detection enabled** - `make test` already includes `-race` flag
4. **Emoji indicators** - Use consistent emoji for step status (✅ ❌ 📦 📐 🔍)

### Technology Specifics

- **Go version**: 1.24+ (from go.mod toolchain)
- **gofmt**: Part of Go toolchain (stdlib)
- **git diff**: Standard git command, `--exit-code` returns 1 if changes exist
- **Make execution**: Use `.NOTPARALLEL` + sequential `$(MAKE)` calls to guarantee order and fail-fast behavior

### File Locations

| Action | Path |
|--------|------|
| Update | `Makefile` (add `ci`, `check-mod-tidy`, `check-fmt` targets) |
| Verify | `.golangci.yml` (lint config already configured) |
| Reference | `docs/project-context.md` for coding standards |
| Reference | `docs/epics.md#Story-3.3` for acceptance criteria |

### Testing Strategy

1. **Positive test**: Run `make ci` on clean repo → all steps pass
2. **go.mod failure**: Add unused dependency, run `make ci` → should fail at check-mod-tidy
3. **gofmt failure**: Add unformatted code, run `make ci` → should fail at check-fmt
4. **lint failure**: Add boundary violation, run `make ci` → should fail at lint
5. **test failure**: Break a test, run `make ci` → should fail at test
6. **Order verification**: Each failure should happen at the expected step (fail-fast)

### CI Pipeline Alignment

This local `make ci` mirrors what will be in GitHub Actions (Epic 7):

| Local Command | CI/CD Step |
|---------------|------------|
| `make check-mod-tidy` | Verify go.mod is tidy |
| `make check-fmt` | Verify formatting |
| `make lint` | Run golangci-lint |
| `make test` | Run tests with race detector |

### Edge Cases to Handle

1. **No git repository**: `git diff` will fail if not in a git repo. For now, assume git is always available (per project setup).

2. **Working tree cleanliness**: `git diff --exit-code` checks the working tree. This workflow is intended to run on a clean working tree (typically after staging/committing) so the “no changes” checks are meaningful.

3. **go.mod already tidy**: `go mod tidy` is idempotent. If already tidy, `git diff` will show no changes and pass.

4. **No Go files**: `git ls-files '*.go'` may return empty (e.g., docs-only change). In that case, formatting step becomes a no-op and should succeed.

### References

- [Source: docs/epics.md#Story-3.3]
- [Source: docs/project-context.md#Makefile-Commands]
- [Source: docs/architecture.md#CI-CD-Pipeline]
- [Source: docs/sprint-artifacts/3-1-configure-golangci-lint-with-boundary-rules.md#Dev-Notes]
- [Source: docs/sprint-artifacts/3-2-setup-unit-test-infrastructure.md#Dev-Notes]
- [Source: docs/sprint-artifacts/epic-2-retro-2025-12-17.md#Key-Learnings]

## Dev Agent Record

### Context Reference

- `docs/project-context.md` - Coding standards and Makefile patterns
- `docs/architecture.md` - CI/CD pipeline structure
- `docs/epics.md` - Story definition and acceptance criteria
- `docs/sprint-artifacts/3-1-configure-golangci-lint-with-boundary-rules.md` - Lint configuration
- `docs/sprint-artifacts/3-2-setup-unit-test-infrastructure.md` - Test and coverage setup
- `Makefile` - Current targets and patterns

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List

- `Makefile` (update - add `ci`, `check-mod-tidy`, `check-fmt` targets)
- `docs/sprint-artifacts/sprint-status.yaml` (update - mark story status)
