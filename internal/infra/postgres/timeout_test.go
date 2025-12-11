package postgres

import (
	"context"
	"testing"
	"time"
)

func TestQueryContext_DefaultTimeout(t *testing.T) {
	ctx := context.Background()

	// With zero timeout, should use default (30s)
	newCtx, cancel := QueryContext(ctx, 0)
	defer cancel()

	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Fatal("Expected context to have deadline")
	}

	// Should be approximately 30 seconds from now
	remaining := time.Until(deadline)
	if remaining < 29*time.Second || remaining > 31*time.Second {
		t.Errorf("Expected ~30s timeout, got %v", remaining)
	}
}

func TestQueryContext_CustomTimeout(t *testing.T) {
	ctx := context.Background()
	customTimeout := 5 * time.Second

	newCtx, cancel := QueryContext(ctx, customTimeout)
	defer cancel()

	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Fatal("Expected context to have deadline")
	}

	// Should be approximately 5 seconds from now
	remaining := time.Until(deadline)
	if remaining < 4*time.Second || remaining > 6*time.Second {
		t.Errorf("Expected ~5s timeout, got %v", remaining)
	}
}

func TestQueryContext_CancelWorks(t *testing.T) {
	ctx := context.Background()

	newCtx, cancel := QueryContext(ctx, 10*time.Second)

	// Cancel should work
	cancel()

	select {
	case <-newCtx.Done():
		// Expected - context should be cancelled
	default:
		t.Error("Expected context to be cancelled")
	}
}

func TestDefaultTimeoutValues(t *testing.T) {
	// Verify constants are correct
	if DefaultConnTimeout != 10*time.Second {
		t.Errorf("DefaultConnTimeout = %v, want 10s", DefaultConnTimeout)
	}

	if DefaultQueryTimeout != 30*time.Second {
		t.Errorf("DefaultQueryTimeout = %v, want 30s", DefaultQueryTimeout)
	}
}
