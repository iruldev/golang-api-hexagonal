// Package main is the entry point for the asynq scheduler service.
// The scheduler enqueues periodic tasks based on cron expressions.
// It runs as a separate process from the worker.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
)

func main() {
	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize zap logger
	logger, err := observability.NewLogger(&cfg.Log, cfg.App.Env)
	if err != nil {
		log.Fatalf("Logger error: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Create Redis options for asynq scheduler
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Create scheduler with UTC timezone
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		logger.Fatal("Failed to load UTC timezone", zap.Error(err))
	}

	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Location: loc,
		LogLevel: asynq.InfoLevel,
		// PostEnqueueFunc is called after a task is enqueued by the scheduler.
		PostEnqueueFunc: func(info *asynq.TaskInfo, err error) {
			if err != nil {
				logger.Error("scheduler failed to enqueue task",
					zap.Error(err),
					zap.String("task_type", info.Type),
				)
				return
			}
			logger.Debug("scheduler enqueued task",
				zap.String("task_id", info.ID),
				zap.String("task_type", info.Type),
				zap.String("queue", info.Queue),
			)
		},
	})

	// Register scheduled jobs
	jobs, err := defineScheduledJobs(logger)
	if err != nil {
		logger.Fatal("Failed to define scheduled jobs", zap.Error(err))
	}
	entryIDs, err := patterns.RegisterScheduledJobs(scheduler, jobs, logger)
	if err != nil {
		logger.Fatal("Failed to register scheduled jobs", zap.Error(err))
	}

	logger.Info("Scheduled jobs registered",
		zap.Int("count", len(entryIDs)),
		zap.Strings("entry_ids", entryIDs),
	)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		logger.Info("Shutting down scheduler...")
		scheduler.Shutdown()
		cancel()
	}()

	logger.Info("Scheduler starting",
		zap.String("timezone", loc.String()),
		zap.String("redis_addr", fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)),
	)

	if err := scheduler.Run(); err != nil {
		logger.Fatal("Scheduler error", zap.Error(err))
	}

	<-ctx.Done()
	logger.Info("Scheduler shutdown complete")
}

// defineScheduledJobs returns all scheduled jobs for the application.
// Add new scheduled jobs here following the ScheduledJob pattern.
func defineScheduledJobs(logger *zap.Logger) ([]patterns.ScheduledJob, error) {
	// Create cleanup task (example - runs daily at midnight UTC)
	cleanupTask, err := tasks.NewCleanupOldNotesTask()
	if err != nil {
		return nil, fmt.Errorf("create cleanup task: %w", err)
	}

	return []patterns.ScheduledJob{
		{
			// Daily cleanup at midnight UTC
			Cronspec:    "0 0 * * *",
			Task:        cleanupTask,
			Description: "Daily cleanup of old archived notes",
		},
		// Examples of other schedules:
		// {
		// 	Cronspec:    "0 * * * *",      // Every hour
		// 	Task:        hourlyTask,
		// 	Description: "Hourly health check",
		// },
		// {
		// 	Cronspec:    "*/5 * * * *",    // Every 5 minutes
		// 	Task:        frequentTask,
		// 	Description: "Frequent cache refresh",
		// },
	}, nil
}
