package containers

import (
	"testing"
	"time"
)

func TestNewPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	start := time.Now()
	pool := NewPostgres(t)
	elapsed := time.Since(start)

	// AC4: Container starts in ≤30 seconds
	if elapsed > 30*time.Second {
		t.Errorf("container took %v to start, expected ≤30s", elapsed)
	}

	// AC2: Pool should be usable
	var result int
	err := pool.QueryRow(t.Context(), "SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("failed to query database: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}

	t.Logf("Container started in %v", elapsed)
}
