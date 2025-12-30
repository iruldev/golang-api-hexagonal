package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockCircuitBreaker implements CircuitBreaker for testing
type mockCircuitBreaker struct {
	name        string
	state       State
	executeFn   func(ctx context.Context, fn func() (any, error)) (any, error)
	execCalled  int
	stateChecks int
}

func (m *mockCircuitBreaker) Execute(ctx context.Context, fn func() (any, error)) (any, error) {
	m.execCalled++
	if m.executeFn != nil {
		return m.executeFn(ctx, fn)
	}
	return fn()
}

func (m *mockCircuitBreaker) State() State {
	m.stateChecks++
	return m.state
}

func (m *mockCircuitBreaker) Name() string {
	return m.name
}

// mockRetrier implements Retrier for testing
type mockRetrier struct {
	name     string
	doFn     func(ctx context.Context, fn func(ctx context.Context) error) error
	doCalled int
}

func (m *mockRetrier) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	m.doCalled++
	if m.doFn != nil {
		return m.doFn(ctx, fn)
	}
	return fn(ctx)
}

func (m *mockRetrier) Name() string {
	return m.name
}

// mockTimeout implements Timeout for testing
type mockTimeout struct {
	name     string
	duration time.Duration
	doFn     func(ctx context.Context, fn func(ctx context.Context) error) error
	doCalled int
}

func (m *mockTimeout) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	m.doCalled++
	if m.doFn != nil {
		return m.doFn(ctx, fn)
	}
	return fn(ctx)
}

func (m *mockTimeout) Name() string {
	return m.name
}

func (m *mockTimeout) Duration() time.Duration {
	return m.duration
}

// mockBulkhead implements Bulkhead for testing
type mockBulkhead struct {
	name          string
	activeCount   int
	waitingCount  int
	doFn          func(ctx context.Context, fn func(ctx context.Context) error) error
	doCalled      int
	maxConcurrent int
}

func (m *mockBulkhead) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	m.doCalled++
	if m.doFn != nil {
		return m.doFn(ctx, fn)
	}
	return fn(ctx)
}

func (m *mockBulkhead) Name() string {
	return m.name
}

func (m *mockBulkhead) ActiveCount() int {
	return m.activeCount
}

func (m *mockBulkhead) WaitingCount() int {
	return m.waitingCount
}

