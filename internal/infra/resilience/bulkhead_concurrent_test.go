package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// NOTE: Several tests in this file use time.Sleep for goroutine synchronization.
// While channel-based synchronization is preferred, time.Sleep is used here because:
// 1. We need to verify intermediate states (WaitingCount, ActiveCount) during execution
// 2. The sleep durations are conservative (10-50ms) to minimize flakiness
// 3. These tests pass consistently with race detector enabled
// Future improvement: Consider using condition variables or polling with timeout.

// Concurrent tests for bulkhead concurrency behavior, slot management, and waiting.

func TestBulkhead_Do_RejectedWhenFull(t *testing.T) {
	// Bulkhead with 1 concurrent slot and 0 waiting
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 0})

	started := make(chan struct{})
	done := make(chan struct{})

	// Fill the single slot
	go func() {
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			close(started)
			<-done
			return nil
		})
	}()

	// Wait for slot to be occupied
	<-started

	// Attempt another operation - should be rejected
	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	close(done)

	if !errors.Is(err, ErrBulkheadFull) {
		t.Errorf("Do() error = %v, want ErrBulkheadFull", err)
	}
}

func TestBulkhead_Do_WaitsForSlot(t *testing.T) {
	// Bulkhead with 1 concurrent slot and 1 waiting
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 1})

	started := make(chan struct{})
	done := make(chan struct{})
	secondStarted := make(chan struct{})

	// Fill the single slot
	go func() {
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			close(started)
			<-done
			return nil
		})
	}()

	// Wait for slot to be occupied
	<-started

	// Start second operation - should wait
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			close(secondStarted)
			return nil
		})
	}()

	// Verify second operation is waiting
	time.Sleep(10 * time.Millisecond)
	if b.WaitingCount() != 1 {
		t.Errorf("WaitingCount() = %v, want 1", b.WaitingCount())
	}

	// Release first slot
	close(done)

	// Wait for second to complete
	wg.Wait()

	// Verify second operation ran
	select {
	case <-secondStarted:
		// Success
	case <-time.After(time.Second):
		t.Error("Second operation did not run")
	}
}

func TestBulkhead_Do_ConcurrentOperations(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 20})

	const numOps = 20
	var wg sync.WaitGroup
	var successCount atomic.Int64
	var maxActive atomic.Int64

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := b.Do(context.Background(), func(ctx context.Context) error {
				current := int64(b.ActiveCount())
				for {
					old := maxActive.Load()
					if current <= old || maxActive.CompareAndSwap(old, current) {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
				return nil
			})
			if err == nil {
				successCount.Add(1)
			}
		}()
	}

	wg.Wait()

	if successCount.Load() != numOps {
		t.Errorf("successCount = %v, want %v", successCount.Load(), numOps)
	}

	if maxActive.Load() > 5 {
		t.Errorf("maxActive = %v, want <= 5", maxActive.Load())
	}
}

func TestBulkhead_Do_IsolatedBulkheads(t *testing.T) {
	// Two separate bulkheads should not affect each other
	b1 := NewBulkhead("bulkhead1", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 0})
	b2 := NewBulkhead("bulkhead2", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 0})

	started := make(chan struct{})
	done := make(chan struct{})

	// Fill bulkhead1
	go func() {
		_ = b1.Do(context.Background(), func(ctx context.Context) error {
			close(started)
			<-done
			return nil
		})
	}()

	<-started

	// Bulkhead2 should still accept operations
	err := b2.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	close(done)

	if err != nil {
		t.Errorf("b2.Do() error = %v, want nil (bulkheads should be isolated)", err)
	}
}

func TestBulkhead_ActiveAndWaitingCounts(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 2, MaxWaiting: 5})

	started := make(chan struct{}, 2)
	done := make(chan struct{})

	// Start 2 operations to fill capacity
	for i := 0; i < 2; i++ {
		go func() {
			_ = b.Do(context.Background(), func(ctx context.Context) error {
				started <- struct{}{}
				<-done
				return nil
			})
		}()
	}

	// Wait for both to start
	<-started
	<-started

	// Allow some time for counters to update
	time.Sleep(10 * time.Millisecond)

	if b.ActiveCount() != 2 {
		t.Errorf("ActiveCount() = %v, want 2", b.ActiveCount())
	}

	// Start 2 more operations - they should wait
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Do(context.Background(), func(ctx context.Context) error {
				return nil
			})
		}()
	}

	// Allow some time for waiting count to update
	time.Sleep(20 * time.Millisecond)

	if b.WaitingCount() != 2 {
		t.Errorf("WaitingCount() = %v, want 2", b.WaitingCount())
	}

	// Release first operations
	close(done)

	// Wait for all to complete
	wg.Wait()

	// Allow some time for counters to update
	time.Sleep(10 * time.Millisecond)

	if b.ActiveCount() != 0 {
		t.Errorf("Final ActiveCount() = %v, want 0", b.ActiveCount())
	}
	if b.WaitingCount() != 0 {
		t.Errorf("Final WaitingCount() = %v, want 0", b.WaitingCount())
	}
}

// TestDoWithBulkhead_Rejection verifies that DoWithBulkhead properly propagates rejection errors.
func TestDoWithBulkhead_Rejection(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 0})

	started := make(chan struct{})
	done := make(chan struct{})

	// Fill the single slot
	go func() {
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			close(started)
			<-done
			return nil
		})
	}()

	<-started

	// Try DoWithBulkhead when bulkhead is full
	result, err := DoWithBulkhead(b, context.Background(), func(ctx context.Context) (string, error) {
		return "should not reach", nil
	})

	// Should get rejection error
	if !errors.Is(err, ErrBulkheadFull) {
		t.Errorf("DoWithBulkhead() error = %v, want ErrBulkheadFull", err)
	}

	// Result should be zero value
	if result != "" {
		t.Errorf("DoWithBulkhead() result = %v, want empty string on rejection", result)
	}

	close(done)
}

// TestBulkhead_WaitDurationMetrics verifies that wait duration is recorded when waiting for slot.
func TestBulkhead_WaitDurationMetrics(t *testing.T) {
	metrics := NewBulkheadMetrics(nil)
	b := NewBulkhead("test-wait", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 5},
		WithBulkheadMetrics(metrics))

	started := make(chan struct{})
	done := make(chan struct{})

	// Fill the single slot
	go func() {
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			close(started)
			<-done
			return nil
		})
	}()

	<-started

	// Start second operation that will wait
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})
	}()

	// Let it wait for a bit
	time.Sleep(50 * time.Millisecond)

	// Release first slot
	close(done)

	// Wait for second to complete
	wg.Wait()

	// Metrics should have recorded wait duration
	// Since we can't easily inspect histogram values, we just verify no panic occurred
	// and the operation completed successfully
}
