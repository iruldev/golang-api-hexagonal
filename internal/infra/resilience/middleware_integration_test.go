package resilience

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

// Full integration tests for resilience wrapper.

func TestResilienceWrapper_CircuitBreakerRejectsWhenOpen(t *testing.T) {
	t.Parallel()

	cb := &mockCircuitBreaker{
		name:  "test-cb",
		state: StateOpen,
		executeFn: func(ctx context.Context, fn func() (any, error)) (any, error) {
			return nil, NewCircuitOpenError(nil)
		},
	}

	wrapper := NewResilienceWrapper(
		WithCircuitBreakerFactory(func(name string) CircuitBreaker { return cb }),
	)

	err := wrapper.Execute(context.Background(), "test-op", func(ctx context.Context) error {
		t.Error("Operation should not have been called when circuit is open")
		return nil
	})

	if err == nil {
		t.Error("Expected error when circuit is open")
	}

	if !IsCircuitOpen(err) {
		t.Errorf("Expected circuit open error, got: %v", err)
	}
}

func TestResilienceWrapper_BulkheadRejectsWhenFull(t *testing.T) {
	t.Parallel()

	bulkhead := &mockBulkhead{
		name:          "test-bulkhead",
		activeCount:   10,
		waitingCount:  5,
		maxConcurrent: 10,
		doFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return NewBulkheadFullError(nil)
		},
	}

	wrapper := NewResilienceWrapper(
		WithWrapperBulkhead(bulkhead),
	)

	err := wrapper.Execute(context.Background(), "test-op", func(ctx context.Context) error {
		t.Error("Operation should not have been called when bulkhead is full")
		return nil
	})

	if err == nil {
		t.Error("Expected error when bulkhead is full")
	}

	if !IsBulkheadFull(err) {
		t.Errorf("Expected bulkhead full error, got: %v", err)
	}
}

func TestResilienceWrapper_ConcurrentExecution(t *testing.T) {
	t.Parallel()

	wrapper := NewResilienceWrapper()

	var count atomic.Int32
	var wg sync.WaitGroup

	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := wrapper.Execute(context.Background(), "concurrent-test", func(ctx context.Context) error {
				count.Add(1)
				return nil
			})
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
	}

	wg.Wait()

	if count.Load() != int32(numGoroutines) {
		t.Errorf("Expected %d operations, got %d", numGoroutines, count.Load())
	}
}