func TestResilienceWrapper_Execute_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func() (*mockCircuitBreaker, *mockRetrier, *mockTimeout, *mockBulkhead)
		operation  func(ctx context.Context) error
		wantErr    bool
	}{
		{
			name: "successful operation passes through all layers",
			setupMocks: func() (*mockCircuitBreaker, *mockRetrier, *mockTimeout, *mockBulkhead) {
				cb := &mockCircuitBreaker{name: "test-cb", state: StateClosed}
				retrier := &mockRetrier{name: "test-retrier"}
				timeout := &mockTimeout{name: "test-timeout", duration: 5 * time.Second}
				bulkhead := &mockBulkhead{name: "test-bulkhead"}
				return cb, retrier, timeout, bulkhead
			},
			operation: func(ctx context.Context) error {
				return nil // Success
			},
			wantErr: false,
		},
		{
			name: "operation with no components configured",
			setupMocks: func() (*mockCircuitBreaker, *mockRetrier, *mockTimeout, *mockBulkhead) {
				return nil, nil, nil, nil
			},
			operation: func(ctx context.Context) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "operation error propagates",
			setupMocks: func() (*mockCircuitBreaker, *mockRetrier, *mockTimeout, *mockBulkhead) {
				return &mockCircuitBreaker{name: "cb", state: StateClosed}, nil, nil, nil
			},
			operation: func(ctx context.Context) error {
				return errors.New("operation failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cb, retrier, timeout, bulkhead := tt.setupMocks()

			opts := []ResilienceWrapperOption{}
			if cb != nil {
				opts = append(opts, WithCircuitBreakerFactory(func(name string) CircuitBreaker {
					return cb
				}))
			}
			if retrier != nil {
				opts = append(opts, WithWrapperRetrier(retrier))
			}
			if timeout != nil {
				opts = append(opts, WithWrapperTimeout(timeout))
			}
			if bulkhead != nil {
				opts = append(opts, WithWrapperBulkhead(bulkhead))
			}

			wrapper := NewResilienceWrapper(opts...)
			err := wrapper.Execute(context.Background(), "test-op", tt.operation)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResilienceWrapper_CompositionOrder(t *testing.T) {
	t.Parallel()

	// This test verifies that the composition order is:
	// Bulkhead → CircuitBreaker → Retry → Timeout (outermost to innermost)
	// Execution order: Bulkhead first, then CB, then Retry, then Timeout, then fn

	var callOrder []string
	var mu sync.Mutex

	recordCall := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		callOrder = append(callOrder, name)
	}

	bulkhead := &mockBulkhead{
		name: "test-bulkhead",
		doFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			recordCall("bulkhead-start")
			err := fn(ctx)
			recordCall("bulkhead-end")
			return err
		},
	}

	cb := &mockCircuitBreaker{
		name:  "test-cb",
		state: StateClosed,
		executeFn: func(ctx context.Context, fn func() (any, error)) (any, error) {
			recordCall("cb-start")
			result, err := fn()
			recordCall("cb-end")
			return result, err
		},
	}

	retrier := &mockRetrier{
		name: "test-retrier",
		doFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			recordCall("retry-start")
			err := fn(ctx)
			recordCall("retry-end")
			return err
		},
	}

	timeout := &mockTimeout{
		name:     "test-timeout",
		duration: 5 * time.Second,
		doFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			recordCall("timeout-start")
			err := fn(ctx)
			recordCall("timeout-end")
			return err
		},
	}

	wrapper := NewResilienceWrapper(
		WithWrapperBulkhead(bulkhead),
		WithCircuitBreakerFactory(func(name string) CircuitBreaker { return cb }),
		WithWrapperRetrier(retrier),
		WithWrapperTimeout(timeout),
	)

	err := wrapper.Execute(context.Background(), "test-op", func(ctx context.Context) error {
		recordCall("operation")
		return nil
	})

	if err != nil {
		t.Errorf("Execute() unexpected error: %v", err)
	}

	// Expected order: bulkhead → cb → retry → timeout → operation → (reverse)
	expected := []string{
		"bulkhead-start",
		"cb-start",
		"retry-start",
		"timeout-start",
		"operation",
		"timeout-end",
		"retry-end",
		"cb-end",
		"bulkhead-end",
	}

	if len(callOrder) != len(expected) {
		t.Errorf("Call order length mismatch: got %v, want %v", callOrder, expected)
	}

	for i, call := range expected {
		if i >= len(callOrder) || callOrder[i] != call {
			t.Errorf("Call order mismatch at index %d: got %v, want %v", i, callOrder, expected)
			break
		}
	}
}

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

