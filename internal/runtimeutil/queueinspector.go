// Package runtimeutil provides runtime utilities and interfaces for application configuration.
package runtimeutil

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors for queue inspection.
var (
	// ErrInvalidQueue is returned when an invalid queue name is specified.
	ErrInvalidQueue = errors.New("invalid queue name")
	// ErrTaskNotFound is returned when a task is not found in the queue.
	ErrTaskNotFound = errors.New("task not found")
)

// QueueStats represents queue statistics at a point in time.
type QueueStats struct {
	Aggregate AggregateStats `json:"aggregate"`
	Queues    []QueueInfo    `json:"queues"`
}

// AggregateStats contains totals across all queues.
type AggregateStats struct {
	TotalEnqueued  int `json:"total_enqueued"`
	TotalActive    int `json:"total_active"`
	TotalPending   int `json:"total_pending"`
	TotalScheduled int `json:"total_scheduled"`
	TotalRetry     int `json:"total_retry"`
	TotalArchived  int `json:"total_archived"`
	TotalCompleted int `json:"total_completed"`
	TotalProcessed int `json:"total_processed"`
	TotalFailed    int `json:"total_failed"`
}

// QueueInfo contains statistics for a single queue.
type QueueInfo struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Active    int    `json:"active"`
	Pending   int    `json:"pending"`
	Scheduled int    `json:"scheduled"`
	Retry     int    `json:"retry"`
	Archived  int    `json:"archived"`
	Completed int    `json:"completed"`
	Processed int    `json:"processed"`
	Failed    int    `json:"failed"`
}

// JobInfo represents a job in a queue.
type JobInfo struct {
	TaskID         string    `json:"task_id"`
	Type           string    `json:"type"`
	PayloadPreview string    `json:"payload_preview"`
	State          string    `json:"state"`
	Queue          string    `json:"queue"`
	MaxRetry       int       `json:"max_retry"`
	Retried        int       `json:"retried"`
	CreatedAt      time.Time `json:"created_at"`
}

// FailedJobInfo represents a failed job in a queue.
type FailedJobInfo struct {
	TaskID         string    `json:"task_id"`
	Type           string    `json:"type"`
	PayloadPreview string    `json:"payload_preview"`
	ErrorMessage   string    `json:"error_message"`
	FailedAt       time.Time `json:"failed_at"`
	RetryCount     int       `json:"retry_count"`
	MaxRetry       int       `json:"max_retry"`
}

// Pagination holds pagination information for list responses.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// JobList contains a paginated list of jobs.
type JobList struct {
	Jobs       []JobInfo  `json:"jobs"`
	Pagination Pagination `json:"pagination"`
}

// FailedJobList contains a paginated list of failed jobs.
type FailedJobList struct {
	FailedJobs []FailedJobInfo `json:"failed_jobs"`
	Pagination Pagination      `json:"pagination"`
}

// QueueInspector defines methods for inspecting job queues.
type QueueInspector interface {
	// GetQueueStats returns statistics for all queues.
	GetQueueStats(ctx context.Context) (*QueueStats, error)

	// GetJobsInQueue returns jobs in a specific queue with pagination.
	GetJobsInQueue(ctx context.Context, queueName string, page, pageSize int) (*JobList, error)

	// GetFailedJobs returns failed jobs in a specific queue with pagination.
	GetFailedJobs(ctx context.Context, queueName string, page, pageSize int) (*FailedJobList, error)

	// DeleteFailedJob deletes a failed job from a queue.
	DeleteFailedJob(ctx context.Context, queueName, taskID string) error

	// RetryFailedJob requeues a failed job for retry.
	RetryFailedJob(ctx context.Context, queueName, taskID string) (*JobInfo, error)
}
