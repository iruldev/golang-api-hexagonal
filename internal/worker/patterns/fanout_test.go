package patterns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/iruldev/golang-api-hexagonal/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockEnqueuer implements tasks.TaskEnqueuer for testing
type mockEnqueuer struct {
	enqueuedTasks []*asynq.Task
	options       [][]asynq.Option
	err           error
	mu            sync.Mutex
}

func (m *mockEnqueuer) Enqueue(_ context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}
	m.enqueuedTasks = append(m.enqueuedTasks, task)
	m.options = append(m.options, opts)
	return &asynq.TaskInfo{ID: fmt.Sprintf("task-%d", len(m.enqueuedTasks)), Queue: worker.QueueDefault}, nil
}

func (m *mockEnqueuer) getEnqueuedTasks() []*asynq.Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enqueuedTasks
}

// ============================================================================
// FanoutRegistry Tests
// ============================================================================

func TestNewFanoutRegistry(t *testing.T) {
	registry := NewFanoutRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.handlers)
	assert.Empty(t, registry.handlers)
}

func TestFanoutRegistry_Register(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	err := registry.Register("user:created", "handler-1", handler)
	require.NoError(t, err)

	handlers := registry.Handlers("user:created")
	require.Len(t, handlers, 1)
	assert.Equal(t, "handler-1", handlers[0].ID)
	assert.Equal(t, worker.QueueDefault, handlers[0].Queue)
}

func TestFanoutRegistry_RegisterMultiple(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	require.NoError(t, registry.Register("user:created", "handler-1", handler))
	require.NoError(t, registry.Register("user:created", "handler-2", handler))
	require.NoError(t, registry.Register("order:completed", "handler-3", handler))

	userHandlers := registry.Handlers("user:created")
	assert.Len(t, userHandlers, 2)

	orderHandlers := registry.Handlers("order:completed")
	assert.Len(t, orderHandlers, 1)
}

func TestFanoutRegistry_RegisterWithQueue(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	err := registry.RegisterWithQueue("user:created", "critical-handler", handler, worker.QueueCritical)
	require.NoError(t, err)

	handlers := registry.Handlers("user:created")
	require.Len(t, handlers, 1)
	assert.Equal(t, worker.QueueCritical, handlers[0].Queue)
}

func TestFanoutRegistry_RegisterWithOptions(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	err := registry.Register("user:created", "handler-1", handler, asynq.MaxRetry(5), asynq.Timeout(30*time.Second))
	require.NoError(t, err)

	handlers := registry.Handlers("user:created")
	require.Len(t, handlers, 1)
	assert.Len(t, handlers[0].Opts, 2)
}

func TestFanoutRegistry_Handlers_Empty(t *testing.T) {
	registry := NewFanoutRegistry()

	handlers := registry.Handlers("nonexistent")
	assert.Nil(t, handlers)
}

func TestFanoutRegistry_Handlers_ReturnsCopy(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	require.NoError(t, registry.Register("user:created", "handler-1", handler))

	handlers1 := registry.Handlers("user:created")
	handlers2 := registry.Handlers("user:created")

	// Modify one copy
	handlers1[0].ID = "modified"

	// Original should not be affected
	assert.Equal(t, "handler-1", handlers2[0].ID)
}

func TestFanoutRegistry_Unregister(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	require.NoError(t, registry.Register("user:created", "handler-1", handler))
	require.NoError(t, registry.Register("user:created", "handler-2", handler))
	require.NoError(t, registry.Register("user:created", "handler-3", handler))

	registry.Unregister("user:created", "handler-2")

	handlers := registry.Handlers("user:created")
	require.Len(t, handlers, 2)
	assert.Equal(t, "handler-1", handlers[0].ID)
	assert.Equal(t, "handler-3", handlers[1].ID)
}

func TestFanoutRegistry_Unregister_NonExistent(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	require.NoError(t, registry.Register("user:created", "handler-1", handler))

	// Should not panic
	registry.Unregister("user:created", "nonexistent")
	registry.Unregister("nonexistent:event", "handler-1")

	handlers := registry.Handlers("user:created")
	assert.Len(t, handlers, 1)
}

