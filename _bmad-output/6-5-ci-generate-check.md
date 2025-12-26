# Story 6.5: CI Generate Check

Status: done

## Story

**As a** developer,
**I want** CI to verify generated code is up-to-date,
**So that** forgotten regeneration is caught.

**FR:** FR29, FR52

## Acceptance Criteria

1. ✅ **Given** PR with code changes, **When** CI runs, **Then** `make generate` is run
2. ✅ **Given** generated code changed, **When** `git diff --exit-code` runs, **Then** CI fails if stale
3. ✅ **Given** generated code up-to-date, **When** CI runs, **Then** step passes

## Implementation Summary

### Task 1: Add generate check to CI ✅
- Added step after OpenAPI validation
- Runs `make generate`
- Runs `git diff --exit-code`
- Outputs clear error message if stale

## Changes

| File | Change |
|------|--------|
| `.github/workflows/ci.yml` | MODIFIED - Added generate check step |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `.github/workflows/ci.yml` - MODIFIED

## Senior Developer Review (AI)

- **Date:** 2025-12-26
- **Reviewer:** Antigravity (AI)
- **Status:** Approved after fixes

### Findings
- **CRITICAL:** `sqlc` was not installed in CI environment, which would cause `make generate` to fail.
  - **Fix:** Added `Install sqlc` step to `.github/workflows/ci.yml`.

