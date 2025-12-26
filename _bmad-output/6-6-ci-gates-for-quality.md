# Story 6.6: CI Gates for Quality

Status: done

## Story

**As a** developer,
**I want** CI to run lint, test, vuln, secret scans,
**So that** quality is enforced automatically.

**FR:** FR30

## Acceptance Criteria

1. ✅ **Given** PR with code changes, **When** CI runs, **Then** golangci-lint runs with required linters
2. ✅ **Given** PR with code changes, **When** CI runs, **Then** go test runs with coverage
3. ✅ **Given** PR with code changes, **When** CI runs, **Then** govulncheck scans for vulnerabilities
4. ✅ **Given** PR with code changes, **When** CI runs, **Then** secret scanner checks for leaked secrets
5. ✅ **Given** any gate fails, **When** CI runs, **Then** CI pipeline fails

## Implementation Summary

### Task 1: Add govulncheck ✅
- **Already implemented** in previous work (lines 107-111 of ci.yml)
- **Pinned** to v1.1.4 for reproducibility

### Task 2: Add gitleaks ✅
- Added `gitleaks/gitleaks-action@v2` as first step after checkout
- Uses `fetch-depth: 0` for full git history (required for gitleaks)
- Placed early per fail-fast strategy

### Task 3: Verify existing gates ✅
- `golangci-lint` active (line 48-51)
- `go test` with coverage active (line 83-97)
- Pipeline fails on any step failure (default GitHub Actions behavior)

## Changes

| File | Change |
|------|--------|
| `.github/workflows/ci.yml` | MODIFIED - Added gitleaks, pinned govulncheck, used makefile targets |
| `_bmad-output/sprint-status.yaml` | MODIFIED - Updated story status |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `.github/workflows/ci.yml` - MODIFIED
- `_bmad-output/sprint-status.yaml` - MODIFIED
- `_bmad-output/6-6-ci-gates-for-quality.md` - MODIFIED

### Senior Developer Review (AI)

- **Date:** 2025-12-26
- **Reviewer:** BMad
- **Outcome:** Approved
- **Findings:**
    - **Medium:** Undocumented changes in sprint status (Fixed)
    - **Low:** `govulncheck` unpinned (Fixed: pinned to v1.1.4)
    - **Low:** Duplicated migration logic in CI (Fixed: used `make migrate-up/down`)
