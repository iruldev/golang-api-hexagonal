package patterns_test

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/iruldev/golang-api-hexagonal/internal/worker"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
)

// Example_scheduledJob demonstrates how to create and register scheduled jobs.
func Example_scheduledJob() {
	// Create a task to run on schedule
	task := asynq.NewTask("cleanup:old_notes", nil)

	// Define the scheduled job with cron expression
	job := patterns.ScheduledJob{
		Cronspec:    "0 0 * * *", // Daily at midnight UTC
		Task:        task,
		Description: "Daily cleanup of old archived notes",
		Opts:        []asynq.Option{asynq.Queue(worker.QueueLow)},
	}

	fmt.Printf("Scheduled job: %s at %s\n", job.Task.Type(), job.Cronspec)
	// Output: Scheduled job: cleanup:old_notes at 0 0 * * *
}

// Example_cronExpressions demonstrates common cron expression patterns.
func Example_cronExpressions() {
	expressions := []struct {
		cronspec string
		meaning  string
	}{
		{"* * * * *", "Every minute"},
		{"0 * * * *", "Every hour"},
		{"0 0 * * *", "Daily at midnight"},
		{"0 9 * * 1", "Every Monday at 9am"},
		{"0 0 1 * *", "First of every month"},
		{"*/5 * * * *", "Every 5 minutes"},
		{"0 18 * * 1-5", "Weekdays at 6pm"},
	}

	for _, e := range expressions {
		fmt.Printf("%s: %s\n", e.cronspec, e.meaning)
	}
	// Output:
	// * * * * *: Every minute
	// 0 * * * *: Every hour
	// 0 0 * * *: Daily at midnight
	// 0 9 * * 1: Every Monday at 9am
	// 0 0 1 * *: First of every month
	// */5 * * * *: Every 5 minutes
	// 0 18 * * 1-5: Weekdays at 6pm
}

// Example_multipleScheduledJobs demonstrates registering multiple jobs.
func Example_multipleScheduledJobs() {
	// In a real application, this would be in cmd/scheduler/main.go
	jobs := []patterns.ScheduledJob{
		{
			Cronspec:    "0 0 * * *",
			Task:        asynq.NewTask("cleanup:old_notes", nil),
			Description: "Daily cleanup",
		},
		{
			Cronspec:    "0 * * * *",
			Task:        asynq.NewTask("health:check", nil),
			Description: "Hourly health check",
		},
		{
			Cronspec:    "0 6 * * 1",
			Task:        asynq.NewTask("report:weekly", nil),
			Description: "Weekly report on Monday 6am",
		},
	}

	for _, job := range jobs {
		fmt.Printf("%s (%s)\n", job.Description, job.Cronspec)
	}
	// Output:
	// Daily cleanup (0 0 * * *)
	// Hourly health check (0 * * * *)
	// Weekly report on Monday 6am (0 6 * * 1)
}

// Example_scheduledJobWithOptions demonstrates using task options with scheduled jobs.
func Example_scheduledJobWithOptions() {
	task := asynq.NewTask("critical:scheduled", nil)

	job := patterns.ScheduledJob{
		Cronspec:    "*/5 * * * *", // Every 5 minutes
		Task:        task,
		Description: "Critical scheduled task",
		Opts: []asynq.Option{
			asynq.Queue(worker.QueueCritical), // Use critical queue
			asynq.MaxRetry(5),                 // More retries for critical tasks
		},
	}

	fmt.Printf("Queue: critical, Cronspec: %s\n", job.Cronspec)
	// Output: Queue: critical, Cronspec: */5 * * * *
}
