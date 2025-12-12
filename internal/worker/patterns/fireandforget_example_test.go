package patterns_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/worker"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
)

// This file contains example usage documentation for the Fire-and-Forget pattern.
// Examples are structured as functions but are not executed as actual tests.

// ExampleFireAndForget_basic demonstrates basic fire-and-forget usage.
// The function returns immediately without waiting for the task to be enqueued.
//
// Usage:
//
//	// In real code, these would be injected dependencies
//	client := worker.NewClient(redisOpt)
//	logger := app.Logger
//
//	// Create a task
//	noteID := uuid.New()
//	task, _ := tasks.NewNoteArchiveTask(noteID)
//
//	// Fire and forget - returns immediately
//	patterns.FireAndForget(context.Background(), client, logger, task)
//
//	// Continue with other work - the task is being enqueued in the background
func ExampleFireAndForget_basic() {
	// This example shows the pattern but doesn't execute
	// to avoid requiring a real Redis connection.
	_ = patterns.FireAndForget
}

// ExampleFireAndForget_withQueueOverride shows how to override the default queue.
// By default, FireAndForget uses the "low" queue, but you can override this.
//
// Usage:
//
//	task := asynq.NewTask("notification:send", []byte(`{"user_id":"123"}`))
//
//	// Override to default queue for slightly higher priority
//	patterns.FireAndForget(ctx, client, logger, task,
//	    asynq.Queue(worker.QueueDefault))
//
//	// Or even critical queue for important notifications
//	patterns.FireAndForget(ctx, client, logger, task,
//	    asynq.Queue(worker.QueueCritical))
func ExampleFireAndForget_withQueueOverride() {
	// Pattern demonstration - queue constants are available
	_ = worker.QueueDefault
	_ = worker.QueueCritical
	_ = worker.QueueLow
}

// ExampleFireAndForget_withOptions shows how to add additional task options.
//
// Usage:
//
//	task := asynq.NewTask("analytics:track", []byte(`{"event":"pageview"}`))
//
//	// Add multiple options - queue override, max retry, etc.
//	patterns.FireAndForget(ctx, client, logger, task,
//	    asynq.Queue(worker.QueueLow), // Explicit low queue (same as default)
//	    asynq.MaxRetry(1),            // Only retry once for analytics
//	)
func ExampleFireAndForget_withOptions() {
	// Pattern demonstration - show available options
	_ = asynq.MaxRetry
	_ = asynq.Queue
}

// ExampleFireAndForget_usecaseIntegration demonstrates how to use FireAndForget in a usecase.
// This is the recommended pattern for triggering fire-and-forget jobs from business logic.
//
// Type definition:
//
//	type NoteUsecase struct {
//	    repo     note.Repository
//	    enqueuer tasks.TaskEnqueuer  // *worker.Client implements this
//	    logger   *zap.Logger
//	}
//
// Method implementation:
//
//	func (u *NoteUsecase) Update(ctx context.Context, noteID uuid.UUID, data UpdateData) error {
//	    // 1. Critical business logic (must succeed)
//	    if err := u.repo.Update(ctx, noteID, data); err != nil {
//	        return err
//	    }
//
//	    // 2. Fire-and-forget for non-critical follow-up
//	    task, err := tasks.NewNoteArchiveTask(noteID)
//	    if err != nil {
//	        u.logger.Warn("failed to create audit task", zap.Error(err))
//	        return nil  // Don't fail main operation
//	    }
//
//	    // Returns immediately - main operation not blocked
//	    patterns.FireAndForget(ctx, u.enqueuer, u.logger, task)
//
//	    return nil
//	}
func ExampleFireAndForget_usecaseIntegration() {
	// Pattern demonstration - show interface usage
	var _ tasks.TaskEnqueuer // *worker.Client implements this
}

// The following are type assertions to ensure examples compile correctly
var (
	_ = context.Background
	_ = uuid.New
	_ = zap.NewNop
	_ = asynq.NewTask
	_ = patterns.FireAndForget
	_ = tasks.NewNoteArchiveTask
	_ = worker.QueueLow
)
