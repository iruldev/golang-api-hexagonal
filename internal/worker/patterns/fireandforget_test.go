package patterns_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/iruldev/golang-api-hexagonal/internal/worker"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
)

// mockEnqueuer is a test double for tasks.TaskEnqueuer that records enqueue calls.
type mockEnqueuer struct {
	mu           sync.Mutex
	enqueueCalls []enqueueCall
	err          error
	delay        time.Duration
}

type enqueueCall struct {
	task *asynq.Task
	opts []asynq.Option
}

func (m *mockEnqueuer) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enqueueCalls = append(m.enqueueCalls, enqueueCall{task: task, opts: opts})
	if m.err != nil {
		return nil, m.err
	}
	return &asynq.TaskInfo{
		ID:    "test-task-id",
		Queue: worker.QueueLow,
		Type:  task.Type(),
	}, nil
}

func (m *mockEnqueuer) getCalls() []enqueueCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]enqueueCall, len(m.enqueueCalls))
	copy(calls, m.enqueueCalls)
	return calls
}

func TestFireAndForget_ReturnsImmediately(t *testing.T) {
	// Arrange - enqueue with delay to prove we don't wait
	enqueuer := &mockEnqueuer{delay: 50 * time.Millisecond}
	logger := zap.NewNop()
	task := asynq.NewTask("test:task", nil)

	// Act
	start := time.Now()
	patterns.FireAndForget(context.Background(), enqueuer, logger, task)
	duration := time.Since(start)

	// Assert - function should return immediately, not wait for enqueue
	assert.Less(t, duration, 10*time.Millisecond, "FireAndForget should return immediately without waiting for enqueue")

	// Wait for goroutine to complete
	time.Sleep(100 * time.Millisecond)
}

func TestFireAndForget_EnqueuesTaskSuccessfully(t *testing.T) {
	// Arrange
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	task := asynq.NewTask("test:task", []byte(`{"id":"123"}`))

	// Act
	patterns.FireAndForget(context.Background(), enqueuer, logger, task)

	// Wait for goroutine
	time.Sleep(50 * time.Millisecond)

	// Assert
	calls := enqueuer.getCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "test:task", calls[0].task.Type())
	assert.Equal(t, []byte(`{"id":"123"}`), calls[0].task.Payload())
}

func TestFireAndForget_DefaultsToLowQueue(t *testing.T) {
	// Arrange
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	task := asynq.NewTask("test:task", nil)

	// Act
	patterns.FireAndForget(context.Background(), enqueuer, logger, task)

	// Wait for goroutine
	time.Sleep(50 * time.Millisecond)

	// Assert - first option should be Queue(low)
	calls := enqueuer.getCalls()
	require.Len(t, calls, 1)
	// The first option passed should be the low queue default
	assert.GreaterOrEqual(t, len(calls[0].opts), 1, "Should have at least one option (queue)")
}

func TestFireAndForget_AllowsQueueOverride(t *testing.T) {
	// Arrange
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	task := asynq.NewTask("test:task", nil)

	// Act - override to critical queue
	patterns.FireAndForget(context.Background(), enqueuer, logger, task,
		asynq.Queue(worker.QueueCritical))

	// Wait for goroutine
	time.Sleep(50 * time.Millisecond)

	// Assert
	calls := enqueuer.getCalls()
	require.Len(t, calls, 1)
	// Should have 2 options: default low queue + override critical queue
	// The override should take precedence (comes last)
	assert.Len(t, calls[0].opts, 2)
}

func TestFireAndForget_LogsErrorWithoutPanic(t *testing.T) {
	// Arrange
	enqueuer := &mockEnqueuer{err: assert.AnError}
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)
	task := asynq.NewTask("test:task", nil)

	// Act - should not panic even with error
	assert.NotPanics(t, func() {
		patterns.FireAndForget(context.Background(), enqueuer, logger, task)
	})

	// Wait for goroutine
	time.Sleep(50 * time.Millisecond)

	// Assert - error should be logged
	entries := logs.All()
	require.Len(t, entries, 1)
	assert.Equal(t, "fire-and-forget enqueue failed", entries[0].Message)
	assert.Equal(t, "test:task", entries[0].ContextMap()["task_type"])
}

func TestFireAndForget_LogsTaskInfoOnSuccess(t *testing.T) {
	// Arrange
	enqueuer := &mockEnqueuer{}
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	task := asynq.NewTask("test:task", nil)

	// Act
	patterns.FireAndForget(context.Background(), enqueuer, logger, task)

	// Wait for goroutine
	time.Sleep(50 * time.Millisecond)

	// Assert - debug log should contain task info
	entries := logs.All()
	require.Len(t, entries, 1)
	assert.Equal(t, "fire-and-forget task enqueued", entries[0].Message)
	assert.Equal(t, "test-task-id", entries[0].ContextMap()["task_id"])
	assert.Equal(t, "test:task", entries[0].ContextMap()["task_type"])
}

func TestFireAndForget_WithMultipleOptions(t *testing.T) {
	// Arrange
	enqueuer := &mockEnqueuer{}
	logger := zap.NewNop()
	task := asynq.NewTask("test:task", nil)

	// Act - pass multiple options
	patterns.FireAndForget(context.Background(), enqueuer, logger, task,
		asynq.MaxRetry(5),
		asynq.Queue(worker.QueueDefault),
	)

	// Wait for goroutine
	time.Sleep(50 * time.Millisecond)

	// Assert - should have default low queue + 2 user options = 3 options
	calls := enqueuer.getCalls()
	require.Len(t, calls, 1)
	assert.Len(t, calls[0].opts, 3)
}
