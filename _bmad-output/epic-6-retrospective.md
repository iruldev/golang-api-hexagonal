# Epic 6 Retrospective: Developer Experience & Build System

**Date:** 2025-12-26
**Epic Status:** ✅ Complete (7/7 stories done)
**Priority:** P2 (Growth)

---

## Executive Summary

Epic 6 focused on improving **Developer Experience and Build System** robustness. All 7 stories were completed successfully, with a major pivot from Google Wire to Uber Fx for dependency injection.

### Key Metrics
| Metric | Value |
|--------|-------|
| Stories Completed | 7/7 (100%) |
| Stories with Reviews | 7 |
| Major Blockers | 1 (Wire/pgx incompatibility) |
| Pivots Made | 1 (Wire → Uber Fx) |

---

## Stories Completed

| Story | Title | Status | Notes |
|-------|-------|--------|-------|
| 6.1 | Implement make setup | ✅ Done | Tool version pinning, .env.local creation |
| 6.2 | Implement make test | ✅ Done | Already satisfied by existing implementation |
| 6.3 | Integration Tests with Test Database | ✅ Done | Safety guards for _test databases |
| 6.4 | Implement make generate | ✅ Done | sqlc integration verified |
| 6.5 | CI Generate Check | ✅ Done | Added sqlc install + git diff check |
| 6.6 | CI Gates for Quality | ✅ Done | Added govulncheck, gitleaks |
| 6.7 | Wire DI Integration | ✅ Done | Pivoted to Uber Fx due to Wire blocker |

---

## What Went Well 🎉

### 1. Makefile Improvements (Stories 6.1-6.4)
- **Tool version pinning** prevents "works on my machine" issues
- **Idempotent setup** with graceful handling of existing installations
- **Safety guards** for test databases prevent accidental data loss

### 2. CI Quality Gates (Stories 6.5-6.6)
- **Fail-fast strategy**: secrets → lint → generate → vuln → tests
- **govulncheck** catches CVEs before deployment
- **gitleaks** prevents secret leakage
- **Generate check** ensures sqlc changes are committed

### 3. Quick Pivot on Wire Blocker (Story 6.7)
- Identified Wire incompatibility with pgx/puddle early
- Successfully pivoted to Uber Fx in same session
- Maintained all acceptance criteria except compile-time-only validation

---

## Challenges & Lessons 📚

### 1. Wire/pgx Incompatibility
**Issue:** Wire fails with `internal error: package "golang.org/x/sync/semaphore" without types` when pgx/puddle is in the import graph.

**Root Cause:** Wire's `go/packages` type loader cannot resolve transitive dependencies from `puddle/v2` → `sync/semaphore`.

**Lesson:** Always prototype DI tools with the actual tech stack early. Don't assume compatibility.

**Action:** Documented blocker; future projects should evaluate FX first if using pgx.

### 2. CI Generate Check Missing sqlc Install
**Issue:** Initial CI step ran `make generate` without installing sqlc first.

**Lesson:** CI must mirror local environment setup explicitly.

**Action:** Added explicit `Install sqlc` step before generate check.

### 3. Uber Fx Trade-offs
**Trade-off:** Fx uses runtime reflection vs Wire's compile-time generation.

**Mitigation:** DI errors occur at app startup (fast feedback) not at runtime request handling.

**Benefit:** Lifecycle hooks (`OnStart`/`OnStop`) are built-in, simplifying server management.

---

## Technical Debt Incurred

| Item | Severity | Epic 7 Impact | Notes |
|------|----------|---------------|-------|
| Fx uses reflection | Low | None | Errors at startup, not runtime |
| Wire setup files remain | Low | Cleanup | Can remove in future cleanup PR |

---

## Action Items for Epic 7

1. **Remove leftover wire.go** if present in repo
2. **Update CONTRIBUTING.md** to document `make setup` requirements
3. **Add Fx module documentation** to development guide
4. **Consider Fx graph visualization** for onboarding

---

## Next Epic Preview

**Epic 7: Governance & Documentation** (Vision/P3)

### Dependencies on Epic 6
- `make setup` documented in CONTRIBUTING.md
- `make generate` for code generation workflow
- CI gates for quality enforcement

### Preparation Needed
- Review existing documentation structure
- Identify key ADRs to document (Wire→Fx decision is one)
- Template for operational runbook

---

## Retrospective Meta

**What worked in this retro process:**
- Story-level dev notes captured implementation challenges
- Code review workflow caught issues (sqlc install, govulncheck pinning)
- Quick pivot when blocker was identified

**What to improve:**
- Earlier prototype testing of DI frameworks
- More explicit CI environment documentation

---

*Retrospective completed: 2025-12-26*
