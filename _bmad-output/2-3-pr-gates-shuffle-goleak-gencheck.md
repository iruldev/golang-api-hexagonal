# Story 2.3: PR Gates (shuffle + goleak + gencheck)

Status: done

## Story

As a **CI system**,
I want quality gates on every PR,
so that hidden coupling and leaks are caught early.

## Acceptance Criteria

1. **AC1:** GitHub Actions runs `make test-shuffle` on PRs
2. **AC2:** GitHub Actions runs `make gencheck` on PRs
3. **AC3:** goleak verification happens via TestMain
4. **AC4:** Failed gates block PR merge

## Tasks / Subtasks

- [x] Task 1: Add test-shuffle to CI workflow (AC: #1)
  - [x] Add `make test-shuffle` step to `.github/workflows/ci.yml`
  - [x] Run after lint step
  - [x] Ensure shuffle failures block PR
- [x] Task 2: Add gencheck to CI workflow (AC: #2)
  - [x] Add `make gencheck` step to `.github/workflows/ci.yml`
  - [x] Verifies all go:generate files are up-to-date
  - [x] Runs after test step
- [x] Task 3: Verify goleak integration (AC: #3)
  - [x] Ensure TestMain in postgres package uses goleak
  - [x] Document TestMain pattern in story
  - [x] Verify leaks fail tests (already implemented in Story 1.3)
- [x] Task 4: Verify PR blocking (AC: #4)
  - [x] Test that failed gates block PR merge
  - [x] Configure branch protection if not already set

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-005:** CI Quality Gates
- **NFR-6:** CI enforcement of quality

### Current CI Workflow

Existing `.github/workflows/ci.yml` already has:
- ✅ golangci-lint
- ✅ OpenAPI validation
- ✅ Generated code check (sqlc)
- ⚠️ Missing: test-shuffle
- ⚠️ Missing: gencheck (mocks)

### CI Steps to Add

```yaml
# After lint step, add:

- name: Run tests with shuffle
  run: make test-shuffle

- name: Verify generated code (mocks)
  run: make gencheck
```

### goleak via TestMain (Already Implemented)

From Story 1.3, `testutil.RunWithGoleak(m)` is already integrated:

```go
// internal/infra/postgres/main_test.go
func TestMain(m *testing.M) {
    testutil.RunWithGoleak(m)
}
```

### Testing Standards

- CI changes verified by pushing to test branch
- Verify test-shuffle fails with order-dependent tests
- Verify gencheck fails with stale mocks

### Previous Story Learnings (Story 1.4, 2.1, 2.2)

- test-shuffle target exists (Story 1.4)
- gencheck target exists (Story 1.4)
- goleak already integrated (Story 1.3)

### References

- [Source: _bmad-output/architecture.md#AD-005 CI Quality Gates]
- [Source: _bmad-output/epics.md#Story 2.3]
- [Source: _bmad-output/prd.md#FR11, FR12, FR14, NFR6]
- [.github/workflows/ci.yml](../../.github/workflows/ci.yml)

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### File List

_Files created/modified during implementation:_
- [ ] `.github/workflows/ci.yml` (add test-shuffle and gencheck steps)
