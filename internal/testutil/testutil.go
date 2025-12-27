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

// RunWithGoleak is a stub for goleak integration (Story 1.3).
// For now, just run tests normally.
func RunWithGoleak(m *testing.M) int {
	// TODO: Story 1.3 will add goleak.VerifyTestMain(m)
	return m.Run()
}
