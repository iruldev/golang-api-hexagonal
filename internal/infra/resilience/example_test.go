package resilience_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/resilience"
)

// This file provides executable examples for the resilience package.
// Examples are displayed in godoc and verified by go test.

// ExampleCircuitBreaker demonstrates how to use the circuit breaker pattern
// to protect against cascading failures.
func ExampleCircuitBreaker() {
	// Create a circuit breaker with default configuration
	cfg := resilience.DefaultCircuitBreakerConfig()
	cb := resilience.NewCircuitBreaker("example-service", cfg)

	// Execute an operation with circuit breaker protection
	result, err := cb.Execute(context.Background(), func() (any, error) {
		// This is your protected operation (database call, external API, etc.)
		return "success", nil
	})

	if err != nil {
		// Check if circuit is open
		if errors.Is(err, resilience.ErrCircuitOpen) {
			fmt.Println("Circuit is open, request rejected")
			return
		}
		fmt.Printf("Operation failed: %v\n", err)
		return
	}

	fmt.Printf("Result: %v, State: %s\n", result, cb.State())
	// Output: Result: success, State: closed
}

// ExampleRetrier demonstrates how to use retry with exponential backoff
// to handle transient failures.
func ExampleRetrier() {
	// Create a retrier with default configuration
	cfg := resilience.DefaultRetryConfig()
	cfg.MaxAttempts = 3 // Limit to 3 attempts for this example

	retrier := resilience.NewRetrier("example-retry", cfg,
		resilience.WithRetryLogger(slog.Default()),
	)

	attempt := 0

	// Execute an operation with retry logic
	err := retrier.Do(context.Background(), func(ctx context.Context) error {
		attempt++
		if attempt < 3 {
			// Simulate transient failure
			return errors.New("temporary error")
		}
		return nil // Success on third attempt
	})

	if err != nil {
		fmt.Printf("All retries failed: %v\n", err)
		return
	}

	fmt.Printf("Succeeded after %d attempts\n", attempt)
	// Output: Succeeded after 3 attempts
}

// ExampleResilienceWrapper demonstrates how to compose multiple
// resilience patterns using the ResilienceWrapper.
func ExampleResilienceWrapper() {
	// Create individual components
	cbCfg := resilience.DefaultCircuitBreakerConfig()
	retryCfg := resilience.DefaultRetryConfig()

	// Create the retrier
	retrier := resilience.NewRetrier("wrapper-retry", retryCfg)

	// Create a timeout using NewTimeout (name, duration)
	timeout := resilience.NewTimeout("wrapper-timeout", 5*time.Second)

	// Create a composed wrapper
	// The wrapper applies patterns in the correct order:
	// CircuitBreaker → Retry → Timeout (outermost to innermost)
	// Note: Timeout is applied per-attempt (innermost), not globally.
	wrapper := resilience.NewResilienceWrapper(
		resilience.WithCircuitBreakerFactory(func(name string) resilience.CircuitBreaker {
			return resilience.NewCircuitBreaker(name, cbCfg)
		}),
		resilience.WithWrapperRetrier(retrier),
		resilience.WithWrapperTimeout(timeout),
	)

	// Execute an operation with all resilience patterns
	err := wrapper.Execute(context.Background(), "my-operation", func(ctx context.Context) error {
		// Your protected operation
		return nil
	})

	if err != nil {
		fmt.Printf("Operation failed: %v\n", err)
		return
	}

	fmt.Println("Operation succeeded with full resilience protection")
	// Output: Operation succeeded with full resilience protection
}

// ExampleTimeout demonstrates how to use timeout wrapper
// to limit operation duration.
func ExampleTimeout() {
	// Create a timeout wrapper with 100ms timeout
	timeout := resilience.NewTimeout("example-timeout", 100*time.Millisecond)

	// Execute a fast operation (should succeed)
	err := timeout.Do(context.Background(), func(ctx context.Context) error {
		// Fast operation
		return nil
	})

	if err != nil {
		fmt.Printf("Operation timed out: %v\n", err)
		return
	}

	fmt.Println("Operation completed within timeout")
	// Output: Operation completed within timeout
}

// ExampleBulkhead demonstrates how to use the bulkhead pattern
// to limit concurrent operations.
func ExampleBulkhead() {
	// Create a bulkhead with max 2 concurrent operations
	cfg := resilience.BulkheadConfig{
		MaxConcurrent: 2,
		MaxWaiting:    5,
	}
	bh := resilience.NewBulkhead("example-bulkhead", cfg)

	// Execute operations - first 2 run concurrently
	err := bh.Do(context.Background(), func(ctx context.Context) error {
		// Your isolated operation
		return nil
	})

	if err != nil {
		if errors.Is(err, resilience.ErrBulkheadFull) {
			fmt.Println("Bulkhead full, request rejected")
			return
		}
		fmt.Printf("Operation failed: %v\n", err)
		return
	}

	fmt.Println("Operation completed within bulkhead")
	// Output: Operation completed within bulkhead
}
