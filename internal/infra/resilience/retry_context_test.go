package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Context cancellation and deadline tests.

func TestRetrier_Do_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	r := NewRetrier("test-cancel", cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var attempts int32
	fn := func(ctx context.Context) error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("always fail")
	}

	err := r.Do(ctx, fn)

	// Should stop before max attempts due to context cancellation
	assert.Less(t, int(atomic.LoadInt32(&attempts)), 10)
	assert.Error(t, err)
}
