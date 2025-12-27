package postgres

import (
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/testutil"
)

// TestMain runs all tests in the postgres package with goroutine leak detection.
// Any leaked goroutines after tests complete will cause test failure.
func TestMain(m *testing.M) {
	testutil.RunWithGoleak(m)
}
