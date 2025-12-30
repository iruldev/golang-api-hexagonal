package resilience

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
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

func TestNewBulkhead(t *testing.T) {
	// Note: MaxConcurrent and MaxWaiting behavior is verified through
	// behavioral tests (TestBulkhead_Do_RejectedWhenFull, TestBulkhead_Do_WaitsForSlot, etc.)
	// This test focuses on basic construction and initial state.
	tests := []struct {
		name  string
		bName string
		cfg   BulkheadConfig
	}{
		{
			name:  "creates bulkhead with config values",
			bName: "test-bulkhead",
			cfg:   BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
		},
		{
			name:  "creates bulkhead with default-like values",
			bName: "default-bulkhead",
			cfg:   BulkheadConfig{MaxConcurrent: 10, MaxWaiting: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBulkhead(tt.bName, tt.cfg)
			if b.Name() != tt.bName {
				t.Errorf("Name() = %v, want %v", b.Name(), tt.bName)
			}
			if b.ActiveCount() != 0 {
				t.Errorf("ActiveCount() = %v, want 0", b.ActiveCount())
			}
			if b.WaitingCount() != 0 {
				t.Errorf("WaitingCount() = %v, want 0", b.WaitingCount())
			}
		})
	}
}

func TestBulkhead_Do_Success(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10})

	called := false
	err := b.Do(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if !called {
		t.Error("Do() did not call the function")
	}
}

func TestBulkhead_Do_Error(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10})
	expectedErr := errors.New("test error")

	err := b.Do(context.Background(), func(ctx context.Context) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("Do() error = %v, want %v", err, expectedErr)
	}
}

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

func TestBulkhead_Do_ContextCancellation(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 5})

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

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start waiting operation
	var wg sync.WaitGroup
	var opErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		opErr = b.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}()

	// Verify operation is waiting
	time.Sleep(10 * time.Millisecond)
	if b.WaitingCount() != 1 {
		t.Errorf("WaitingCount() = %v, want 1", b.WaitingCount())
	}

	// Cancel context
	cancel()

	// Wait for operation to return
	wg.Wait()
	close(done)

	if !errors.Is(opErr, context.Canceled) {
		t.Errorf("Do() error = %v, want context.Canceled", opErr)
	}
}

func TestBulkhead_Do_SlotReleasedOnError(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 0})

	// First operation fails
	_ = b.Do(context.Background(), func(ctx context.Context) error {
		return errors.New("test error")
	})

	// Slot should be released, so second operation should succeed
	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Second Do() error = %v, want nil (slot should be released)", err)
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

func TestBulkhead_WithOptions(t *testing.T) {
	metrics := NoopBulkheadMetrics()

	t.Run("WithBulkheadMetrics", func(t *testing.T) {
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
			WithBulkheadMetrics(metrics))

		err := b.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("Do() error = %v, want nil", err)
		}
	})

	t.Run("WithBulkheadLogger", func(t *testing.T) {
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
			WithBulkheadLogger(nil)) // Should use default

		err := b.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("Do() error = %v, want nil", err)
		}
	})

	t.Run("WithNilMetrics", func(t *testing.T) {
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
			WithBulkheadMetrics(nil)) // Should be noop

		err := b.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("Do() error = %v, want nil", err)
		}
	})
}

func TestBulkheadPresets(t *testing.T) {
	cfg := BulkheadConfig{MaxConcurrent: 10, MaxWaiting: 100}
	presets := NewBulkheadPresets(cfg)

	t.Run("ForDatabase returns bulkhead", func(t *testing.T) {
		b := presets.ForDatabase()
		if b.Name() != "database" {
			t.Errorf("ForDatabase().Name() = %v, want database", b.Name())
		}
	})

	t.Run("ForExternalAPI returns bulkhead", func(t *testing.T) {
		b := presets.ForExternalAPI()
		if b.Name() != "external_api" {
			t.Errorf("ForExternalAPI().Name() = %v, want external_api", b.Name())
		}
	})

	t.Run("Default returns bulkhead", func(t *testing.T) {
		b := presets.Default()
		if b.Name() != "default" {
			t.Errorf("Default().Name() = %v, want default", b.Name())
		}
	})

	t.Run("ForOperation creates custom bulkhead", func(t *testing.T) {
		b := presets.ForOperation("custom", 3, 5)
		if b.Name() != "custom" {
			t.Errorf("ForOperation().Name() = %v, want custom", b.Name())
		}
	})
}

