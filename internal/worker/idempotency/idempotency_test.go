package idempotency

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockStore is a test implementation of the Store interface.
type MockStore struct {
	mu        sync.Mutex
	keys      map[string]time.Time
	results   map[string][]byte
	checkErr  error
	failMode  FailMode
	callCount int
}

func NewMockStore() *MockStore {
	return &MockStore{
		keys:    make(map[string]time.Time),
		results: make(map[string][]byte),
	}
}

func (m *MockStore) Check(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.checkErr != nil {
		if m.failMode == FailOpen {
			return true, nil
		}
		return false, m.checkErr
	}

	if key == "" {
		return true, nil
	}

	expiry, exists := m.keys[key]
	if exists && time.Now().Before(expiry) {
		return false, nil // Duplicate
	}

	m.keys[key] = time.Now().Add(ttl)
	return true, nil // New
}

func (m *MockStore) StoreResult(ctx context.Context, key string, result []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results["result:"+key] = result
	return nil
}

func (m *MockStore) GetResult(ctx context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result, ok := m.results["result:"+key]
	return result, ok, nil
}

func (m *MockStore) SetError(err error, mode FailMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkErr = err
	m.failMode = mode
}

func (m *MockStore) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// === FailMode Tests ===

func TestFailMode_String(t *testing.T) {
	tests := []struct {
		name     string
		mode     FailMode
		expected string
	}{
		{"FailOpen", FailOpen, "fail-open"},
		{"FailClosed", FailClosed, "fail-closed"},
		{"Unknown", FailMode(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

// === MockStore Tests ===

func TestMockStore_Check_FirstOccurrence(t *testing.T) {
	// Arrange
	store := NewMockStore()
	ctx := context.Background()

	// Act
	isNew, err := store.Check(ctx, "test-key", time.Hour)

	// Assert
	require.NoError(t, err)
	assert.True(t, isNew, "First occurrence should return true")
}

func TestMockStore_Check_Duplicate(t *testing.T) {
	// Arrange
	store := NewMockStore()
	ctx := context.Background()

	// Act - First check
	isNew1, err1 := store.Check(ctx, "test-key", time.Hour)
	require.NoError(t, err1)
	assert.True(t, isNew1)

	// Act - Second check (duplicate)
	isNew2, err2 := store.Check(ctx, "test-key", time.Hour)
	require.NoError(t, err2)
	assert.False(t, isNew2, "Second occurrence should return false")
}

func TestMockStore_Check_EmptyKey(t *testing.T) {
	// Arrange
	store := NewMockStore()
	ctx := context.Background()

	// Act
	isNew, err := store.Check(ctx, "", time.Hour)

	// Assert
	require.NoError(t, err)
	assert.True(t, isNew, "Empty key should return true (no idempotency)")
}

func TestMockStore_StoreAndGetResult(t *testing.T) {
	// Arrange
	store := NewMockStore()
	ctx := context.Background()
	key := "test-key"
	expectedResult := []byte(`{"status": "success"}`)

	// Act - Store
	err := store.StoreResult(ctx, key, expectedResult, time.Hour)
	require.NoError(t, err)

	// Act - Get
	result, found, err := store.GetResult(ctx, key)

	// Assert
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, expectedResult, result)
}

func TestMockStore_GetResult_NotFound(t *testing.T) {
	// Arrange
	store := NewMockStore()
	ctx := context.Background()

	// Act
	result, found, err := store.GetResult(ctx, "nonexistent")

	// Assert
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, result)
}

// === IdempotentHandler Tests ===

func TestIdempotentHandler_ProcessesNewTask(t *testing.T) {
	// Arrange
	store := NewMockStore()
	processed := false
	handler := func(ctx context.Context, task *asynq.Task) error {
		processed = true
		return nil
	}

	keyExtractor := func(t *asynq.Task) string {
		return "task-key-1"
	}

	wrappedHandler := IdempotentHandler(store, keyExtractor, time.Hour, handler)
	task := asynq.NewTask("test:task", nil)

	// Act
	err := wrappedHandler(context.Background(), task)

	// Assert
	require.NoError(t, err)
	assert.True(t, processed, "New task should be processed")
}

func TestIdempotentHandler_SkipsDuplicateTask(t *testing.T) {
	// Arrange
	store := NewMockStore()
	processCount := 0
	handler := func(ctx context.Context, task *asynq.Task) error {
		processCount++
		return nil
	}

	keyExtractor := func(t *asynq.Task) string {
		return "task-key-1"
	}

	wrappedHandler := IdempotentHandler(store, keyExtractor, time.Hour, handler)
	task := asynq.NewTask("test:task", nil)

	// Act - First call
	err1 := wrappedHandler(context.Background(), task)
	require.NoError(t, err1)

	// Act - Second call (duplicate)
	err2 := wrappedHandler(context.Background(), task)
	require.NoError(t, err2)

	// Assert
	assert.Equal(t, 1, processCount, "Duplicate task should be skipped")
}

func TestIdempotentHandler_ProcessesWithEmptyKey(t *testing.T) {
	// Arrange
	store := NewMockStore()
	processed := false
	handler := func(ctx context.Context, task *asynq.Task) error {
		processed = true
		return nil
	}

	keyExtractor := func(t *asynq.Task) string {
		return "" // Empty key = no idempotency
	}

	wrappedHandler := IdempotentHandler(store, keyExtractor, time.Hour, handler)
	task := asynq.NewTask("test:task", nil)

	// Act
	err := wrappedHandler(context.Background(), task)

	// Assert
	require.NoError(t, err)
	assert.True(t, processed, "Task with empty key should be processed")
	assert.Equal(t, 0, store.GetCallCount(), "Store should not be called for empty key")
}

func TestIdempotentHandler_FailOpen_ProcessesOnError(t *testing.T) {
	// Arrange
	store := NewMockStore()
	store.SetError(errors.New("redis connection failed"), FailOpen)

	processed := false
	handler := func(ctx context.Context, task *asynq.Task) error {
		processed = true
		return nil
	}

	keyExtractor := func(t *asynq.Task) string {
		return "task-key-1"
	}

	wrappedHandler := IdempotentHandler(store, keyExtractor, time.Hour, handler,
		WithHandlerFailMode(FailOpen),
		WithHandlerLogger(zap.NewNop()),
	)
	task := asynq.NewTask("test:task", nil)

	// Act
	err := wrappedHandler(context.Background(), task)

	// Assert
	require.NoError(t, err)
	assert.True(t, processed, "Task should be processed on Redis error (fail-open)")
}

func TestIdempotentHandler_FailClosed_ReturnsErrorOnFailure(t *testing.T) {
	// Arrange
	store := NewMockStore()
	store.SetError(errors.New("redis connection failed"), FailClosed)

	processed := false
	handler := func(ctx context.Context, task *asynq.Task) error {
		processed = true
		return nil
	}

	keyExtractor := func(t *asynq.Task) string {
		return "task-key-1"
	}

	wrappedHandler := IdempotentHandler(store, keyExtractor, time.Hour, handler,
		WithHandlerFailMode(FailClosed),
		WithHandlerLogger(zap.NewNop()),
	)
	task := asynq.NewTask("test:task", nil)

	// Act
	err := wrappedHandler(context.Background(), task)

	// Assert
	require.Error(t, err)
	assert.False(t, processed, "Task should not be processed on Redis error (fail-closed)")
}

func TestIdempotentHandler_PropagatesHandlerError(t *testing.T) {
	// Arrange
	store := NewMockStore()
	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, task *asynq.Task) error {
		return expectedErr
	}

	keyExtractor := func(t *asynq.Task) string {
		return "task-key-1"
	}

	wrappedHandler := IdempotentHandler(store, keyExtractor, time.Hour, handler)
	task := asynq.NewTask("test:task", nil)

	// Act
	err := wrappedHandler(context.Background(), task)

	// Assert
	assert.Equal(t, expectedErr, err, "Handler error should be propagated")
}

func TestIdempotentHandler_Concurrency(t *testing.T) {
	// Arrange
	store := NewMockStore()
	var mu sync.Mutex
	processCount := 0
	handler := func(ctx context.Context, task *asynq.Task) error {
		mu.Lock()
		processCount++
		mu.Unlock()
		return nil
	}

	keyExtractor := func(t *asynq.Task) string {
		return "shared-key"
	}

	wrappedHandler := IdempotentHandler(store, keyExtractor, time.Hour, handler)
	task := asynq.NewTask("test:task", nil)

	// Act - Concurrent calls
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = wrappedHandler(context.Background(), task)
		}()
	}
	wg.Wait()

	// Assert
	assert.Equal(t, 1, processCount, "Only one task should be processed despite concurrent calls")
}

// === Nil Parameter Validation Tests ===

func TestIdempotentHandler_PanicsOnNilStore(t *testing.T) {
	// Arrange
	handler := func(ctx context.Context, task *asynq.Task) error { return nil }
	keyExtractor := func(t *asynq.Task) string { return "key" }

	// Act & Assert
	assert.Panics(t, func() {
		IdempotentHandler(nil, keyExtractor, time.Hour, handler)
	}, "Should panic when store is nil")
}

func TestIdempotentHandler_PanicsOnNilKeyExtractor(t *testing.T) {
	// Arrange
	store := NewMockStore()
	handler := func(ctx context.Context, task *asynq.Task) error { return nil }

	// Act & Assert
	assert.Panics(t, func() {
		IdempotentHandler(store, nil, time.Hour, handler)
	}, "Should panic when keyExtractor is nil")
}

func TestIdempotentHandler_PanicsOnNilHandler(t *testing.T) {
	// Arrange
	store := NewMockStore()
	keyExtractor := func(t *asynq.Task) string { return "key" }

	// Act & Assert
	assert.Panics(t, func() {
		IdempotentHandler(store, keyExtractor, time.Hour, nil)
	}, "Should panic when handler is nil")
}
