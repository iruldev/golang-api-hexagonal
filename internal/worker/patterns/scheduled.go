// Package patterns provides reusable async job patterns for the worker infrastructure.
// This file contains the Scheduled Job pattern for periodic task execution.
//
// Scheduled Job Pattern:
//
// The scheduled job pattern is used for periodic tasks that run on a cron schedule:
//   - Database cleanup (daily)
//   - Report generation (weekly)
//   - Health checks (hourly)
//   - Cache invalidation (periodic)
//
// Scheduler runs as a SEPARATE process from Worker:
//   - Scheduler: Enqueues tasks on schedule (cmd/scheduler/main.go)
//   - Worker: Processes enqueued tasks (cmd/worker/main.go)
//
// Cron Expression Format (5 fields):
//
//	┌───────────── minute (0-59)
//	│ ┌───────────── hour (0-23)
//	│ │ ┌───────────── day of month (1-31)
//	│ │ │ ┌───────────── month (1-12)
//	│ │ │ │ ┌───────────── day of week (0-6, Sun=0)
//	│ │ │ │ │
//	* * * * *
//
// Common Examples:
//   - Every minute:     "* * * * *"
//   - Every hour:       "0 * * * *"
//   - Daily midnight:   "0 0 * * *"
//   - Monday 9am:       "0 9 * * 1"
//   - First of month:   "0 0 1 * *"
package patterns

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// ScheduledJob represents a periodic job configuration.
// Use this struct to define jobs that run on a cron schedule.
type ScheduledJob struct {
	// Cronspec is the cron expression (5 fields) defining when to run.
	// Examples: "0 0 * * *" (daily), "0 * * * *" (hourly), "*/5 * * * *" (every 5 min)
	Cronspec string

	// Task is the asynq task to enqueue when the schedule triggers.
	Task *asynq.Task

	// Opts are optional asynq options applied to the task (Queue, MaxRetry, etc.)
	Opts []asynq.Option

	// Description is a human-readable description for logging and documentation.
	Description string
}

// RegisterScheduledJobs registers all scheduled jobs with the scheduler.
// Returns entry IDs for each registered job for tracking purposes.
//
// Example usage:
//
//	jobs := []patterns.ScheduledJob{
//		{Cronspec: "0 0 * * *", Task: cleanupTask, Description: "Daily cleanup"},
//		{Cronspec: "0 * * * *", Task: healthTask, Description: "Hourly health check"},
//	}
//	entryIDs, err := patterns.RegisterScheduledJobs(scheduler, jobs, logger)
func RegisterScheduledJobs(scheduler *asynq.Scheduler, jobs []ScheduledJob, logger *zap.Logger) ([]string, error) {
	var entryIDs []string

	for _, job := range jobs {
		entryID, err := scheduler.Register(job.Cronspec, job.Task, job.Opts...)
		if err != nil {
			return entryIDs, fmt.Errorf("register job %s (%s): %w", job.Task.Type(), job.Description, err)
		}
		entryIDs = append(entryIDs, entryID)

		logger.Info("registered scheduled job",
			zap.String("entry_id", entryID),
			zap.String("task_type", job.Task.Type()),
			zap.String("cronspec", job.Cronspec),
			zap.String("description", job.Description),
		)
	}

	return entryIDs, nil
}

// ValidateCronspec validates a cron expression by attempting to parse it.
// This is useful for validating user-provided cron expressions before registration.
//
// Returns nil if the cronspec is valid, error otherwise.
func ValidateCronspec(cronspec string) error {
	// Use robfig/cron parser directly (same parser asynq uses internally)
	// This avoids creating a real scheduler/Redis connection for validation
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronspec)
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", cronspec, err)
	}
	return nil
}
