# Testing Guide: golang-api-hexagonal

> **Analisis Testing** - Deep Scan Findings  
> **Tanggal:** 2025-12-27  
> **Fokus:** Review struktur testing & rekomendasi perbaikan

---

## 1. Executive Summary

| Metrik | Nilai | Status |
|--------|-------|--------|
| **Total Test Files** | 49 | ✅ Good |
| **Coverage Threshold** | ≥80% (domain+app) | ✅ Enforced |
| **Test Status** | All Passing | ✅ |
| **Packages Without Tests** | 2 (`sqlcgen`, `metrics`) | ⚠️ Note |
| **Integration Tests** | Available | ✅ |
| **Benchmark Tests** | 1 (`redact`) | ✅ |

---

## 2. Test Distribution by Layer

### 2.1 Domain Layer (4 test files)
```
internal/domain/
├── audit_test.go      ✅ Entity validation
├── id_test.go         ✅ ID value object
├── pagination_test.go ✅ Pagination value object
└── user_test.go       ✅ User entity validation
```
**Status:** ✅ Clean - Domain layer well-tested

### 2.2 Application Layer (6 test files)
```
internal/app/
├── auth_test.go       ✅ AuthParser tests
├── errors_test.go     ✅ AppError tests
├── audit/
│   └── service_test.go ✅ AuditService tests
└── user/
    ├── create_user_test.go ✅ CreateUserUseCase
    ├── get_user_test.go    ✅ GetUserUseCase
    └── list_users_test.go  ✅ ListUsersUseCase
```
**Status:** ✅ Clean - All use cases tested with mocks

### 2.3 Transport Layer (18 test files)
```
internal/transport/http/
├── router_test.go              ✅ Router integration
├── contract/
│   ├── error_test.go           ✅ RFC 7807 errors
│   ├── json_test.go            ✅ JSON utilities
│   └── user_test.go            ✅ User DTOs
├── ctxutil/
│   ├── claims_test.go          ✅ JWT context
│   └── trace_test.go           ✅ Trace context
├── handler/
│   ├── health_test.go          ✅ Health endpoint
│   ├── ready_test.go           ✅ Ready endpoint
│   ├── user_test.go            ✅ User handler (22KB - large!)
│   ├── helpers_test.go         ✅ Test utilities
│   ├── integration_test.go     ✅ Integration tests
│   ├── integration_idor_test.go ✅ IDOR security tests
│   └── metrics_audit_test.go   ✅ Metrics + audit
└── middleware/
    ├── auth_test.go            ✅ JWT auth (23KB - large!)
    ├── auth_bridge_test.go     ✅
    ├── auth_test_helper_test.go ✅
    ├── body_limiter_test.go    ✅
    ├── logging_test.go         ✅
    ├── metrics_test.go         ✅
    ├── ratelimit_test.go       ✅
    ├── requestid_test.go       ✅
    ├── response_wrapper_test.go ✅
    ├── security_test.go        ✅
    └── tracing_test.go         ✅
```
**Status:** ⚠️ Tests lengkap tapi beberapa file sangat besar

### 2.4 Infrastructure Layer (15 test files)
```
internal/infra/
├── config/
│   ├── config_test.go              ✅
│   └── config_pool_validation_test.go ✅
├── fx/
│   └── module_test.go              ✅ DI graph tests
├── observability/
│   ├── logger_test.go              ✅
│   ├── metrics_test.go             ✅
│   └── tracer_test.go              ✅
└── postgres/
    ├── pool_test.go                ✅
    ├── resilient_pool_test.go      ✅
    ├── user_repo_test.go           ✅
    ├── audit_event_repo_test.go    ✅
    ├── citext_integration_test.go  ✅ Integration test
    └── test_helpers_test.go        ⚠️ EMPTY (22 bytes)
```
**Status:** ⚠️ Mostly good, tapi ada empty test helper file

### 2.5 Shared Layer (4 test files)
```
internal/shared/
├── metrics/
│   └── (no test files)            ⚠️ Missing
└── redact/
    ├── redactor_test.go           ✅ Comprehensive
    ├── benchmark_test.go          ✅ Performance tests
    └── robustness_test.go         ✅ Edge cases
```
**Status:** ⚠️ `metrics/` package tanpa tests

---

## 3. Issues yang Ditemukan

### 3.1 🔴 Critical Issues

