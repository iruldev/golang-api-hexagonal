# Story 6.2: Implement make test

Status: done

## Story

**As a** developer,
**I want** `make test` to run all unit tests,
**So that** I can verify changes quickly.

**FR:** FR26

## Acceptance Criteria

1. ✅ **Given** codebase with tests, **When** `make test` is run, **Then** all unit tests execute
2. ✅ **Given** tests run, **When** complete, **Then** coverage report is generated (`coverage.out`)
3. ✅ **Given** tests run, **When** complete, **Then** exit code reflects test status

## Implementation Status from Audit

> [!TIP]
> **ALREADY IMPLEMENTED** - Current `make test` satisfies all ACs.

Current implementation (Makefile line 102-105):
```makefile
## test: Run all tests
.PHONY: test
test:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

**AC Analysis:**
- **AC#1** ✅ - `./...` runs all unit tests
- **AC#2** ✅ - `-coverprofile=coverage.out` generates coverage
- **AC#3** ✅ - Go test naturally exits non-zero on failure

## Tasks / Subtasks

- [x] Task 1: Verify existing implementation (AC: #1, #2, #3) ✅
  - All acceptance criteria already satisfied

## Dev Notes

### Additional Coverage Target

There's also a `make coverage` target (lines 107-126) that:
- Runs tests for `domain` and `app` packages only
- Enforces 80% coverage threshold
- Shows coverage summary

## Changes

| File | Change |
|------|--------|
| Makefile | Improved `test` target with `ARGS` support and `clean` target hygiene |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `Makefile`

## Senior Developer Review (AI)

_Reviewer: BMad (AI)_

### Findings
- **Low Issue (Usability):** `make test` was hardcoded to run all tests without option for arguments.
- **Low Issue (Maintainability):** `make clean` did not remove `coverage.out`.

### Actions Taken
- ✅ **Fixed:** Updated `test` target to accept `ARGS` (e.g., `make test ARGS="-run MyTest"`).
- ✅ **Fixed:** Updated `clean` target to remove `coverage.out`.
- **Status:** Story verified as **done**. All ACs met and usability improved.
