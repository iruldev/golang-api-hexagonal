package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskEnqueuer defines interface for enqueueing tasks from usecase layer.
// This allows usecases to depend on interface, not concrete worker.Client.
type TaskEnqueuer interface {
	Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}
