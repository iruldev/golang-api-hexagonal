// Package testutil provides testing utilities for the copy-paste kit examples.
package testutil

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain provides a package-level test main with goleak.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