| Issue | Location | Impact |
|-------|----------|--------|
| Empty test helper file | `infra/postgres/test_helpers_test.go` (22 bytes) | Test reference yang tidak dipakai |

### 3.2 🟡 Medium Issues

| Issue | Location | Recommendation |
|-------|----------|----------------|
| Large test file | `handler/user_test.go` (22KB) | Split by test category |
| Large test file | `middleware/auth_test.go` (23KB) | Split by scenario |
| Missing tests | `shared/metrics/` | Add interface tests |
| No integration test runner | Root level | Add `make test-integration` guide |

### 3.3 🟢 Low Priority

| Issue | Location | Note |
|-------|----------|------|
| Generated code no tests | `sqlcgen/` | Expected - generated code |
| Temporary folder | `transport/internal_review_tmp/` | Cleanup candidate |

---

## 4. Testing Patterns Analysis

### 4.1 ✅ Good Patterns Found

```go
// Table-driven tests (seen throughout)
func TestUser_Validate(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        wantErr error
    }{
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}

// Mock interfaces (app layer)
type mockUserRepo struct {
    mock.Mock
}

// Test helpers (shared across tests)
func setupTestHandler(t *testing.T) *UserHandler {
    // ...
}
```

### 4.2 ⚠️ Inconsistencies Found

| Pattern | Current | Standard |
|---------|---------|----------|
| Mock location | Scattered in test files | Should be in `_mocks/` or `mocks.go` |
| Test file naming | `*_test.go` ✅ | Consistent |
| Helper location | Mixed (`helpers_test.go`, `auth_test_helper_test.go`) | Should consolidate |
| Integration tests | `*_integration_test.go` | Consider using build tags |

---

## 5. Rekomendasi Refactoring Testing

### 5.1 Immediate Actions

#### A. Fix Empty Test Helper
```bash
# Either delete or implement
rm internal/infra/postgres/test_helpers_test.go
# OR implement actual helpers
```

#### B. Add Tests for metrics Package
```go
// internal/shared/metrics/http_metrics_test.go
package metrics

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestHTTPMetrics_Interface(t *testing.T) {
    // Verify interface compliance
    var _ HTTPMetrics = (*mockHTTPMetrics)(nil)
}
```

### 5.2 Short-term Improvements

#### A. Split Large Test Files

**Before:**
```
handler/user_test.go (22KB, ~700 lines)
```

**After:**
```
handler/
├── user_test.go           # Core CRUD tests
├── user_validation_test.go # Validation edge cases
└── user_error_test.go     # Error handling scenarios
```

#### B. Consolidate Mocks

**Create shared mock package:**
```
internal/
└── testutil/
    └── mocks/
        ├── user_repo.go
        ├── audit_repo.go
        └── tx_manager.go
```

### 5.3 Long-term Standards

#### A. Use Build Tags for Integration Tests

```go
// +build integration

package postgres

func TestUserRepo_Integration(t *testing.T) {
    // Requires real DB
}
```

```makefile
## test-unit: Run unit tests only
test-unit:
    go test -v ./... -short

## test-integration: Run integration tests
test-integration:
    go test -v -tags=integration ./...
```

#### B. Add Test Coverage Visualization

```makefile
## coverage-html: Generate HTML coverage report
coverage-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    open coverage.html
```

---

## 6. Test Execution Commands

### Current Commands
```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run integration tests (requires DB)
make test-integration
```

### Recommended Additions
```bash
# Run tests for specific package
go test -v ./internal/app/user/...

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./internal/shared/redact/

# Generate coverage report
go test -coverprofile=cover.out ./...
go tool cover -func=cover.out
```

---

## 7. Testing Checklist for New Features

Untuk setiap fitur baru, pastikan:

- [ ] Unit tests untuk domain entities
- [ ] Unit tests untuk use cases dengan mocks
- [ ] Unit tests untuk handlers
- [ ] Unit tests untuk middleware if affected
- [ ] Integration tests untuk database operations
- [ ] Error case coverage
- [ ] Edge case coverage
- [ ] Coverage ≥80% untuk domain+app

---

## 8. Summary

### Strengths ✅
- 49 test files dengan coverage yang baik
- Table-driven tests konsisten
- Domain & app layer well-tested
- Integration tests tersedia
- Benchmark tests untuk performance-critical code

