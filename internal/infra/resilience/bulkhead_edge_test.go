package resilience

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// Edge case tests for bulkhead error handling, panic recovery, and context cancellation.

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
