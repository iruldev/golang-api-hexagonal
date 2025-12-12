// Package worker provides Asynq background job processing infrastructure.
package worker

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

// Queue priority constants.
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

// Server wraps asynq.Server for background task processing.
type Server struct {
	srv *asynq.Server
	mux *asynq.ServeMux
}

// NewServer creates a new Asynq server with the given Redis options and config.
func NewServer(redisOpt asynq.RedisClientOpt, cfg config.AsynqConfig) *Server {
	// Apply defaults
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 10
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: cfg.Concurrency,
		Queues: map[string]int{
			QueueCritical: 6, // weight 6
			QueueDefault:  3, // weight 3
			QueueLow:      1, // weight 1
		},
		ShutdownTimeout: cfg.ShutdownTimeout,
		RetryDelayFunc:  asynq.DefaultRetryDelayFunc,
	})

	return &Server{
		srv: srv,
		mux: asynq.NewServeMux(),
	}
}

// HandleFunc registers a task handler for the given pattern (task type).
func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

// Use adds middleware to the server's mux.
func (s *Server) Use(mws ...asynq.MiddlewareFunc) {
	s.mux.Use(mws...)
}

// Start begins processing tasks. This method blocks until Shutdown is called.
func (s *Server) Start() error {
	return s.srv.Run(s.mux)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	s.srv.Shutdown()
}
