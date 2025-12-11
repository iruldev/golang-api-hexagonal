# Epic 4 Retrospective: Database & Persistence

**Date:** 2025-12-11  
**Epic Status:** ✅ Complete (7/7 stories)

---

## 📊 Summary

Epic 4 established the database layer with PostgreSQL integration, migrations, type-safe queries, and health checks.

| Metric | Value |
|--------|-------|
| Stories completed | 7/7 (100%) |
| New files created | ~20 |
| Test coverage | All tests passing |
| Lint issues | 0 |

---

## ✅ What Went Well

### 1. **pgx/v5 Native Integration**
Used pgx/v5 directly instead of database/sql wrapper. Benefits:
- Better performance (no reflection)
- Native PostgreSQL types (UUID, TIMESTAMPTZ)
- Connection pooling built-in

### 2. **Type-Safe Queries with sqlc**
sqlc generates Go code from SQL. Benefits:
- Compile-time type checking
- No ORM overhead
- Queries in migration files = single source of truth

### 3. **Domain-Separated Code Generation**
sqlc configured for multi-output:
```
postgres/users/  ← From users.sql
postgres/note/   ← From note.sql
```
Clean hexagonal architecture compliance.

### 4. **Makefile Migration Targets**
Consistent developer experience:
```bash
make migrate-up
make migrate-down N=2
make migrate-create NAME=x
```

### 5. **Readiness Probe with DB Check**
`/readyz` endpoint pings database. Returns 503 when DB is down.

---

## 🔧 What Could Be Improved

### 1. **Duplicate Schema Issue (Story 4.4)**
Initially had `db/schema/` AND `db/migrations/` with same content.
**Resolution:** Removed duplicate, migrations are now single source of truth.

### 2. **RouterDeps Breaking Change (Story 4.7)**
Changed `NewRouter(cfg)` to `NewRouter(RouterDeps{...})`.
Required updating 10+ test files.
**Lesson:** Consider backward-compatible changes first.

### 3. **Duplicate Task Entries in Stories**
Multiple stories had duplicate task lists (copy-paste artifact).
**Lesson:** Better story document cleanup before review.

---

## 📚 Lessons Learned

1. **Migrations as Schema Source** - Point sqlc to migrations, not separate schema files
2. **Interface in Handler Package** - Defined DBHealthChecker in handlers (not infra) for clean dependencies
3. **RouterDeps Pattern** - Good for optional dependencies, but plan for test updates
4. **N Parameter in Make** - `$(or $(N),1)` pattern for optional Make parameters

---

## 🔮 Impact on Future Epics

### Epic 5: Observability Suite
- DB health check already integrated with readiness probe ✓
- Logger already wired in router ✓
- Ready for metrics and tracing implementation

### Future Database Work
- Add more query files in `db/queries/`
- Run `make gen` to generate
- Output automatically goes to `postgres/{domain}/`

---

## 📁 Key Files Created

| File | Purpose |
|------|---------|
| `internal/infra/postgres/postgres.go` | Connection pooling |
| `internal/infra/postgres/timeout.go` | Query timeout helpers |
| `internal/infra/postgres/health.go` | DB health checker |
| `internal/infra/postgres/users/` | sqlc generated |
| `internal/infra/postgres/note/` | sqlc generated |
| `db/migrations/000001_*.sql` | Users table |
| `db/migrations/000002_*.sql` | Notes table |
| `db/queries/users.sql` | User CRUD |
| `db/queries/note.sql` | Note CRUD |
| `sqlc.yaml` | Multi-output config |

---

## 🎯 Action Items for Next Epic

1. ~None blocking~ - Ready to proceed
2. Consider adding database connection metrics (Epic 5)
3. Consider adding query timing metrics (Epic 5)