func TestFanoutRegistry_ThreadSafety(t *testing.T) {
	registry := NewFanoutRegistry()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = registry.Register("test", fmt.Sprintf("handler-%d", n),
				func(_ context.Context, _ FanoutEvent) error { return nil })
		}(i)
	}

	// Concurrent reads while writing
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = registry.Handlers("test")
		}()
	}

	wg.Wait()
	handlers := registry.Handlers("test")
	assert.Len(t, handlers, 100)
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestFanoutRegistry_Register_Validation(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	tests := []struct {
		name      string
		eventType string
		handlerID string
		fn        FanoutHandlerFunc
		wantErr   error
	}{
		{"empty event type", "", "handler", handler, ErrEmptyEventType},
		{"empty handler ID", "event", "", handler, ErrEmptyHandlerID},
		{"nil handler", "event", "handler", nil, ErrNilHandler},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.eventType, tt.handlerID, tt.fn)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestFanoutRegistry_RegisterWithQueue_Validation(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	tests := []struct {
		name      string
		eventType string
		handlerID string
		fn        FanoutHandlerFunc
		queue     string
		wantErr   error
	}{
		{"empty event type", "", "handler", handler, worker.QueueCritical, ErrEmptyEventType},
		{"empty handler ID", "event", "", handler, worker.QueueCritical, ErrEmptyHandlerID},
		{"nil handler", "event", "handler", nil, worker.QueueCritical, ErrNilHandler},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterWithQueue(tt.eventType, tt.handlerID, tt.fn, tt.queue)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestFanoutRegistry_Register_Duplicate(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	// First registration should succeed
	err := registry.Register("user:created", "handler-1", handler)
	require.NoError(t, err)

	// Second registration with same ID should fail
	err = registry.Register("user:created", "handler-1", handler)
	assert.ErrorIs(t, err, ErrDuplicateHandler)

	// Same handler ID on different event type should succeed
	err = registry.Register("order:completed", "handler-1", handler)
	require.NoError(t, err)
}

func TestFanoutRegistry_RegisterWithQueue_EmptyQueueUsesDefault(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	err := registry.RegisterWithQueue("user:created", "handler-1", handler, "")
	require.NoError(t, err)

	handlers := registry.Handlers("user:created")
	require.Len(t, handlers, 1)
	assert.Equal(t, worker.QueueDefault, handlers[0].Queue)
}

// ============================================================================
// Fanout Function Tests
// ============================================================================

func TestFanout_SingleHandler(t *testing.T) {
	registry := NewFanoutRegistry()
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	require.NoError(t, registry.Register("user:created", "welcome-email", handler))

	event := FanoutEvent{
		Type:    "user:created",
		Payload: json.RawMessage(`{"user_id": "123"}`),
	}

	errors := Fanout(context.Background(), enqueuer, registry, logger, event)

	assert.Empty(t, errors)
	assert.Len(t, enqueuer.getEnqueuedTasks(), 1)
	assert.Equal(t, "fanout:user:created:welcome-email", enqueuer.enqueuedTasks[0].Type())
}

func TestFanout_MultipleHandlers(t *testing.T) {
	registry := NewFanoutRegistry()
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	require.NoError(t, registry.Register("user:created", "welcome-email", handler))
	require.NoError(t, registry.Register("user:created", "default-settings", handler))
	require.NoError(t, registry.Register("user:created", "notify-admin", handler))

	event := FanoutEvent{
		Type:    "user:created",
		Payload: json.RawMessage(`{"user_id": "123"}`),
	}

	errs := Fanout(context.Background(), enqueuer, registry, logger, event)

	assert.Empty(t, errs)
	assert.Len(t, enqueuer.getEnqueuedTasks(), 3)
}

func TestFanout_NoHandlers(t *testing.T) {
	registry := NewFanoutRegistry()
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()

	event := FanoutEvent{
		Type:    "nonexistent:event",
		Payload: json.RawMessage(`{}`),
	}

	errs := Fanout(context.Background(), enqueuer, registry, logger, event)

	assert.Nil(t, errs)
	assert.Empty(t, enqueuer.getEnqueuedTasks())
}

func TestFanout_EnqueueError(t *testing.T) {
	registry := NewFanoutRegistry()
	enqueuer := &mockEnqueuer{err: errors.New("redis connection failed")}
	logger := zap.NewNop()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	require.NoError(t, registry.Register("user:created", "handler-1", handler))
	require.NoError(t, registry.Register("user:created", "handler-2", handler))

	event := FanoutEvent{
		Type:    "user:created",
		Payload: json.RawMessage(`{}`),
	}

	errs := Fanout(context.Background(), enqueuer, registry, logger, event)

	// Both handlers should fail
	assert.Len(t, errs, 2)
}

func TestFanout_PartialFailure(t *testing.T) {
	registry := NewFanoutRegistry()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	// Create a custom enqueuer that fails on second call
	customEnqueuer := &partialFailEnqueuer{failOnCall: 2}

	require.NoError(t, registry.Register("user:created", "handler-1", handler))
	require.NoError(t, registry.Register("user:created", "handler-2", handler))
	require.NoError(t, registry.Register("user:created", "handler-3", handler))

	event := FanoutEvent{
		Type:    "user:created",
		Payload: json.RawMessage(`{}`),
	}

	logger := zap.NewNop()
	errs := Fanout(context.Background(), customEnqueuer, registry, logger, event)

	// Only one handler should fail
	assert.Len(t, errs, 1)
}

type partialFailEnqueuer struct {
	callCount  int
	failOnCall int
	mu         sync.Mutex
}

func (e *partialFailEnqueuer) Enqueue(_ context.Context, task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	e.mu.Lock()
	e.callCount++
	count := e.callCount
	e.mu.Unlock()

	if count == e.failOnCall {
		return nil, errors.New("simulated failure")
	}
	return &asynq.TaskInfo{ID: fmt.Sprintf("task-%d", count), Queue: worker.QueueDefault}, nil
}

func TestFanout_AutoSetTimestamp(t *testing.T) {
	registry := NewFanoutRegistry()
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	_ = registry.Register("test", "handler", handler)

	before := time.Now().UTC()
	event := FanoutEvent{
		Type:    "test",
		Payload: json.RawMessage(`{}`),
		// Timestamp not set
	}

	_ = Fanout(context.Background(), enqueuer, registry, logger, event)
	after := time.Now().UTC()

	// Verify payload was marshaled with timestamp
	require.Len(t, enqueuer.enqueuedTasks, 1)
	var savedEvent FanoutEvent
	err := json.Unmarshal(enqueuer.enqueuedTasks[0].Payload(), &savedEvent)
	require.NoError(t, err)

	assert.False(t, savedEvent.Timestamp.IsZero())
	assert.True(t, savedEvent.Timestamp.After(before) || savedEvent.Timestamp.Equal(before))
	assert.True(t, savedEvent.Timestamp.Before(after) || savedEvent.Timestamp.Equal(after))
}

func TestFanout_PreserveTimestamp(t *testing.T) {
	registry := NewFanoutRegistry()
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	_ = registry.Register("test", "handler", handler)

	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	event := FanoutEvent{
		Type:      "test",
		Payload:   json.RawMessage(`{}`),
		Timestamp: fixedTime,
	}

	_ = Fanout(context.Background(), enqueuer, registry, logger, event)

	require.Len(t, enqueuer.enqueuedTasks, 1)
	var savedEvent FanoutEvent
	err := json.Unmarshal(enqueuer.enqueuedTasks[0].Payload(), &savedEvent)
	require.NoError(t, err)

	assert.Equal(t, fixedTime, savedEvent.Timestamp)
}

func TestFanout_EventMetadata(t *testing.T) {
	registry := NewFanoutRegistry()
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	handler := func(_ context.Context, _ FanoutEvent) error { return nil }

	_ = registry.Register("test", "handler", handler)

	event := FanoutEvent{
		Type:    "test",
		Payload: json.RawMessage(`{"key": "value"}`),
		Metadata: map[string]string{
			"trace_id":       "abc123",
			"correlation_id": "def456",
		},
	}

	_ = Fanout(context.Background(), enqueuer, registry, logger, event)

	require.Len(t, enqueuer.enqueuedTasks, 1)
	var savedEvent FanoutEvent
	err := json.Unmarshal(enqueuer.enqueuedTasks[0].Payload(), &savedEvent)
	require.NoError(t, err)

	assert.Equal(t, "abc123", savedEvent.Metadata["trace_id"])
	assert.Equal(t, "def456", savedEvent.Metadata["correlation_id"])
}

// ============================================================================
// FanoutDispatcher Tests
// ============================================================================

func TestNewFanoutDispatcher(t *testing.T) {
	registry := NewFanoutRegistry()
	logger := zap.NewNop()

	dispatcher := NewFanoutDispatcher(registry, logger)

	assert.NotNil(t, dispatcher)
	assert.Equal(t, registry, dispatcher.registry)
	assert.Equal(t, logger, dispatcher.logger)
}

func TestFanoutDispatcher_Handle_Success(t *testing.T) {
	registry := NewFanoutRegistry()
	logger := zap.NewNop()
	var handlerCalled int32

	_ = registry.Register("user:created", "test-handler", func(_ context.Context, event FanoutEvent) error {
		atomic.AddInt32(&handlerCalled, 1)
		assert.Equal(t, "user:created", event.Type)
		return nil
	})

	dispatcher := NewFanoutDispatcher(registry, logger)

	event := FanoutEvent{
		Type:    "user:created",
		Payload: json.RawMessage(`{"user_id": "123"}`),
	}
	payload, _ := json.Marshal(event)
	task := asynq.NewTask("fanout:user:created:test-handler", payload)

	err := dispatcher.Handle(context.Background(), task)

	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&handlerCalled))
}

func TestFanoutDispatcher_Handle_HandlerError(t *testing.T) {
	registry := NewFanoutRegistry()
	logger := zap.NewNop()

	_ = registry.Register("user:created", "failing-handler", func(_ context.Context, _ FanoutEvent) error {
		return errors.New("handler failed")
	})

	dispatcher := NewFanoutDispatcher(registry, logger)

	event := FanoutEvent{Type: "user:created", Payload: json.RawMessage(`{}`)}
	payload, _ := json.Marshal(event)
	task := asynq.NewTask("fanout:user:created:failing-handler", payload)

	err := dispatcher.Handle(context.Background(), task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler failed")
}

func TestFanoutDispatcher_Handle_InvalidTaskType(t *testing.T) {
	registry := NewFanoutRegistry()
	logger := zap.NewNop()
	dispatcher := NewFanoutDispatcher(registry, logger)

	tests := []struct {
		name     string
		taskType string
	}{
		{"missing parts", "fanout"},
		{"only two parts", "fanout:user"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := asynq.NewTask(tt.taskType, []byte(`{}`))
			err := dispatcher.Handle(context.Background(), task)

			assert.Error(t, err)
			assert.True(t, errors.Is(err, asynq.SkipRetry))
		})
	}
}

func TestFanoutDispatcher_Handle_HandlerNotFound(t *testing.T) {
	registry := NewFanoutRegistry()
	logger := zap.NewNop()
	dispatcher := NewFanoutDispatcher(registry, logger)

	event := FanoutEvent{Type: "user:created", Payload: json.RawMessage(`{}`)}
	payload, _ := json.Marshal(event)
	task := asynq.NewTask("fanout:user:created:nonexistent", payload)

	err := dispatcher.Handle(context.Background(), task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler nonexistent not found")
	assert.True(t, errors.Is(err, asynq.SkipRetry))
}

func TestFanoutDispatcher_Handle_InvalidPayload(t *testing.T) {
	registry := NewFanoutRegistry()
	logger := zap.NewNop()

	_ = registry.Register("user:created", "handler", func(_ context.Context, _ FanoutEvent) error {
		return nil
	})

	dispatcher := NewFanoutDispatcher(registry, logger)

	task := asynq.NewTask("fanout:user:created:handler", []byte(`{invalid json`))

	err := dispatcher.Handle(context.Background(), task)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, asynq.SkipRetry))
}