func TestBulkheadPresets_WithOptions(t *testing.T) {
	cfg := BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10}
	metrics := NoopBulkheadMetrics()
	presets := NewBulkheadPresets(cfg, WithBulkheadMetrics(metrics))

	// All preset bulkheads should have metrics
	err := presets.ForDatabase().Do(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("ForDatabase().Do() error = %v, want nil", err)
	}

	err = presets.ForExternalAPI().Do(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("ForExternalAPI().Do() error = %v, want nil", err)
	}
}

func TestDoWithBulkhead(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10})

	t.Run("returns result on success", func(t *testing.T) {
		result, err := DoWithBulkhead(b, context.Background(), func(ctx context.Context) (string, error) {
			return "hello", nil
		})

		if err != nil {
			t.Errorf("DoWithBulkhead() error = %v, want nil", err)
		}
		if result != "hello" {
			t.Errorf("DoWithBulkhead() result = %v, want hello", result)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		expectedErr := errors.New("test error")
		result, err := DoWithBulkhead(b, context.Background(), func(ctx context.Context) (string, error) {
			return "", expectedErr
		})

		if !errors.Is(err, expectedErr) {
			t.Errorf("DoWithBulkhead() error = %v, want %v", err, expectedErr)
		}
		if result != "" {
			t.Errorf("DoWithBulkhead() result = %v, want empty string", result)
		}
	})
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

func TestBulkheadMetrics_Operations(t *testing.T) {
	metrics := NewBulkheadMetrics(nil)

	t.Run("SetActive", func(t *testing.T) {
		metrics.SetActive("test", 5)
		// No panic = success (gauge is internal)
	})

	t.Run("SetWaiting", func(t *testing.T) {
		metrics.SetWaiting("test", 3)
		// No panic = success
	})

	t.Run("RecordOperation", func(t *testing.T) {
		metrics.RecordOperation("test", "success")
		metrics.RecordOperation("test", "rejected")
		metrics.RecordOperation("test", "error")
		// No panic = success
	})

	t.Run("RecordWaitDuration", func(t *testing.T) {
		metrics.RecordWaitDuration("test", 0.05)
		// No panic = success
	})

	t.Run("Reset", func(t *testing.T) {
		metrics.Reset()
		// No panic = success
	})
}

func TestNoopBulkheadMetrics(t *testing.T) {
	metrics := NoopBulkheadMetrics()
	if metrics == nil {
		t.Error("NoopBulkheadMetrics() returned nil")
	}
}

// TestBulkhead_Do_PanicReleasesSlot verifies that if the wrapped function panics,
// the bulkhead slot is still released (Critical Warning #6).
func TestBulkhead_Do_PanicReleasesSlot(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 0})

	// Execute function that panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic but none occurred")
			}
		}()

		_ = b.Do(context.Background(), func(ctx context.Context) error {
			panic("test panic")
		})
	}()

	// Slot should be released - verify second operation works
	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Slot not released after panic: %v", err)
	}

	// Verify counters are correct
	if b.ActiveCount() != 0 {
		t.Errorf("ActiveCount() = %v, want 0 after panic", b.ActiveCount())
	}
}

// TestBulkhead_WithNonNilLogger verifies that a custom non-nil logger is used.
func TestBulkhead_WithNonNilLogger(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	customLogger := slog.New(handler)

	b := NewBulkhead("test-logger", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
		WithBulkheadLogger(customLogger))

	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}

	// Verify custom logger was used (should have log output)
	if buf.Len() == 0 {
		t.Error("Custom logger was not used - no log output")
	}

	if !bytes.Contains(buf.Bytes(), []byte("test-logger")) {
		t.Error("Log output doesn't contain bulkhead name")
	}
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

// TestNewBulkhead_InvalidConfig verifies that NewBulkhead panics on invalid configuration.
func TestNewBulkhead_InvalidConfig(t *testing.T) {
	t.Run("MaxConcurrent zero panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for MaxConcurrent=0 but none occurred")
			}
		}()
		NewBulkhead("test", BulkheadConfig{MaxConcurrent: 0, MaxWaiting: 10})
	})

	t.Run("MaxConcurrent negative panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for MaxConcurrent=-1 but none occurred")
			}
		}()
		NewBulkhead("test", BulkheadConfig{MaxConcurrent: -1, MaxWaiting: 10})
	})

	t.Run("MaxWaiting negative panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for MaxWaiting=-1 but none occurred")
			}
		}()
		NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: -1})
	})

	t.Run("MaxWaiting zero is valid", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic for MaxWaiting=0: %v", r)
			}
		}()
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 0})
		if b == nil {
			t.Error("Expected valid bulkhead, got nil")
		}
	})
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