func TestResilienceWrapper_TimeoutTriggersBeforeRetryExhaustion(t *testing.T) {
	t.Parallel()

	timeout := &mockTimeout{
		name:     "test-timeout",
		duration: 100 * time.Millisecond,
		doFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return NewTimeoutExceededError(context.DeadlineExceeded)
		},
	}

	retryCount := 0
	retrier := &mockRetrier{
		name: "test-retrier",
		doFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			// Pass through to timeout
			retryCount++
			return fn(ctx)
		},
	}

	wrapper := NewResilienceWrapper(
		WithWrapperRetrier(retrier),
		WithWrapperTimeout(timeout),
	)

	err := wrapper.Execute(context.Background(), "test-op", func(ctx context.Context) error {
		t.Error("Operation should not have been called when timeout triggers")
		return nil
	})

	if err == nil {
		t.Error("Expected error when timeout triggers")
	}

	if !IsTimeoutExceeded(err) {
		t.Errorf("Expected timeout exceeded error, got: %v", err)
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

func TestResilienceWrapper_AllComponentsMockable(t *testing.T) {
	t.Parallel()

	// Verify all interfaces can be mocked
	cb := &mockCircuitBreaker{name: "mock-cb", state: StateClosed}
	retrier := &mockRetrier{name: "mock-retrier"}
	timeout := &mockTimeout{name: "mock-timeout", duration: time.Second}
	bulkhead := &mockBulkhead{name: "mock-bulkhead"}

	wrapper := NewResilienceWrapper(
		WithCircuitBreakerFactory(func(name string) CircuitBreaker { return cb }),
		WithWrapperRetrier(retrier),
		WithWrapperTimeout(timeout),
		WithWrapperBulkhead(bulkhead),
	)

	err := wrapper.Execute(context.Background(), "test", func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify all mocks were called
	if cb.execCalled != 1 {
		t.Errorf("CircuitBreaker.Execute called %d times, want 1", cb.execCalled)
	}
	if retrier.doCalled != 1 {
		t.Errorf("Retrier.Do called %d times, want 1", retrier.doCalled)
	}
	if timeout.doCalled != 1 {
		t.Errorf("Timeout.Do called %d times, want 1", timeout.doCalled)
	}
	if bulkhead.doCalled != 1 {
		t.Errorf("Bulkhead.Do called %d times, want 1", bulkhead.doCalled)
	}
}

func TestCircuitBreakerFactory(t *testing.T) {
	t.Parallel()

	cfg := DefaultCircuitBreakerConfig()
	factory := NewCircuitBreakerFactory(cfg)

	// Get two circuit breakers with different names
	cb1 := factory("operation-1")
	cb2 := factory("operation-2")

	// Same name should return the same instance
	cb1Again := factory("operation-1")

	if cb1.Name() != "operation-1" {
		t.Errorf("Expected name 'operation-1', got %s", cb1.Name())
	}

	if cb2.Name() != "operation-2" {
		t.Errorf("Expected name 'operation-2', got %s", cb2.Name())
	}

	// Verify caching works (same pointer)
	if cb1 != cb1Again {
		t.Error("Expected factory to return cached instance for same name")
	}

	// Different names should have different instances
	if cb1 == cb2 {
		t.Error("Expected different instances for different names")
	}
}

func TestCircuitBreakerPresets(t *testing.T) {
	t.Parallel()

	cfg := DefaultCircuitBreakerConfig()
	presets := NewCircuitBreakerPresets(cfg)

	tests := []struct {
		name     string
		getCB    func() CircuitBreaker
		wantName string
	}{
		{
			name:     "database preset",
			getCB:    presets.ForDatabase,
			wantName: "database",
		},
		{
			name:     "external_api preset",
			getCB:    presets.ForExternalAPI,
			wantName: "external_api",
		},
		{
			name:     "default preset",
			getCB:    presets.Default,
			wantName: "default",
		},
		{
			name:     "custom operation",
			getCB:    func() CircuitBreaker { return presets.ForOperation("custom-op") },
			wantName: "custom-op",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := tt.getCB()
			if cb.Name() != tt.wantName {
				t.Errorf("Expected name '%s', got '%s'", tt.wantName, cb.Name())
			}

			// Verify initial state is closed
			if cb.State() != StateClosed {
				t.Errorf("Expected initial state closed, got %s", cb.State())
			}
		})
	}

	// Verify factory is exposed
	factory := presets.Factory()
	if factory == nil {
		t.Error("Factory() should not return nil")
	}

	cb := factory("via-factory")
	if cb.Name() != "via-factory" {
		t.Errorf("Expected name 'via-factory', got %s", cb.Name())
	}
}

func TestResilienceWrapper_NilOptions(t *testing.T) {
	t.Parallel()

	// Test that nil options don't cause panics
	wrapper := NewResilienceWrapper(
		WithCircuitBreakerFactory(nil),
		WithWrapperRetrier(nil),
		WithWrapperTimeout(nil),
		WithWrapperBulkhead(nil),
		WithWrapperTracer(nil),
		WithWrapperLogger(nil),
	)

	err := wrapper.Execute(context.Background(), "test", func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error with nil options: %v", err)
	}
}

func TestResilienceWrapper_ContextPropagation(t *testing.T) {
	t.Parallel()

	type ctxKey string
	key := ctxKey("test-key")

	wrapper := NewResilienceWrapper()

	ctx := context.WithValue(context.Background(), key, "test-value")

	err := wrapper.Execute(ctx, "test", func(ctx context.Context) error {
		val := ctx.Value(key)
		if val != "test-value" {
			t.Error("Context value not propagated")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestErrorType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantType string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantType: "none",
		},
		{
			name:     "circuit open error",
			err:      NewCircuitOpenError(nil),
			wantType: "circuit_open",
		},
		{
			name:     "bulkhead full error",
			err:      NewBulkheadFullError(nil),
			wantType: "bulkhead_full",
		},
		{
			name:     "timeout exceeded error",
			err:      NewTimeoutExceededError(nil),
			wantType: "timeout",
		},
		{
			name:     "max retries exceeded error",
			err:      NewMaxRetriesExceededError(nil),
			wantType: "max_retries",
		},
		{
			name:     "unknown error",
			err:      errors.New("some error"),
			wantType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorType(tt.err)
			if got != tt.wantType {
				t.Errorf("errorType(%v) = %s, want %s", tt.err, got, tt.wantType)
			}
		})
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
