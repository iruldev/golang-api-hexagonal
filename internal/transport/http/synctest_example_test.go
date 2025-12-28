//go:build go1.25

package http

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

// TestTimeoutWithSynctest demonstrates deterministic timeout testing using Go 1.25's synctest package.
// This test completes instantly because synctest uses a virtual clock.
func TestTimeoutWithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create a context with 5 second timeout
		// In synctest bubble, this uses virtual time
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Simulate an operation that takes 3 seconds
		done := make(chan struct{})
		go func() {
			// This sleep advances virtual clock instantly!
			time.Sleep(3 * time.Second)
			close(done)
		}()

		// Wait for all goroutines in the bubble to be blocked
		synctest.Wait()

		// Verify operation completed before timeout
		select {
		case <-done:
			// Success - operation completed
		case <-ctx.Done():
			t.Fatal("unexpected timeout - operation should complete in 3s")
		}
	})
}

// TestContextDeadlineWithSynctest tests context deadline behavior deterministically.
func TestContextDeadlineWithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create context with 1 second deadline
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Track if operation was cancelled
		var cancelled bool
		done := make(chan struct{})

		go func() {
			defer close(done)
			select {
			case <-time.After(2 * time.Second): // Would exceed deadline
				// Operation completed (shouldn't happen)
			case <-ctx.Done():
				cancelled = true
			}
		}()

		// Wait for goroutine to block, then let time advance
		synctest.Wait()

		// Goroutine should have received context cancellation
		<-done

		if !cancelled {
			t.Error("expected operation to be cancelled by context deadline")
		}
	})
}

// TestRetryWithBackoffSynctest demonstrates testing retry logic with exponential backoff.
func TestRetryWithBackoffSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		attempts := 0
		maxAttempts := 3
		backoff := 100 * time.Millisecond

		for attempts < maxAttempts {
			attempts++

			// Simulate operation that fails on first 2 attempts
			if attempts < maxAttempts {
				// Wait with exponential backoff
				time.Sleep(backoff * time.Duration(attempts))
				continue
			}

			// Success on final attempt
			break
		}

		if attempts != maxAttempts {
			t.Errorf("expected %d attempts, got %d", maxAttempts, attempts)
		}
	})
}
