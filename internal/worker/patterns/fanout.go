// Package patterns provides reusable async job patterns for the worker infrastructure.
// This file contains the Fanout pattern for event-driven workflows.
//
// Fanout Pattern:
//
// The fanout pattern is used for broadcasting events to multiple handlers:
//   - User created → SendWelcomeEmail, CreateDefaultSettings, NotifyAdmin
//   - Order completed → SendReceipt, UpdateInventory, NotifyShipping
//   - Note archived → UpdateSearchIndex, NotifyOwner, CreateBackup
//
// Each handler is completely isolated:
//   - Separate Task: Each handler gets its own asynq task
//   - Independent Retry: Each task retries according to its own configuration
//   - Separate Queue: Handlers can run on different queues
//   - Independent Failure: One handler failure doesn't affect others
//
// Architecture:
//   - ENQUEUE SIDE (API): FanoutRegistry + Fanout() enqueues tasks to Redis
//   - PROCESS SIDE (Worker): FanoutDispatcher routes tasks to handlers
//
// Example:
//
//	registry := patterns.NewFanoutRegistry()
//	registry.Register("user:created", "welcome-email", sendWelcomeEmail)
//	registry.Register("user:created", "default-settings", createDefaultSettings)
//
//	event := patterns.FanoutEvent{Type: "user:created", Payload: userJSON}
//	patterns.Fanout(ctx, enqueuer, registry, logger, event)
package patterns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/iruldev/golang-api-hexagonal/internal/worker"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
	"go.uber.org/zap"
)

// FanoutEvent represents an event to be broadcast to multiple handlers.
type FanoutEvent struct {
	// Type is the event type identifier (e.g., "user:created", "order:completed")
	Type string `json:"type"`

	// Payload is the event data as raw JSON
	Payload json.RawMessage `json:"payload"`

	// Metadata contains optional event metadata (timestamp, source, correlation_id)
	Metadata map[string]string `json:"metadata,omitempty"`

	// Timestamp is when the event was created (auto-set if zero)
	Timestamp time.Time `json:"timestamp"`
}

// FanoutHandlerFunc is the signature for fanout event handlers.
type FanoutHandlerFunc func(ctx context.Context, event FanoutEvent) error

// FanoutHandler represents a registered handler for fanout events.
type FanoutHandler struct {
	ID      string            // Unique handler identifier
	Handler FanoutHandlerFunc // Handler function
	Queue   string            // Target queue (default: "default")
	Opts    []asynq.Option    // Additional asynq options
}

// FanoutRegistry manages registration of fanout handlers.
// It is thread-safe and can be used concurrently.
type FanoutRegistry struct {
	mu       sync.RWMutex
	handlers map[string][]FanoutHandler // eventType -> handlers
}

// NewFanoutRegistry creates a new fanout registry.
func NewFanoutRegistry() *FanoutRegistry {
	return &FanoutRegistry{
		handlers: make(map[string][]FanoutHandler),
	}
}

// ErrEmptyEventType is returned when eventType is empty.
var ErrEmptyEventType = errors.New("eventType cannot be empty")

// ErrEmptyHandlerID is returned when handlerID is empty.
var ErrEmptyHandlerID = errors.New("handlerID cannot be empty")

// ErrNilHandler is returned when handler function is nil.
var ErrNilHandler = errors.New("handler function cannot be nil")

// ErrDuplicateHandler is returned when a handler with the same ID already exists.
var ErrDuplicateHandler = errors.New("handler with this ID already registered for event type")

