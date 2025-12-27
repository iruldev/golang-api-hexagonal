# Story 2.4: Race Detection Policy (nightly/selective)

Status: done

## Story

As a **CI system**,
I want race detection without slowing PRs,
so that races are caught without blocking velocity.

## Acceptance Criteria

1. **AC1:** `make test-race-selective` runs race on packages in `scripts/race_packages.txt`
2. **AC2:** Nightly workflow runs full race detection
3. **AC3:** `scripts/race_packages.txt` lists high-risk packages
4. **AC4:** Race failures notify team via workflow alert

## Tasks / Subtasks

- [x] Task 1: Create race_packages.txt (AC: #3)
  - [x] Create `scripts/race_packages.txt`
  - [x] List high-risk packages (infra/postgres, app/*, transport/*)
  - [x] Document format in comments
- [x] Task 2: Add test-race-selective target (AC: #1)
  - [x] Add `make test-race-selective` to Makefile
  - [x] Read packages from `scripts/race_packages.txt`
  - [x] Run `go test -race` on listed packages
- [x] Task 3: Create nightly workflow (AC: #2)
  - [x] Create `.github/workflows/nightly.yml`
  - [x] Run full `go test -race ./...` nightly
  - [x] Schedule for off-hours (e.g., 3 AM UTC)
- [x] Task 4: Configure failure alerts (AC: #4)
  - [x] Add failure notification in nightly workflow
  - [x] Can use GitHub notification or Slack webhook

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-006:** Race detection strategy
- **NFR-8:** Performance vs thoroughness tradeoff

### race_packages.txt Format

```txt
# High-risk packages for race detection
# One package path per line, relative to module root
# Lines starting with # are comments

internal/infra/postgres
internal/app/user
internal/app/audit
internal/transport/http/handler
internal/transport/http/middleware
```

### Makefile Target

```makefile
## test-race-selective: Run race detection on high-risk packages
.PHONY: test-race-selective
test-race-selective:
	@echo "🏎️ Running race detection on high-risk packages..."
	@cat scripts/race_packages.txt | grep -v '^#' | grep -v '^$$' | \
		xargs -I {} go test -race -v ./{}
	@echo "✅ Race detection complete"
```

### Nightly Workflow

```yaml
name: Nightly

on:
  schedule:
    - cron: '0 3 * * *'  # 3 AM UTC daily
  workflow_dispatch:  # Manual trigger

jobs:
  race-detection:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Full race detection
        run: go test -race -timeout 30m ./...
      - name: Notify on failure
        if: failure()
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: 1,  # Or use dedicated channel
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '❌ Nightly race detection failed!'
            })
```

### Testing Standards

- Verify `make test-race-selective` runs on listed packages only
- Verify nightly workflow triggers on schedule
- Test failure notification works

### Previous Story Learnings (Story 2.3)

- CI steps should have clear comments
- Reference story number for traceability
- GitHub Actions handles failures by default (blocks PR)

### References

- [Source: _bmad-output/architecture.md#AD-006 Race Detection]
- [Source: _bmad-output/epics.md#Story 2.4]
- [Source: _bmad-output/prd.md#FR4, FR13, NFR8]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### Filed List

_Files created/modified during implementation:_
- [x] `scripts/race_packages.txt` (new)
- [x] `Makefile` (add test-race-selective target)
- [x] `.github/workflows/nightly.yml` (new)
- [x] `internal/infra/postgres/user_repo_test.go` (modified for CI support)
- [x] `internal/infra/postgres/test_helpers_test.go` (new helper for tests)

### Review Fixes (AI)
- Fixed integration tests to use `testcontainers` fallback for CI
- Added untracked files to git
