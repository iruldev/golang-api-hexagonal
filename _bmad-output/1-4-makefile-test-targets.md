# Story 1.4: Makefile Test Targets

Status: done

## Story

As a **developer**,
I want simple make commands for testing,
so that I don't need to remember go test flags.

## Acceptance Criteria

1. **AC1:** `make test-unit` runs unit tests with coverage
2. **AC2:** `make test-shuffle` runs with `-shuffle=on`
3. **AC3:** `make gencheck` verifies generated files are up-to-date
4. **AC4:** `make test` runs unit + shuffle
5. **AC5:** Coverage report generated to `coverage.out`

## Tasks / Subtasks

- [x] Task 1: Add test-unit target (AC: #1, #5)
  - [x] Add `test-unit` target to Makefile
  - [x] Runs `go test` with `-coverprofile=coverage.out`
  - [x] Excludes integration tests (no `-tags=integration`)
- [x] Task 2: Add test-shuffle target (AC: #2)
  - [x] Add `test-shuffle` target to Makefile
  - [x] Runs with `-shuffle=on` flag
  - [x] Ensures deterministic test isolation
- [x] Task 3: Add gencheck target (AC: #3)
  - [x] Add `gencheck` target to Makefile
  - [x] Runs `go generate ./...` and checks for diffs
  - [x] Fails if generated files are out of sync
- [x] Task 4: Update test target (AC: #4)
  - [x] Modify existing `test` target to run unit + shuffle
  - [x] Or create combined workflow

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-005:** CI Quality Gates include shuffle and gencheck
- **Pattern 5:** Test execution patterns

### Existing Makefile Targets

Current test-related targets in Makefile:
- `test` - runs all tests with race detection
- `test-integration` - runs integration tests
- `coverage` - runs coverage check

### New Targets to Add

```makefile
## test-unit: Run unit tests with coverage
.PHONY: test-unit
test-unit:
	@echo "🧪 Running unit tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "✅ Unit tests complete"

## test-shuffle: Run tests with shuffle enabled
.PHONY: test-shuffle
test-shuffle:
	@echo "🔀 Running tests with shuffle..."
	$(GOTEST) -v -race -shuffle=on ./...
	@echo "✅ Shuffle tests complete"

## gencheck: Verify generated files are up-to-date
.PHONY: gencheck
gencheck:
	@echo "🔍 Checking generated files..."
	@go generate ./...
	@if git diff --exit-code --quiet; then \
		echo "✅ Generated files are up-to-date"; \
	else \
		echo "❌ Generated files are out of sync. Run 'go generate ./...' and commit changes."; \
		git diff --stat; \
		exit 1; \
	fi
```

### Testing Standards

- Run `make test-unit` to verify unit tests work
- Run `make test-shuffle` to verify shuffle works
- Run `make gencheck` to verify gencheck works
- Verify `coverage.out` is generated after test-unit

### Previous Story Learnings (Story 1.1-1.3)

- Makefile already has `mocks` target (Story 1.2)
- Use consistent emoji and messaging pattern
- Use `$(GOTEST)` variable for consistency

### References

- [Source: _bmad-output/architecture.md#AD-005 CI Quality Gates]
- [Source: _bmad-output/epics.md#Story 1.4]
- [Source: _bmad-output/prd.md#FR2, FR11]

## Dev Agent Record

### Agent Model Used

### Agent Model Used

Antigravity (Google Deepmind)

### Debug Log References

- Verified `test-unit` runs coverage
- Verified `test-shuffle` runs shuffle
- Verified `gencheck` detects dirty state
- Verified `test` runs shuffle + coverage

### Completion Notes List

- All Makefile targets implemented and verified.
- `make test` updated to include shuffle.
- `make help` fixed to hide Story 6.3 note.

### File List

_Files created/modified during implementation:_
- [x] `Makefile` (add test-unit, test-shuffle, gencheck targets)
