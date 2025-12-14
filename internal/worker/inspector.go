// Package worker provides Asynq background job processing infrastructure.
package worker

import (
	"context"

	"github.com/hibiken/asynq"

	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// ValidQueues contains the known queue names.
var ValidQueues = []string{QueueCritical, QueueDefault, QueueLow}

// IsValidQueue checks if a queue name is valid.
func IsValidQueue(queue string) bool {
	for _, q := range ValidQueues {
		if q == queue {
			return true
		}
	}
	return false
}

// AsynqQueueInspector implements QueueInspector using asynq.Inspector.
type AsynqQueueInspector struct {
	inspector *asynq.Inspector
}

// NewAsynqQueueInspector creates a new AsynqQueueInspector.
func NewAsynqQueueInspector(redisOpt asynq.RedisClientOpt) *AsynqQueueInspector {
	return &AsynqQueueInspector{
		inspector: asynq.NewInspector(redisOpt),
	}
}

// GetQueueStats returns statistics for all queues.
func (i *AsynqQueueInspector) GetQueueStats(ctx context.Context) (*runtimeutil.QueueStats, error) {
	stats := &runtimeutil.QueueStats{
		Aggregate: runtimeutil.AggregateStats{},
		Queues:    make([]runtimeutil.QueueInfo, 0, len(ValidQueues)),
	}

	for _, queueName := range ValidQueues {
		qInfo, err := i.inspector.GetQueueInfo(queueName)
		if err != nil {
			// If we can't get info for a known queue, it might be a connection issue
			// or the queue is truly in a bad state. We should report this.
			// However, asynq returns error if queue doesn't exist in Redis yet.
			// Since we know these are valid queue names, we'll try to distinguish
			// or just return the error to be safe if it's critical.

			// For now, we will assume that if we can't inspect a valid queue,
			// something is wrong with the infrastructure (Redis).
			return nil, err
		}

		queueInfo := runtimeutil.QueueInfo{
			Name:      queueName,
			Size:      qInfo.Size,
			Active:    qInfo.Active,
			Pending:   qInfo.Pending,
			Scheduled: qInfo.Scheduled,
			Retry:     qInfo.Retry,
			Archived:  qInfo.Archived,
			Completed: qInfo.Completed,
			Processed: int(qInfo.Processed),
			Failed:    int(qInfo.Failed),
		}
		stats.Queues = append(stats.Queues, queueInfo)

		// Aggregate stats
		stats.Aggregate.TotalEnqueued += qInfo.Size
		stats.Aggregate.TotalActive += qInfo.Active
		stats.Aggregate.TotalPending += qInfo.Pending
		stats.Aggregate.TotalScheduled += qInfo.Scheduled
		stats.Aggregate.TotalRetry += qInfo.Retry
		stats.Aggregate.TotalArchived += qInfo.Archived
		stats.Aggregate.TotalCompleted += qInfo.Completed
		stats.Aggregate.TotalProcessed += int(qInfo.Processed)
		stats.Aggregate.TotalFailed += int(qInfo.Failed)
	}

	return stats, nil
}

// GetJobsInQueue returns jobs in a specific queue with pagination.
func (i *AsynqQueueInspector) GetJobsInQueue(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.JobList, error) {
	if !IsValidQueue(queueName) {
		return nil, runtimeutil.ErrInvalidQueue
	}

	// Default pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// List pending tasks
	tasks, err := i.inspector.ListPendingTasks(queueName, asynq.PageSize(pageSize), asynq.Page(page))
	if err != nil {
		return nil, err
	}

	// Get queue info for total count
	qInfo, err := i.inspector.GetQueueInfo(queueName)
	if err != nil {
		return nil, err
	}

	jobs := make([]runtimeutil.JobInfo, 0, len(tasks))
	for _, t := range tasks {
		jobs = append(jobs, runtimeutil.JobInfo{
			TaskID:         t.ID,
			Type:           t.Type,
			PayloadPreview: truncatePayload(t.Payload, 100),
			State:          t.State.String(),
			Queue:          t.Queue,
			MaxRetry:       t.MaxRetry,
			Retried:        t.Retried,
			CreatedAt:      t.NextProcessAt, // Pending tasks use NextProcessAt
		})
	}

	totalPages := (qInfo.Pending + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return &runtimeutil.JobList{
		Jobs: jobs,
		Pagination: runtimeutil.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      qInfo.Pending,
			TotalPages: totalPages,
		},
	}, nil
}

// GetFailedJobs returns failed jobs in a specific queue with pagination.
func (i *AsynqQueueInspector) GetFailedJobs(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.FailedJobList, error) {
	if !IsValidQueue(queueName) {
		return nil, runtimeutil.ErrInvalidQueue
	}

	// Default pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// List archived (failed/dead) tasks
	tasks, err := i.inspector.ListArchivedTasks(queueName, asynq.PageSize(pageSize), asynq.Page(page))
	if err != nil {
		return nil, err
	}

	// Get queue info for total count
	qInfo, err := i.inspector.GetQueueInfo(queueName)
	if err != nil {
		return nil, err
	}

	failedJobs := make([]runtimeutil.FailedJobInfo, 0, len(tasks))
	for _, t := range tasks {
		failedJobs = append(failedJobs, runtimeutil.FailedJobInfo{
			TaskID:         t.ID,
			Type:           t.Type,
			PayloadPreview: truncatePayload(t.Payload, 100),
			ErrorMessage:   t.LastErr,
			FailedAt:       t.LastFailedAt,
			RetryCount:     t.Retried,
			MaxRetry:       t.MaxRetry,
		})
	}

	totalPages := (qInfo.Archived + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return &runtimeutil.FailedJobList{
		FailedJobs: failedJobs,
		Pagination: runtimeutil.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      qInfo.Archived,
			TotalPages: totalPages,
		},
	}, nil
}

// DeleteFailedJob deletes a failed job from a queue.
func (i *AsynqQueueInspector) DeleteFailedJob(ctx context.Context, queueName, taskID string) error {
	if !IsValidQueue(queueName) {
		return runtimeutil.ErrInvalidQueue
	}

	err := i.inspector.DeleteTask(queueName, taskID)
	if err != nil {
		if err == asynq.ErrTaskNotFound {
			return runtimeutil.ErrTaskNotFound
		}
		return err
	}
	return nil
}

// RetryFailedJob requeues a failed job for retry.
func (i *AsynqQueueInspector) RetryFailedJob(ctx context.Context, queueName, taskID string) (*runtimeutil.JobInfo, error) {
	if !IsValidQueue(queueName) {
		return nil, runtimeutil.ErrInvalidQueue
	}

	err := i.inspector.RunTask(queueName, taskID)
	if err != nil {
		if err == asynq.ErrTaskNotFound {
			return nil, runtimeutil.ErrTaskNotFound
		}
		return nil, err
	}

	// Return basic job info - task is now pending
	return &runtimeutil.JobInfo{
		TaskID: taskID,
		Queue:  queueName,
		State:  "pending",
	}, nil
}

// truncatePayload truncates payload to maxLen characters.
func truncatePayload(payload []byte, maxLen int) string {
	s := string(payload)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
