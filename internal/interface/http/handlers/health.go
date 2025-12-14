// Package handlers contains HTTP request handlers for the API.
package handlers

import (
	"context"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// HealthData represents the health check data.
type HealthData struct {
	Status string `json:"status"`
}

// HealthHandler returns the health status of the service.
// Response format: {"success": true, "data": {"status": "ok"}}
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, HealthData{Status: "ok"})
}

// DBHealthChecker checks database health.
type DBHealthChecker interface {
	Ping(ctx context.Context) error
}

// ReadyzHandler handles readiness probe requests.
type ReadyzHandler struct {
	dbChecker       DBHealthChecker
	redisChecker    DBHealthChecker
	kafkaChecker    DBHealthChecker
	rabbitmqChecker DBHealthChecker
}

// NewReadyzHandler creates a new ReadyzHandler with optional checkers.
func NewReadyzHandler(dbChecker DBHealthChecker) *ReadyzHandler {
	return &ReadyzHandler{dbChecker: dbChecker}
}

// WithRedis adds Redis health checker to the readiness handler.
func (h *ReadyzHandler) WithRedis(redisChecker DBHealthChecker) *ReadyzHandler {
	h.redisChecker = redisChecker
	return h
}

// WithKafka adds Kafka health checker to the readiness handler.
func (h *ReadyzHandler) WithKafka(kafkaChecker DBHealthChecker) *ReadyzHandler {
	h.kafkaChecker = kafkaChecker
	return h
}

// WithRabbitMQ adds RabbitMQ health checker to the readiness handler.
func (h *ReadyzHandler) WithRabbitMQ(rabbitmqChecker DBHealthChecker) *ReadyzHandler {
	h.rabbitmqChecker = rabbitmqChecker
	return h
}

// ServeHTTP handles the readiness check request.
// Returns 200 if service is ready, 503 if any dependency is unavailable.
func (h *ReadyzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database if available
	if h.dbChecker != nil {
		if err := h.dbChecker.Ping(ctx); err != nil {
			response.Error(w, http.StatusServiceUnavailable, response.ErrServiceUnavailable, "database unavailable")
			return
		}
	}

	// Check Redis if available
	if h.redisChecker != nil {
		if err := h.redisChecker.Ping(ctx); err != nil {
			response.Error(w, http.StatusServiceUnavailable, response.ErrServiceUnavailable, "redis unavailable")
			return
		}
	}

	// Check Kafka if available (Story 13.1)
	if h.kafkaChecker != nil {
		if err := h.kafkaChecker.Ping(ctx); err != nil {
			response.Error(w, http.StatusServiceUnavailable, response.ErrServiceUnavailable, "kafka unavailable")
			return
		}
	}

	// Check RabbitMQ if available (Story 13.2)
	if h.rabbitmqChecker != nil {
		if err := h.rabbitmqChecker.Ping(ctx); err != nil {
			response.Error(w, http.StatusServiceUnavailable, response.ErrServiceUnavailable, "rabbitmq unavailable")
			return
		}
	}

	response.Success(w, HealthData{Status: "ready"})
}
