// Package testutil provides shared test helpers and utilities.
//
// This package contains common testing utilities used across the codebase:
//   - TestContext: provides a context with test timeout
//   - RunWithGoleak: runs tests with goroutine leak detection
//
// Subpackages:
//   - assert: test assertion helpers using go-cmp
//   - containers: testcontainers helpers for integration tests
//   - fixtures: test data builders and factories
//   - mocks: generated mock implementations
package testutil

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// defaultTestTimeout defines the standard execution limit for tests.
const defaultTestTimeout = 30 * time.Second

// TestContext returns a context with a standard timeout.
// Automatically cancelled when the test completes.
func TestContext(t testing.TB) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	t.Cleanup(cancel)
	return ctx
}

// TestContextWithTimeout returns a context with custom timeout.
// Automatically cancelled when test completes.
func TestContextWithTimeout(t testing.TB, timeout time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

// RunWithGoleak runs tests with goroutine leak detection.
// Use this in TestMain for packages with integration tests.
//
// Example usage in test file:
//
//	func TestMain(m *testing.M) {
//	    testutil.RunWithGoleak(m)
//	}
func RunWithGoleak(m *testing.M) {
	// Ignore known background goroutines
	opts := []goleak.Option{
		goleak.IgnoreCurrent(),
		// pgx connection pool health check
		goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"),
		// otel telemetry exporters
		goleak.IgnoreTopFunction("go.opentelemetry.io/otel/sdk/trace.(*batchSpanProcessor).processQueue"),
	}

	goleak.VerifyTestMain(m, opts...)
}
