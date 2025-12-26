# Epic 4 Retrospective: API Contract & Reliability

**Date:** 2025-12-25
**Epic Status:** ✅ Complete (8/8 stories done)

---

## 🎯 Epic Summary

Epic 4 focused on hardening the API contract and improving reliability. All 8 stories were completed successfully.

---

## ✅ Stories Completed

| Story | Title | Key Outcome |
|-------|-------|-------------|
| 4.1 | Reject JSON with Unknown Fields | `DecodeJSONStrict` with `DisallowUnknownFields()` |
| 4.2 | Reject JSON with Trailing Data | Extended decoder with `More()` check |
| 4.3 | Non-Mutating Health Endpoint | Verified `/ready` is idempotent |
| 4.4 | Configure HTTP Timeouts | Added `ReadHeaderTimeout`, `MaxHeaderBytes` |
| 4.5 | Implement Graceful Shutdown | Verified SIGTERM handling, tracer flush |
| 4.6 | Return Location Header on 201 | Added Location header with resource path injection |
| 4.7 | Generate OpenAPI 3.1 Spec | Created `docs/openapi.yaml`, `make openapi` |
| 4.8 | OpenAPI Contract Tests in CI | Added Spectral validation to CI workflow |

---

## 🏆 Key Achievements

1. **Strict JSON Parsing**: No longer accepts malformed requests (unknown fields, trailing data)
2. **Slowloris Mitigation**: HTTP header timeouts configured system-wide
3. **OpenAPI Documentation**: Complete API spec with all endpoints, schemas, and examples
4. **CI Enhancements**: OpenAPI validation runs on every push
5. **Clean Architecture**: `resourcePath` injection broke handler-router coupling

---

## 📚 Lessons Learned

### What Went Well
- **Verification-first stories** (4.3, 4.5) saved time by confirming existing implementations
- **Docker-based tools** in Makefile enabled consistent validation across environments
- **Layered refactoring** (4.6: Location header → resourcePath injection → BasePath constant)

### What Could Be Improved
- **Local tool availability**: Spectral/npx issues required Docker fallback
- **OpenAPI validation tooling**: Should be established earlier in project lifecycle
- **Version pinning**: Discovered need to pin golangci-lint, goose, spectral versions

### Technical Debt Addressed
- Removed hardcoded `/api/v1` strings across codebase
- Pinned tool versions in Makefile and CI for reproducibility
- Fixed `$(PWD)` vs `$(CURDIR)` inconsistency

---

## 📋 Action Items for Future Epics

| Action | Priority | Target |
|--------|----------|--------|
| Add response contract tests (kin-openapi) | Medium | Epic 6+ |
| Explore OpenAPI code generation | Low | Future |
| Document tool version pinning policy | Low | README |

---

## 📊 Metrics

- **Stories Completed:** 8/8 (100%)
- **Verification-Only Stories:** 2 (4.3, 4.5)
- **New Implementation Stories:** 6
- **Files Created:** 3 (openapi.yaml, spectral.yaml, etc.)
- **Files Modified:** 10+

---

## 🔗 Related Documents

- [Epic 3 Retrospective](./epic-3-retrospective.md)
- [OpenAPI Spec](../docs/openapi.yaml)
- [Architecture](./architecture.md)
