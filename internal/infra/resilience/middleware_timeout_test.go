package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Timeout integration and error type tests.

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