### Areas for Improvement ⚠️
- Beberapa test files terlalu besar (22-23KB)
- Mock definitions tersebar (tidak consolidated)
- Empty test helper file
- Missing tests untuk `shared/metrics/`
- Integration test setup bisa lebih streamlined

### Priority Actions
1. **HIGH:** Delete atau implement empty `test_helpers_test.go`
2. **MEDIUM:** Split large test files (`user_test.go`, `auth_test.go`)
3. **MEDIUM:** Add tests untuk `shared/metrics/`
4. **LOW:** Consolidate mocks ke shared package
5. **LOW:** Add build tags untuk integration tests

---

*Dokumentasi ini dihasilkan oleh BMad Method - Document Project Workflow*

---

## 9. Test Templates

Gunakan template berikut untuk menjaga konsistensi penulisan test baru.

### 9.1 Lokasi Template

Template tersedia di folder `docs/templates/`:
- **Unit Test:** `docs/templates/unit_test.go.tmpl`
- **Integration Test:** `docs/templates/integration_test.go.tmpl`

### 9.2 Cara Penggunaan

1. **Unit Test:**
   ```bash
   cp docs/templates/unit_test.go.tmpl internal/mypackage/mypackage_test.go
   # Edit package name dan function test
   ```

2. **Integration Test:**
   ```bash
   cp docs/templates/integration_test.go.tmpl internal/mypackage/mypackage_integration_test.go
   # File ini otomatis menggunakan build tag //go:build integration
   ```

### 9.3 Fitur Template

- **Standard Imports:** Termasuk `testify` dan `goleak`.
- **Leak Detection:** `TestMain` sudah terkonfigurasi dengan `goleak.VerifyTestMain`.
- **Table-Driven:** Struktur test menggunakan table-driven tests yang konsisten.
- **Container Support:** Integration template sudah termasuk setup `testcontainers`.

---

## 10. Synctest: Deterministic Time Testing (Go 1.25+)

> **Status:** ✅ Active - Go 1.25 Upgrade Completed
> **Current Go Version:** 1.25.5

Go 1.25 (released August 2025) memperkenalkan package `testing/synctest` untuk testing concurrent code dengan waktu deterministik. Project ini telah di-upgrade untuk menggunakan fitur ini.

Go 1.25 (released August 2025) memperkenalkan package `testing/synctest` untuk testing concurrent code dengan waktu deterministik.

### 10.1 Mengapa Synctest?

| Masalah | Solusi Synctest |
|---------|------------------|
| Flaky tests karena `time.Sleep` | Waktu virtual, instant advancement |
| Non-deterministic goroutine scheduling | Isolated "bubble" execution |
| Slow tests waiting for timeouts | Clock advances instantly |

### 10.2 Kapan Menggunakan Synctest

✅ **Gunakan untuk:**
- Timeout tests
- Timer/ticker behavior
- Context deadline tests
- Retry logic dengan backoff

❌ **Jangan gunakan untuk:**
- Database integration tests
- HTTP client/server tests
- Tests yang memerlukan real time

### 10.3 Contoh Penggunaan

```go
//go:build go1.25

package example

import (
    "context"
    "testing"
    "testing/synctest"
    "time"
)

func TestTimeoutBehavior(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        // Dalam bubble ini, time.Sleep tidak benar-benar menunggu
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        // Start operation
        done := make(chan struct{})
        go func() {
            time.Sleep(3 * time.Second) // Instant!
            close(done)
        }()

        // Wait for completion
        synctest.Wait() // Waits until all goroutines blocked

        select {
        case <-done:
            // Success - operation completed before timeout
        case <-ctx.Done():
            t.Fatal("unexpected timeout")
        }
    })
}
```

### 10.4 Kandidat Refactoring

Setelah upgrade ke Go 1.25, file berikut dapat direfactor:

| File | Reason |
|------|--------|
| `internal/transport/http/timeout_test.go` | HTTP timeout tests |
| `internal/transport/http/context_cancel_test.go` | Context cancellation |

### 10.5 Adoption Status

Project telah di-upgrade ke Go 1.25.5.
- `go.mod` updated.
- `internal/transport/http/synctest_example_test.go` added as reference.
- `internal/transport/http/timeout_refactored_test.go` implements deterministic timeout tests.

Developers encouraged to use `synctest` for new time-sensitive tests.

### 10.6 Referensi

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [testing/synctest Documentation](https://pkg.go.dev/testing/synctest)