func TestFanoutDispatcher_Handle_Isolation(t *testing.T) {
	// Verify that handler isolation works - each handler runs in its own task
	registry := NewFanoutRegistry()
	logger := zap.NewNop()

	var handler1Called, handler2Called int32

	_ = registry.Register("test", "handler-1", func(_ context.Context, _ FanoutEvent) error {
		atomic.AddInt32(&handler1Called, 1)
		return errors.New("handler 1 failed")
	})
	_ = registry.Register("test", "handler-2", func(_ context.Context, _ FanoutEvent) error {
		atomic.AddInt32(&handler2Called, 1)
		return nil
	})

	dispatcher := NewFanoutDispatcher(registry, logger)

	event := FanoutEvent{Type: "test", Payload: json.RawMessage(`{}`)}
	payload, _ := json.Marshal(event)

	// Handler 1 fails
	task1 := asynq.NewTask("fanout:test:handler-1", payload)
	err1 := dispatcher.Handle(context.Background(), task1)
	assert.Error(t, err1)

	// Handler 2 succeeds (isolated from handler 1's failure)
	task2 := asynq.NewTask("fanout:test:handler-2", payload)
	err2 := dispatcher.Handle(context.Background(), task2)
	assert.NoError(t, err2)

	assert.Equal(t, int32(1), atomic.LoadInt32(&handler1Called))
	assert.Equal(t, int32(1), atomic.LoadInt32(&handler2Called))
}

// ============================================================================
// TaskTypePrefix Tests
// ============================================================================

func TestTaskTypePrefix(t *testing.T) {
	prefix := TaskTypePrefix()
	assert.Equal(t, "fanout:", prefix)
}
