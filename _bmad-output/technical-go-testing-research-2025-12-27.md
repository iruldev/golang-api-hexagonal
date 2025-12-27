# Technical Research: Go Testing Architecture Best Practices (2025)

> **Research Date:** 2025-12-27  
> **Research Type:** Technical  
> **Topic:** Go Testing Architecture Best Practices  
> **Sources:** 16 verified citations from Go official docs, Google, Uber, Testcontainers

---

## Executive Summary

Riset ini mengumpulkan best practices terkini untuk testing architecture di Go, dengan fokus pada:
- Test organization patterns
- Mock strategy & consolidation
- Integration tests dengan Testcontainers
- Concurrency testing (Go 1.25 synctest)
- Race & goroutine leak prevention

**Key Decisions:**
1. ✅ Adopt `go-cmp` untuk equality checks
2. ✅ Adopt `uber-go/mock` (maintained fork)
3. ✅ Adopt `testcontainers-go` untuk integration tests
4. ✅ Leverage Go 1.25 `testing/synctest` untuk concurrency
5. ✅ Add `goleak` untuk goroutine leak detection

---

## Table of Contents

1. [Test Organization Patterns](#1-test-organization-patterns)
2. [Equality & Diff Standard](#2-equality--diff-standard)
3. [Mock Strategy 2025](#3-mock-strategy-2025)
4. [Integration Tests dengan Testcontainers](#4-integration-tests-dengan-testcontainers)
5. [Go 1.25 synctest](#5-go-125-synctest)
6. [Race & Goroutine Leak Prevention](#6-race--goroutine-leak-prevention)
7. [Recommended Dependencies](#7-recommended-dependencies)
8. [CI Quality Gates](#8-ci-quality-gates)
9. [Architecture Decision Records](#9-architecture-decision-records)

---

## 1. Test Organization Patterns

### Finding: Table-driven tests + subtests adalah standar

Pattern yang paling konsisten dipakai di ekosistem Go:

- **Table-driven tests + `t.Run`** untuk:
  - Output lebih readable
  - Bisa run test case tertentu via `-run`
  - Menghindari "fail fast" yang bikin debugging sulit
  
- **Prinsip "keep going"**: 
  - Prefer `t.Error` untuk mismatch supaya semua perbedaan terlihat
  - Pakai `t.Fatal` terutama untuk setup yang kalau gagal test tidak bisa lanjut

- **Test helpers** sebaiknya **return error/value**, bukan menerima `*testing.T` untuk nge-fail dari dalam helper

**Sources:**
- [Using Subtests and Sub-benchmarks - Go Blog](https://go.dev/blog/subtests)
- [Go Wiki: Go Test Comments](https://tip.golang.org/wiki/TestComments)
- [Google Go Style Guide - Best Practices](https://google.github.io/styleguide/go/best-practices.html)

### Blueprint: Target Folder Structure

```
internal/testutil/
├── fixtures/           # Builders, factory data
│   ├── user.go
│   └── audit.go
├── assert/             # Custom assertions (wrap cmp.Diff)
│   └── http.go
├── mocks/              # Generated mocks (satu lokasi)
│   ├── user_repo.go
│   ├── audit_repo.go
│   └── generate.go     # go:generate directives
└── containers/         # Testcontainers setup
    └── postgres.go

testdata/               # Static test data (Go tool treats specially)
```

---

## 2. Equality & Diff Standard

### Finding: `go-cmp` adalah standar modern, bukan `reflect.DeepEqual`

Google Go community docs secara eksplisit merekomendasikan `cmp.Equal` / `cmp.Diff` karena:
- Stabilitas hasil diff
- Menghindari sensitivitas `reflect.DeepEqual` terhadap unexported fields
- Output yang lebih readable

**Sources:**
- [Google Go Test Comments - Equality Comparison](https://go.googlesource.com/wiki/+show/f1b5852e36015129577061df5d48fc2d80aaf387/TestComments.md)
- [google/go-cmp GitHub](https://github.com/google/go-cmp)

### Decision

```go
// ✅ Use this
import "github.com/google/go-cmp/cmp"

if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("mismatch (-want +got):\n%s", diff)
}

// ❌ Avoid this
if !reflect.DeepEqual(want, got) {
    t.Errorf("want %v, got %v", want, got)
}
```

---

## 3. Mock Strategy 2025

### Finding: `golang/mock` archived, `uber-go/mock` adalah fork maintained

- Repo `golang/mock` sudah **archived** 
- README mengarahkan ke fork `uber-go/mock`
- Uber fork supports "dua rilis Go terbaru" mengikuti Go release policy

**Sources:**
- [golang/mock (archived)](https://github.com/golang/mock)
- [uber-go/mock (maintained)](https://github.com/uber-go/mock)

### Decision: Hybrid Approach

| Interface Size | Strategy | Rationale |
|----------------|----------|-----------|
| Small (< 5 methods) | Manual fakes/stubs | Lebih readable, less brittle |
| Large (≥ 5 methods) | Generated mocks | Efficiency, consistency |

**Implementation:**
- Semua generated mocks di `internal/testutil/mocks/`
- `go:generate` untuk mockgen
- CI check "generated files up-to-date"

```bash
# Install
go install go.uber.org/mock/mockgen@latest

# Generate (in mocks/generate.go)
//go:generate mockgen -destination=user_repo.go -package=mocks github.com/iruldev/golang-api-hexagonal/internal/domain UserRepository
```

---

## 4. Integration Tests dengan Testcontainers

### Finding: Testcontainers + Wait Strategies = Anti-Flaky

Testcontainers for Go menyediakan:
- **Module Postgres** dengan best practice setup
- **Wait strategies** agar tidak flaky (log readiness + listening port)
- Default timeout 60s, poll interval 100ms (bisa override)

**⚠️ Important:** Reusable containers **tidak cocok untuk CI** (hanya untuk local dev)

**Sources:**
- [Testcontainers Postgres Module](https://golang.testcontainers.org/modules/postgres/)
- [Wait Strategy Introduction](https://golang.testcontainers.org/features/wait/introduction/)
- [Testcontainers Desktop Docs](https://testcontainers.com/desktop/docs/)

### Implementation

```go
// internal/testutil/containers/postgres.go
package containers

import (
    "context"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

func NewPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
    return postgres.Run(ctx,
        "postgres:15-alpine",
        postgres.WithDatabase("test_db"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready").
                WithOccurrence(2).
                WithStartupTimeout(60*time.Second),
        ),
    )
}
```

---

## 5. Go 1.25 synctest

### Finding: `testing/synctest` untuk deterministic concurrency tests

Go 1.25 membawa `testing/synctest` untuk:
- Menguji kode concurrent dengan "bubble" + fake clock
- Mengurangi flakiness time-based tests
- Menggantikan sleep-based tests

**Sources:**
- [Go 1.25 Release Notes](https://tip.golang.org/doc/go1.25)
- [Testing concurrent code with testing/synctest - Go Blog](https://go.dev/blog/synctest)

### Use Cases

- Code dengan goroutines + timers
- Context deadline handling
- Background job testing

```go
import "testing/synctest"

func TestTimeout(t *testing.T) {
    synctest.Run(func() {
        ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()
        
        // Fake clock advances instantly
        time.Sleep(2 * time.Second)
        
        if ctx.Err() == nil {
            t.Error("expected context to be cancelled")
        }
    })
}
```

---

## 6. Race & Goroutine Leak Prevention

### Finding: `-race` + `goleak` adalah standar industry

**Race Detector:**
- Menginstrument memory access untuk deteksi data race
- Cocok untuk integration/load tests (overhead bisa besar)
- Tidak untuk production

**Goroutine Leak Detection:**
- `goleak` dari Uber bisa dipasang per-test atau package-level (`TestMain`)
- Untuk `t.Parallel`, lebih aman pakai `VerifyTestMain`

**Sources:**
- [Introducing the Go Race Detector - Go Blog](https://go.dev/blog/race-detector)
- [uber-go/goleak GitHub](https://github.com/uber-go/goleak)

### Implementation

```go
// In package test file
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}

// Or per-test
func TestSomething(t *testing.T) {
    defer goleak.VerifyNone(t)
    // test code
}
```

---

## 7. Recommended Dependencies

### Must-Have (High ROI)

| Package | Purpose | Source |
|---------|---------|--------|
| `github.com/google/go-cmp/cmp` | Diff yang jelas, stabil | [GitHub](https://github.com/google/go-cmp) |
| `go.uber.org/mock` + `mockgen` | Mock maintained | [GitHub](https://github.com/uber-go/mock) |
| `go.uber.org/goleak` | Anti goroutine leak | [GitHub](https://github.com/uber-go/goleak) |
| `testcontainers-go` + postgres module | Integration tests | [Docs](https://golang.testcontainers.org/modules/postgres/) |

### Nice to Have

| Package | Purpose | When |
|---------|---------|------|
| `testing/synctest` (stdlib Go 1.25) | Concurrency testing | Timer/context heavy code |

---

## 8. CI Quality Gates

### Recommended Test Commands

```makefile
## test-unit: Fast unit tests
test-unit:
	go test -short ./...

## test-shuffle: Detect hidden dependencies
test-shuffle:
	go test -shuffle=on ./...

## test-race: Race detection (selective/nightly)
test-race:
	go test -race ./...

## test-integration: With testcontainers
test-integration:
	go test -tags=integration -race ./...

## test-nocache: Force fresh run
test-nocache:
	go test -count=1 ./...
```

---

## 9. Architecture Decision Records

### ADR-001: Test Equality Standard

**Context:** Perlu standar untuk compare values di tests.

**Decision:** Adopt `github.com/google/go-cmp` untuk semua equality checks.

**Consequences:**
- ✅ Output diff yang readable
- ✅ Safe untuk unexported fields dengan cmpopts
- ⚠️ Tambah dependency

### ADR-002: Mock Generation Standard

**Context:** Mocks tersebar dan inkonsisten.

**Decision:** 
- Generated mocks dengan `uber-go/mock` untuk interface besar
- Manual fakes untuk interface kecil
- Semua mocks di `internal/testutil/mocks/`

**Consequences:**
- ✅ Konsistensi dan maintainability
- ✅ CI dapat verify generated files
- ⚠️ Perlu go:generate discipline

### ADR-003: Integration Test Infrastructure

**Context:** DB integration tests flaky atau tidak ada.

**Decision:** Adopt `testcontainers-go` dengan explicit wait strategies.

**Consequences:**
- ✅ Reproducible di local dan CI
- ✅ Clean slate per test
- ⚠️ Docker required di CI

### ADR-004: Concurrency Testing Standard

**Context:** Time-based tests flaky.

**Decision:** Gunakan `testing/synctest` (Go 1.25) untuk timer/context code.

**Consequences:**
- ✅ Deterministic tests
- ✅ Fast execution (no real sleeps)
- ⚠️ Requires Go 1.25+

### ADR-005: Reliability Quality Gates

**Context:** Need to catch subtle concurrency bugs.

**Decision:** 
- `-shuffle=on` untuk detect hidden coupling
- `-race` selective/nightly
- `goleak` in TestMain

**Consequences:**
- ✅ Higher confidence in code quality
- ⚠️ Longer CI times (race detector)

---

## Conclusion

Riset ini memberikan blueprint lengkap untuk testing architecture yang "international-grade":

1. **Structure**: Centralized testutil dengan fixtures, mocks, containers
2. **Equality**: go-cmp instead of reflect.DeepEqual
3. **Mocks**: uber-go/mock dengan centralized location
4. **Integration**: testcontainers dengan wait strategies
5. **Concurrency**: synctest untuk deterministic tests
6. **Reliability**: shuffle, race, goleak

Semua keputusan ini bisa langsung diimplementasikan dan di-track via ADRs.

---

*Research conducted by BMad Research Workflow*
*Sources verified via web search with 16 citations*
