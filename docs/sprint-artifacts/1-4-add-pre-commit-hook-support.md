# Story 1.4: Add Pre-commit Hook Support

Status: Done

## Story

As a developer,
I want `make hooks` to install pre-commit hooks,
So that I catch lint issues before pushing.

## Acceptance Criteria

1. **Given** I have run `make hooks`
   **When** I commit code with lint violations
   **Then** the pre-commit hook runs and blocks the commit
   **And** I see which violations need fixing

2. **Given** a fresh clone of the repository
   **When** I run `make hooks`
   **Then** the git pre-commit hook is installed in `.git/hooks/`
   **And** the installation succeeds with a confirmation message

3. **Given** the pre-commit hook is installed
   **When** I attempt to commit code that passes lint
   **Then** the commit succeeds without blocking

4. **Given** the pre-commit hook runs on commit
   **When** lint violations are detected
   **Then** the commit is blocked with exit code 1
   **And** the output shows the specific violations with file:line references

## Tasks / Subtasks

- [x] Task 1: Create pre-commit hook script (AC: #1, #4)
  - [x] 1.1 Create `.githooks/pre-commit` script in project root
  - [x] 1.2 Script should run `golangci-lint run --config policy/golangci.yml` on staged files
  - [x] 1.3 Script should exit 1 if violations found, blocking commit
  - [x] 1.4 Script should output clear lint violation messages

- [x] Task 2: Implement `make hooks` command (AC: #2)
  - [x] 2.1 Add `hooks` target to Makefile
  - [x] 2.2 Configure git to use `.githooks` directory OR symlink hook to `.git/hooks/`
  - [x] 2.3 Output confirmation message on successful installation
  - [x] 2.4 Add `hooks` to `.PHONY` targets

- [x] Task 3: Update documentation (AC: #2)
  - [x] 3.1 Add `hooks` command to `make help` under Development Workflow
  - [x] 3.2 Verify `project_context.md` already documents `make hooks`

- [x] Task 4: Verify implementation (AC: #1, #3, #4)
  - [x] 4.1 Run `make hooks` and verify hook is installed
  - [x] 4.2 Create file with intentional lint violation, attempt commit → blocked
  - [x] 4.3 Fix violation, attempt commit → succeeds
  - [x] 4.4 Verify hook runs fast (within seconds, not minutes)

## Dev Notes

### Pre-commit Hook Strategy

**Two options for git hooks:**

1. **Git core.hooksPath (RECOMMENDED):**
   ```bash
   git config core.hooksPath .githooks
   ```
   - Keeps hooks in repository
   - No symlink needed
   - Works for all developers who run `make hooks`

2. **Symlink approach:**
   ```bash
   ln -sf ../../.githooks/pre-commit .git/hooks/pre-commit
   ```
   - Each dev must run it
   - Hook not tracked in git (`.git/hooks/` is not versioned)

**Recommendation:** Use `git config core.hooksPath .githooks` as it's cleaner and the hooks are versioned.

### Pre-commit Script Implementation

```bash
#!/bin/bash
# .githooks/pre-commit
# Pre-commit hook to run lint checks

set -e

echo "🔍 Running pre-commit lint check..."

# Run golangci-lint with policy config
if ! golangci-lint run --config policy/golangci.yml ./...; then
    echo ""
    echo "❌ Lint check failed. Please fix the violations above before committing."
    echo "   Run 'make lint' to see all issues."
    exit 1
fi

echo "✓ Lint check passed!"
exit 0
```

### Makefile Target Pattern

```makefile
.PHONY: hooks

# Install pre-commit hooks
hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit
	@echo "✓ Pre-commit hooks installed. Git will use .githooks/ directory."
```

### Previous Story Learnings (from Story 1.3)

- Makefile pattern follows consistent style with `@echo "✓ ..."` for success messages
- `reset` target works correctly with `docker compose down -v`
- Lint uses `--config policy/golangci.yml` as established in Story 1.1
- Help target uses organized categories (Development Workflow, Testing, etc.)
- New targets should be added to `.PHONY` declaration

### File Structure

```
project-root/
├── .githooks/
│   └── pre-commit        # [NEW] Pre-commit hook script
├── Makefile              # [MODIFY] Add `hooks` target
└── docs/sprint-artifacts/
    └── sprint-status.yaml  # [MODIFY] Update status to ready-for-dev
```

### NFR Compliance

| NFR | Requirement | How Addressed |
|-----|-------------|---------------|
| NFR-DX4 | Pre-commit hooks available | `make hooks` installs pre-commit hook |
| NFR-P5 | make lint ≤60sec | Pre-commit uses same lint command |

### Edge Cases to Handle

1. **No golangci-lint installed:** Script should fail gracefully with helpful message
2. **Hook already installed:** `make hooks` should be idempotent (run multiple times safely)
3. **No staged Go files:** Hook should skip lint check and succeed
4. **Git not initialized:** `make hooks` should fail gracefully

### Testing Strategy

1. **Installation Test:**
   - Run `make hooks`
   - Verify `git config core.hooksPath` returns `.githooks`
   - Verify `.githooks/pre-commit` is executable

2. **Blocking Test:**
   - Create file with intentional error (e.g., unused variable)
   - Stage file with `git add`
   - Run `git commit` → should fail
   - Verify exit code is 1

3. **Success Test:**
   - Fix the lint violation
   - Run `git commit` → should succeed

### Architecture Compliance

- No layer violations - this is infrastructure/tooling only
- Uses existing `policy/golangci.yml` for lint configuration
- Follows established Makefile patterns from Stories 1.1-1.3

### References

- [Source: docs/epics.md#Story 1.4](file:///docs/epics.md) - Story requirements (FR9)
- [Source: docs/prd.md#Developer Experience](file:///docs/prd.md) - FR9: `make hooks` installs pre-commit hooks
- [Source: project_context.md#Key Commands](file:///project_context.md) - Already lists `make hooks`
- [Source: docs/sprint-artifacts/1-3-implement-makefile-developer-commands.md](file:///docs/sprint-artifacts/1-3-implement-makefile-developer-commands.md) - Previous story patterns

## Dev Agent Record

### Context Reference

- Story generated by create-story workflow on 2025-12-15
- Epic 1: Foundation & Quality Gates (MVP) - in-progress
- Previous stories: 1.1 (done), 1.2 (done), 1.3 (done)

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

None required.

### Completion Notes List

- ✅ Created `.githooks/pre-commit` script with full lint check functionality
- ✅ Script handles edge cases: missing golangci-lint, no staged Go files, and git not initialized
- ✅ Added `make hooks` target using `git config core.hooksPath .githooks` approach
- ✅ Updated `make help` with hooks command in Development Workflow section
- ✅ Verified `git config core.hooksPath` returns `.githooks` after running `make hooks`
- ✅ Pre-commit script runs fast (seconds) with lint via `golangci-lint --config policy/golangci.yml`
- ✅ All lint checks pass (0 issues)
- ✅ Tests pass (proto coverage is pre-existing issue, unrelated to this story)

### File List

- [NEW] `.githooks/pre-commit` - Pre-commit hook script
- [MODIFY] `Makefile` - Added `hooks` target and `.PHONY` entry
- [MODIFY] `docs/sprint-artifacts/sprint-status.yaml` - Updated story status

## Senior Developer Review (AI)

- **Date:** 2025-12-15
- **Reviewer:** Code Review Workflow
- **Outcome:** Approved with Improvements

### Findings
- **Optimization:** Pre-commit hook was checking all files. Updated to use `--new-from-rev=HEAD` to only lint changed code. This satisfies the "fast execution" requirement more robustly.
- **Verification:**
  - `make hooks` installs correctly.
  - Pre-commit blocking logic verified (exit code 1 on violations).
  - Documentation in `Makefile` and `help` is correct.

### Status
- Moved to **Done** after applying performance fix.
