// Package main is the entry point for the background worker service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/worker"
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

	// Create Redis options for asynq
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Create worker server
	srv := worker.NewServer(redisOpt, cfg.Asynq)

	// Add middleware (order: recovery first, then tracing, then logging)
	srv.Use(
		worker.RecoveryMiddleware(logger),
		worker.TracingMiddleware(),
		worker.LoggingMiddleware(logger),
	)

	// Register task handlers
	noteArchiveHandler := tasks.NewNoteArchiveHandler(logger)
	srv.HandleFunc(tasks.TypeNoteArchive, noteArchiveHandler.Handle)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		logger.Info("Shutting down worker...")
		srv.Shutdown()
		cancel()
	}()

	logger.Info("Worker starting",
		zap.Int("concurrency", cfg.Asynq.Concurrency),
		zap.String("redis_addr", fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)),
	)

	if err := srv.Start(); err != nil {
		logger.Fatal("Worker error", zap.Error(err))
	}

	<-ctx.Done()
	logger.Info("Worker shutdown complete")
}
