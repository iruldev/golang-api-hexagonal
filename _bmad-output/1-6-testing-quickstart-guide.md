# Story 1.6: Testing Quickstart Guide

Status: done

## Story

As a **new developer**,
I want a one-page testing guide,
so that I understand the test structure in <30 minutes.

## Acceptance Criteria

1. **AC1:** `docs/testing-quickstart.md` exists (≤1 page)
2. **AC2:** Covers: directory structure, make targets, naming conventions
3. **AC3:** Includes copy-paste examples for unit and integration tests
4. **AC4:** Links to detailed docs for advanced topics

## Tasks / Subtasks

- [x] Task 1: Create testing-quickstart.md (AC: #1, #2)
  - [x] Create `docs/testing-quickstart.md`
  - [x] Document test directory structure
  - [x] Document make targets (test, test-unit, test-shuffle, test-integration)
  - [x] Document naming conventions
- [x] Task 2: Add code examples (AC: #3)
  - [x] Add copy-paste unit test example
  - [x] Add copy-paste integration test example
  - [x] Add mock usage example
- [x] Task 3: Add links and references (AC: #4)
  - [x] Link to testutil package docs
  - [x] Link to architecture.md testing section
  - [x] Link to goleak docs

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **Pattern 1:** Test organization and naming conventions
- **Pattern 2:** Mock conventions with uber-go/mock
- **Pattern 3:** TestMain with goleak

### Document Structure

```markdown
# Testing Quickstart Guide

## Directory Structure
internal/
├── testutil/           # Shared test helpers
│   ├── assert/         # Assertion helpers
│   ├── containers/     # Testcontainers helpers
│   ├── fixtures/       # Test data builders
│   └── mocks/          # Generated mocks
├── domain/*_test.go    # Unit tests
├── app/*_test.go       # Use case tests
└── infra/postgres/     # Integration tests

## Make Targets
| Command | Description |
|---------|-------------|
| make test | Run all tests with coverage + shuffle |
| make test-unit | Run unit tests only |
| make test-shuffle | Run with shuffle enabled |
| make test-integration | Run integration tests |

## Naming Conventions
- Files: `*_test.go`
- Functions: `Test<Unit>_<Scenario>` or `Test<Unit>_<Scenario>_<Expected>`
- Table-driven: Use `cases` or `tests` variable

## Example: Unit Test
[copy-paste example]

## Example: Integration Test with TestMain
[copy-paste example]

## Further Reading
- [TestUtil Package](../internal/testutil/)
- [Architecture Testing Patterns](../_bmad-output/architecture.md)
```

### Testing Standards

- Maximum ~1 printed page (~60 lines)
- Copy-paste examples must compile
- Use existing project patterns

### Previous Story Learnings (Story 1.1-1.5)

- testutil package exists with TestContext, RunWithGoleak
- mocks generated via uber-go/mock
- Make targets: test, test-unit, test-shuffle, test-integration, gencheck
- postgres/main_test.go has TestMain with goleak

### References

- [Source: _bmad-output/architecture.md testing patterns]
- [Source: _bmad-output/epics.md#Story 1.6]
- [Source: _bmad-output/prd.md#FR25, NFR17]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### File List

_Files created/modified during implementation:_
- [x] `docs/testing-quickstart.md` (new file)