// Register adds a handler for an event type.
// handlerID must be unique within the event type.
// Returns error if validation fails or handler already exists.
func (r *FanoutRegistry) Register(eventType, handlerID string, fn FanoutHandlerFunc, opts ...asynq.Option) error {
	if eventType == "" {
		return ErrEmptyEventType
	}
	if handlerID == "" {
		return ErrEmptyHandlerID
	}
	if fn == nil {
		return ErrNilHandler
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate handlerID
	for _, h := range r.handlers[eventType] {
		if h.ID == handlerID {
			return ErrDuplicateHandler
		}
	}

	handler := FanoutHandler{
		ID:      handlerID,
		Handler: fn,
		Queue:   worker.QueueDefault, // Default queue
		Opts:    opts,
	}
	r.handlers[eventType] = append(r.handlers[eventType], handler)
	return nil
}

// RegisterWithQueue adds a handler with a specific queue for an event type.
// Returns error if validation fails or handler already exists.
func (r *FanoutRegistry) RegisterWithQueue(eventType, handlerID string, fn FanoutHandlerFunc, queue string, opts ...asynq.Option) error {
	if eventType == "" {
		return ErrEmptyEventType
	}
	if handlerID == "" {
		return ErrEmptyHandlerID
	}
	if fn == nil {
		return ErrNilHandler
	}
	if queue == "" {
		queue = worker.QueueDefault
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate handlerID
	for _, h := range r.handlers[eventType] {
		if h.ID == handlerID {
			return ErrDuplicateHandler
		}
	}

	handler := FanoutHandler{
		ID:      handlerID,
		Handler: fn,
		Queue:   queue,
		Opts:    opts,
	}
	r.handlers[eventType] = append(r.handlers[eventType], handler)
	return nil
}

// Handlers returns all handlers for an event type.
func (r *FanoutRegistry) Handlers(eventType string) []FanoutHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	handlers := r.handlers[eventType]
	if handlers == nil {
		return nil
	}
	result := make([]FanoutHandler, len(handlers))
	copy(result, handlers)
	return result
}

// Unregister removes a handler for an event type.
func (r *FanoutRegistry) Unregister(eventType, handlerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handlers := r.handlers[eventType]
	for i, h := range handlers {
		if h.ID == handlerID {
			r.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

// Fanout broadcasts an event to all registered handlers.
// Each handler receives a separate task and processes independently.
// Failures in one handler don't affect others.
//
// Returns []error containing any enqueue failures. Callers should check:
//   - len(errors) == 0: All handlers enqueued successfully
//   - len(errors) > 0: Some handlers failed to enqueue (partial success)
//   - len(errors) == len(handlers): Complete failure
//
// Note: Timestamp is auto-set to time.Now().UTC() if zero.
func Fanout(
	ctx context.Context,
	enqueuer tasks.TaskEnqueuer,
	registry *FanoutRegistry,
	logger *zap.Logger,
	event FanoutEvent,
) []error {
	// Auto-set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	handlers := registry.Handlers(event.Type)
	if len(handlers) == 0 {
		logger.Warn("no handlers registered for event",
			zap.String("event_type", event.Type),
		)
		return nil
	}

	var errors []error
	for _, h := range handlers {
		taskType := fmt.Sprintf("fanout:%s:%s", event.Type, h.ID)
		payload, err := json.Marshal(event)
		if err != nil {
			errors = append(errors, fmt.Errorf("marshal event for handler %s: %w", h.ID, err))
			continue
		}

		task := asynq.NewTask(taskType, payload, h.Opts...)
		opts := []asynq.Option{asynq.Queue(h.Queue)}
		opts = append(opts, h.Opts...)

		info, err := enqueuer.Enqueue(ctx, task, opts...)
		if err != nil {
			logger.Error("fanout enqueue failed",
				zap.Error(err),
				zap.String("event_type", event.Type),
				zap.String("handler_id", h.ID),
			)
			errors = append(errors, fmt.Errorf("enqueue handler %s: %w", h.ID, err))
			continue
		}

		logger.Debug("fanout task enqueued",
			zap.String("task_id", info.ID),
			zap.String("event_type", event.Type),
			zap.String("handler_id", h.ID),
			zap.String("queue", info.Queue),
		)
	}

	return errors
}

// FanoutDispatcher handles fanout tasks on the worker side.
// It routes incoming tasks to the correct registered handler.
type FanoutDispatcher struct {
	registry *FanoutRegistry
	logger   *zap.Logger
}

// NewFanoutDispatcher creates a new dispatcher with the given registry.
func NewFanoutDispatcher(registry *FanoutRegistry, logger *zap.Logger) *FanoutDispatcher {
	return &FanoutDispatcher{
		registry: registry,
		logger:   logger,
	}
}

// Handle processes a fanout task by routing to the correct handler.
// Task type format: "fanout:{eventType}:{handlerID}"
func (d *FanoutDispatcher) Handle(ctx context.Context, t *asynq.Task) error {
	taskID, _ := asynq.GetTaskID(ctx)

	// Parse task type: "fanout:{eventType}:{handlerID}"
	// Event type may contain colons (e.g., "user:created"), so we need to:
	// 1. Strip "fanout:" prefix
	// 2. Extract handlerID as the last segment after the final ":"
	// 3. Everything between is the eventType
	taskType := t.Type()
	if !strings.HasPrefix(taskType, "fanout:") {
		d.logger.Error("invalid fanout task type - missing prefix",
			zap.String("task_type", taskType),
			zap.String("task_id", taskID),
		)
		return fmt.Errorf("invalid fanout task type: %s: %w", taskType, asynq.SkipRetry)
	}

	// Remove "fanout:" prefix
	remaining := strings.TrimPrefix(taskType, "fanout:")

	// Find the last colon to split eventType and handlerID
	lastColonIdx := strings.LastIndex(remaining, ":")
	if lastColonIdx == -1 || lastColonIdx == 0 || lastColonIdx == len(remaining)-1 {
		d.logger.Error("invalid fanout task type - cannot parse",
			zap.String("task_type", taskType),
			zap.String("task_id", taskID),
		)
		return fmt.Errorf("invalid fanout task type: %s: %w", taskType, asynq.SkipRetry)
	}

	eventType := remaining[:lastColonIdx]
	handlerID := remaining[lastColonIdx+1:]

	// Find handler
	handlers := d.registry.Handlers(eventType)
	for _, h := range handlers {
		if h.ID == handlerID {
			var event FanoutEvent
			if err := json.Unmarshal(t.Payload(), &event); err != nil {
				d.logger.Error("unmarshal fanout event failed",
					zap.Error(err),
					zap.String("event_type", eventType),
					zap.String("handler_id", handlerID),
					zap.String("task_id", taskID),
				)
				return fmt.Errorf("unmarshal event: %w: %w", err, asynq.SkipRetry)
			}

			d.logger.Debug("processing fanout event",
				zap.String("event_type", eventType),
				zap.String("handler_id", handlerID),
				zap.String("task_id", taskID),
			)

			return h.Handler(ctx, event)
		}
	}

	d.logger.Error("handler not found for fanout event",
		zap.String("event_type", eventType),
		zap.String("handler_id", handlerID),
		zap.String("task_id", taskID),
	)
	return fmt.Errorf("handler %s not found for event %s: %w", handlerID, eventType, asynq.SkipRetry)
}

// TaskTypePrefix returns the prefix for fanout task types.
// Use this for registering with asynq's pattern matching.
func TaskTypePrefix() string {
	return "fanout:"
}