// TestBulkhead_Do_ContextCancellation_WithLogger verifies that cancelled operations are logged.
func TestBulkhead_Do_ContextCancellation_WithLogger(t *testing.T) {
	var buf bytes.Buffer
	var mu sync.Mutex
	handler := slog.NewTextHandler(&lockedWriter{w: &buf, mu: &mu}, &slog.HandlerOptions{Level: slog.LevelDebug})
	customLogger := slog.New(handler)

	b := NewBulkhead("test-cancel", BulkheadConfig{MaxConcurrent: 1, MaxWaiting: 5},
		WithBulkheadLogger(customLogger))

	started := make(chan struct{})
	done := make(chan struct{})

	// Fill the single slot
	var wgFirst sync.WaitGroup
	wgFirst.Add(1)
	go func() {
		defer wgFirst.Done()
		_ = b.Do(context.Background(), func(ctx context.Context) error {
			close(started)
			<-done
			return nil
		})
	}()

	<-started

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start waiting operation
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = b.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}()

	// Wait for operation to enter waiting state
	time.Sleep(10 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for cancelled operation to return
	wg.Wait()

	// Release first operation and wait for it to complete
	close(done)
	wgFirst.Wait()

	// Verify cancelled operation was logged (safe to read now)
	mu.Lock()
	logOutput := buf.String()
	mu.Unlock()

	if !bytes.Contains([]byte(logOutput), []byte("cancelled")) {
		t.Error("Cancelled operation was not logged")
	}
	if !bytes.Contains([]byte(logOutput), []byte("test-cancel")) {
		t.Error("Log output doesn't contain bulkhead name")
	}
}

// lockedWriter wraps an io.Writer with a mutex for thread-safe writes.
type lockedWriter struct {
	w  *bytes.Buffer
	mu *sync.Mutex
}

func (lw *lockedWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}

// TestBulkheadPresets_ForOperation_OptionsPropagation verifies that preset options
// are propagated to bulkheads created via ForOperation.
func TestBulkheadPresets_ForOperation_OptionsPropagation(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	customLogger := slog.New(handler)

	cfg := BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10}
	presets := NewBulkheadPresets(cfg, WithBulkheadLogger(customLogger))

	// Create a custom bulkhead via ForOperation - should inherit the logger option
	b := presets.ForOperation("custom-op", 3, 5)

	// Execute an operation to trigger logging
	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}

	// Verify custom logger was used (preset option was propagated)
	if buf.Len() == 0 {
		t.Error("Preset logger option was not propagated to ForOperation bulkhead - no log output")
	}

	if !bytes.Contains(buf.Bytes(), []byte("custom-op")) {
		t.Error("Log output doesn't contain ForOperation bulkhead name")
	}
}

// TestBulkheadPresets_ForOperation_AdditionalOptions verifies that additional options
// passed to ForOperation override preset options.
func TestBulkheadPresets_ForOperation_AdditionalOptions(t *testing.T) {
	var presetBuf bytes.Buffer
	presetHandler := slog.NewTextHandler(&presetBuf, &slog.HandlerOptions{Level: slog.LevelDebug})
	presetLogger := slog.New(presetHandler)

	var customBuf bytes.Buffer
	customHandler := slog.NewTextHandler(&customBuf, &slog.HandlerOptions{Level: slog.LevelDebug})
	customLogger := slog.New(customHandler)

	cfg := BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10}
	presets := NewBulkheadPresets(cfg, WithBulkheadLogger(presetLogger))

	// Create a custom bulkhead with its own logger - should override preset logger
	b := presets.ForOperation("override-test", 3, 5, WithBulkheadLogger(customLogger))

	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}

	// Additional logger should be used, not preset logger
	if customBuf.Len() == 0 {
		t.Error("Additional logger option was not applied - no log output in custom buffer")
	}

	// Preset logger should NOT have output for this bulkhead
	if bytes.Contains(presetBuf.Bytes(), []byte("override-test")) {
		t.Error("Preset logger should have been overridden by additional option")
	}
}
