# Story 3.4: Adoption Guide + Copy-Paste Kit

Status: done

## Story

As a **adopting team lead**,
I want a complete adoption guide,
so that I can bring these patterns to my service.

## Acceptance Criteria

1. **AC1:** `docs/adoption-guide.md` with checklist
2. **AC2:** Copy-paste kit: `internal/testutil/`, Makefile snippets, CI workflow
3. **AC3:** Step-by-step migration guide for brownfield
4. **AC4:** Expected adoption time: ≤1 day

## Tasks / Subtasks

- [x] Task 1: Create adoption guide document (AC: #1, #4)
  - [x] Create `docs/adoption-guide.md`
  - [x] Add adoption checklist
  - [x] Include time estimates per step
  - [x] Target ≤1 day total
- [x] Task 2: Prepare copy-paste kit (AC: #2)
  - [x] Document `internal/testutil/` setup
  - [x] Include Makefile test target snippets
  - [x] Include CI workflow examples
- [x] Task 3: Write brownfield migration guide (AC: #3)
  - [x] Step-by-step migration instructions
  - [x] Address common migration challenges
  - [x] Include rollback strategies

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **FR-26:** Adoption guide
- **FR-27:** Copy-paste kit
- **NFR-20:** Developer experience

### Existing Assets to Document

| Component | Path | Purpose |
|-----------|------|---------|
| Test utilities | `internal/testutil/` | goleak TestMain, containers |
| Container helpers | `internal/testutil/containers/` | PostgreSQL, migrations, tx |
| Domain errors | `internal/domain/errors/` | Error types with stable codes |
| Error mapping | `internal/transport/http/contract/error.go` | RFC 7807 responses |
| Makefile | `Makefile` | Test targets (test, test-shuffle, etc.) |
| PR CI | `.github/workflows/ci.yml` | PR quality gates |
| Nightly CI | `.github/workflows/nightly.yml` | Race detection, integration |

### Adoption Guide Structure

```markdown
# Adoption Guide

## Prerequisites
- Go 1.21+
- Docker (for testcontainers)
- PostgreSQL knowledge

## Quick Start (30 min)
1. Copy `internal/testutil/`
2. Add Makefile targets
3. Run first test

## Full Integration (4-6 hours)
1. Test utilities
2. Container helpers
3. Error handling
4. CI workflows

## Checklist
- [ ] testutil/testutil.go
- [ ] testutil/containers/
- [ ] domain/errors/
- [ ] Makefile targets
- [ ] CI workflows
```

### Copy-Paste Kit Contents

```
copy-paste-kit/
├── testutil/
│   ├── testutil.go          # TestMain with goleak
│   └── containers/
│       ├── postgres.go      # NewPostgres(t)
│       ├── migrate.go       # Migrate(t, pool)
│       ├── tx.go            # WithTx(t, pool, fn)
│       └── truncate.go      # Truncate(t, pool, tables)
├── domain-errors/
│   ├── errors.go            # DomainError type
│   └── codes.go             # Stable error codes
├── makefile-snippets.md     # Test targets
└── ci-workflow.yml          # GitHub Actions
```

### Brownfield Migration Guide

Key sections:
1. **Assessment** - Evaluate current test setup
2. **Incremental adoption** - Start with new tests
3. **Migration steps** - Convert existing tests
4. **Rollback plan** - Revert if issues

### Time Estimates

| Step | Time |
|------|------|
| Copy testutil | 15 min |
| Configure containers | 30 min |
| Migrate first test | 45 min |
| Add error types | 1 hour |
| CI setup | 1 hour |
| **Total** | **~4 hours** |

### References

- [Source: _bmad-output/architecture.md#Developer Experience]
- [Source: _bmad-output/epics.md#Story 3.4]
- [Source: _bmad-output/prd.md#FR26, FR27, NFR20]
- [Existing: internal/testutil/]
- [Existing: docs/]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### File List

_Files created/modified during implementation:_
- [x] `docs/adoption-guide.md` (new - 607 lines)
- [x] `docs/copy-paste-kit/` (new directory with testutil/, domain-errors/)
- [x] `_bmad-output/sprint-status.yaml` (status sync)

## Senior Developer Review (AI)

_Reviewer: @bmad-bmm-workflows-code-review on 2025-12-28_

### Findings
- **[SUCCESS]**: Comprehensive adoption guide created with all required sections
- **[SUCCESS]**: Copy-paste kit includes all code snippets inline in adoption-guide.md
- **[SUCCESS]**: Brownfield migration guide with assessment, incremental strategy, rollback
- **[SUCCESS]**: Adoption time estimated at 4-6 hours (within ≤1 day AC4)
- **[HIGH] Fixed**: Story tasks were unmarked - marked all as complete
- **[MEDIUM] Fixed**: `docs/copy-paste-kit/testutil/containers/truncate.go` was untracked in git. Added to git.

### Outcome
**Approved** with automated fixes applied.

### Fixes
- **[HIGH] Fixed**: Missing `truncate.go` in copy-paste kit and adoption guide. Copied from reference and updated guide.
- **[MEDIUM] Fixed**: Added untracked `truncate.go` to git.
