// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event for publishing to message brokers.
type Event struct {
	// ID is the unique identifier for the event.
	ID string `json:"id"`

	// Type is the event type (e.g., "user.created", "order.completed").
	Type string `json:"type"`

	// Payload is the event data as JSON.
	Payload json.RawMessage `json:"payload"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}

// NewEvent creates a new Event with generated ID and current timestamp.
func NewEvent(eventType string, payload interface{}) (Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   data,
		Timestamp: time.Now().UTC(),
	}, nil
}

// EventPublisher defines event publishing abstraction for swappable implementations.
// Implement this interface for Kafka, RabbitMQ, NATS, or other message brokers.
//
// Usage Example:
//
//	// Create and publish an event
//	event, _ := runtimeutil.NewEvent("user.created", map[string]string{"user_id": "123"})
//	publisher.Publish(ctx, "users", event)
//
//	// Async publish (fire and forget)
//	publisher.PublishAsync(ctx, "notifications", event)
//
// Implementing Kafka Publisher:
//
//	type KafkaPublisher struct {
//	    producer *kafka.Producer
//	}
//
//	func (p *KafkaPublisher) Publish(ctx context.Context, topic string, event Event) error {
//	    data, _ := json.Marshal(event)
//	    return p.producer.Produce(ctx, topic, data)
//	}
type EventPublisher interface {
	// Publish sends an event synchronously and waits for confirmation.
	// Returns error if the event could not be published.
	Publish(ctx context.Context, topic string, event Event) error

	// PublishAsync sends an event asynchronously.
	// Returns immediately, event is sent in background.
	// Errors are logged but not returned.
	PublishAsync(ctx context.Context, topic string, event Event) error
}

// NopEventPublisher is a no-op event publisher for testing.
// All events are discarded silently.
type NopEventPublisher struct{}

// NewNopEventPublisher creates a new NopEventPublisher.
func NewNopEventPublisher() EventPublisher {
	return &NopEventPublisher{}
}

// Publish is a no-op and always returns nil.
func (p *NopEventPublisher) Publish(_ context.Context, _ string, _ Event) error {
	return nil
}

// PublishAsync is a no-op and always returns nil.
func (p *NopEventPublisher) PublishAsync(_ context.Context, _ string, _ Event) error {
	return nil
}

// -----------------------------------------------------------------------------
// Event Consumer Interface (Story 13.3)
// -----------------------------------------------------------------------------

// Sentinel errors for event consumer operations.
var (
	// ErrConsumerClosed is returned when operations are attempted on a closed consumer.
	ErrConsumerClosed = errors.New("consumer closed")

	// ErrProcessingTimeout is returned when event processing exceeds the configured timeout.
	ErrProcessingTimeout = errors.New("processing timeout exceeded")

	// ErrMaxRetriesExceeded is returned when all retry attempts are exhausted.
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)

// EventHandler processes a consumed event.
// Return nil for success, error to signal processing failure.
// The handler should be idempotent as events may be redelivered on failure.
//
// Example:
//
//	handler := func(ctx context.Context, event Event) error {
//	    log.Info("processing event", "id", event.ID, "type", event.Type)
//	    // Process event...
//	    return nil
//	}
type EventHandler func(ctx context.Context, event Event) error

// EventConsumer defines event consumption abstraction for swappable implementations.
// Implement this interface for Kafka, RabbitMQ, NATS, or other message brokers.
//
// Usage Example:
//
//	// Create handler function
//	handler := func(ctx context.Context, event Event) error {
//	    log.Info("received event", "id", event.ID, "type", event.Type)
//	    // Process event...
//	    return nil
//	}
//
//	// Subscribe to topic
//	err := consumer.Subscribe(ctx, "orders", handler)
//
// Implementing Kafka Consumer:
//
//	type KafkaConsumer struct {
//	    consumerGroup sarama.ConsumerGroup
//	}
//
//	func (c *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
//	    return c.consumerGroup.Consume(ctx, []string{topic}, &consumerHandler{handler})
//	}
//
// Metrics (implemented by concrete consumers):
//
//	event_consumer_messages_total{topic, status}    - Counter: Total messages consumed
//	event_consumer_errors_total{topic, error_type}  - Counter: Failed processing attempts
//	event_consumer_duration_seconds{topic}          - Histogram: Processing latency
//	event_consumer_retry_total{topic}               - Counter: Retry attempts
//	event_consumer_lag{topic, partition}            - Gauge: Consumer lag (broker-specific)
type EventConsumer interface {
	// Subscribe starts consuming events from the specified topic.
	// The handler is called for each received event.
	// This method blocks until ctx is cancelled or an error occurs.
	// Returns error if subscription fails or consumer is closed.
	Subscribe(ctx context.Context, topic string, handler EventHandler) error

	// Close gracefully stops the consumer.
	// In-flight message processing completes before return.
	// Returns error if cleanup fails.
	Close() error
}

// ConsumerConfig provides options for consumer behavior.
// Use DefaultConsumerConfig() to get sensible defaults.
type ConsumerConfig struct {
	// GroupID is the consumer group identifier (Kafka/RabbitMQ).
	// Required for broker implementations.
	GroupID string

	// MaxRetries is the number of retry attempts for failed processing.
	// Default: 3
	MaxRetries int

	// Concurrency is the number of concurrent message handlers.
	// Default: 1 (sequential processing)
	Concurrency int

	// ProcessingTimeout is the maximum time for handling a single event.
	// Default: 30 seconds
	ProcessingTimeout time.Duration

	// AutoAck automatically acknowledges messages after successful handling.
	// If false, handlers must explicitly acknowledge.
	// Default: true
	AutoAck bool
}

// DefaultConsumerConfig returns sensible defaults for consumer configuration.
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		MaxRetries:        3,
		Concurrency:       1,
		ProcessingTimeout: 30 * time.Second,
		AutoAck:           true,
	}
}

// Validate checks the ConsumerConfig for invalid values.
// Returns error if any field has an invalid value.
func (c ConsumerConfig) Validate() error {
	if c.MaxRetries < 0 {
		return errors.New("MaxRetries must be >= 0")
	}
	if c.Concurrency < 1 {
		return errors.New("Concurrency must be >= 1")
	}
	if c.ProcessingTimeout < 0 {
		return errors.New("ProcessingTimeout must be >= 0")
	}
	return nil
}

// -----------------------------------------------------------------------------
// NopEventConsumer - No-op implementation for testing
// -----------------------------------------------------------------------------

// NopEventConsumer is a no-op event consumer for testing.
// Subscribe returns immediately without blocking.
// All operations are no-ops.
type NopEventConsumer struct {
	mu     sync.Mutex
	closed bool
}

// NewNopEventConsumer creates a new NopEventConsumer.
func NewNopEventConsumer() EventConsumer {
	return &NopEventConsumer{}
}

// Subscribe is a no-op and returns nil immediately.
// It does NOT block, unlike real consumer implementations.
// Returns ErrConsumerClosed if consumer has been closed.
func (c *NopEventConsumer) Subscribe(_ context.Context, _ string, _ EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrConsumerClosed
	}
	return nil
}

// Close marks the consumer as closed.
// Subsequent Subscribe calls will return ErrConsumerClosed.
func (c *NopEventConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

// -----------------------------------------------------------------------------
// MockEventConsumer - Test double for behavior verification
// -----------------------------------------------------------------------------

// MockEventConsumer is a test double that captures handler calls.
// Use SimulateEvent to trigger the registered handler with test events.
type MockEventConsumer struct {
	handler    EventHandler
	topic      string
	events     []Event
	mu         sync.Mutex
	closed     bool
	cancelFunc context.CancelFunc
}

// NewMockEventConsumer creates a new MockEventConsumer.
func NewMockEventConsumer() *MockEventConsumer {
	return &MockEventConsumer{
		events: make([]Event, 0),
	}
}

// Subscribe stores the handler for later simulation.
// Blocks until ctx is cancelled or Close is called.
func (m *MockEventConsumer) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return ErrConsumerClosed
	}
	m.handler = handler
	m.topic = topic
	ctx, m.cancelFunc = context.WithCancel(ctx)
	m.mu.Unlock()

	<-ctx.Done() // Block until cancelled
	return nil
}

// Close marks consumer as closed and cancels any blocking Subscribe.
func (m *MockEventConsumer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	return nil
}

// SimulateEvent triggers the handler with a test event.
// Returns error if no handler is subscribed or handler returns error.
func (m *MockEventConsumer) SimulateEvent(event Event) error {
	m.mu.Lock()
	if m.handler == nil {
		m.mu.Unlock()
		return errors.New("no handler subscribed")
	}
	handler := m.handler
	m.events = append(m.events, event)
	m.mu.Unlock()

	return handler(context.Background(), event)
}

// HandlerCalled returns true if the handler was called at least once.
func (m *MockEventConsumer) HandlerCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events) > 0
}

// LastEvent returns the last event passed to the handler.
// Returns zero Event if no events have been processed.
func (m *MockEventConsumer) LastEvent() Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.events) == 0 {
		return Event{}
	}
	return m.events[len(m.events)-1]
}

// Events returns all events that have been simulated.
func (m *MockEventConsumer) Events() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]Event, len(m.events))
	copy(result, m.events)
	return result
}

// Topic returns the topic that was subscribed to.
func (m *MockEventConsumer) Topic() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.topic
}

// -----------------------------------------------------------------------------
// Dead Letter Queue Interface (Story 13.4)
// -----------------------------------------------------------------------------

// DLQEvent represents a failed event moved to the dead letter queue.
type DLQEvent struct {
	// OriginalEvent is the event that failed processing.
	OriginalEvent Event `json:"original_event"`

	// ErrorMessage is the formatted error from the last failure.
	ErrorMessage string `json:"error_message"`

	// RetryCount is the number of processing attempts made.
	RetryCount int `json:"retry_count"`

	// FailedAt is when the event was moved to DLQ.
	FailedAt time.Time `json:"failed_at"`

	// SourceTopic is the original topic/queue the event came from.
	SourceTopic string `json:"source_topic"`

	// StackTrace is optional debug information (if enabled).
	StackTrace string `json:"stack_trace,omitempty"`

	// Headers contains optional custom metadata.
	Headers map[string]string `json:"headers,omitempty"`
}

// DeadLetterQueue defines the interface for dead letter queue operations.
// Implement this interface for Kafka, RabbitMQ, or other message brokers.
//
// Usage Example:
//
//	// Configure DLQ
//	cfg := runtimeutil.DefaultDLQConfig()
//	cfg.TopicName = "orders.dlq"
//
//	// Wrap handler with DLQ support
//	handler := runtimeutil.NewDLQHandler(myHandler, dlq, cfg)
//	consumer.Subscribe(ctx, "orders", handler)
//
// Metrics (implemented by concrete DLQ implementations):
//
//	event_dlq_total{source_topic, error_type}    - Counter: Events moved to DLQ
//	event_dlq_errors_total{source_topic}         - Counter: DLQ write failures
//	event_dlq_queue_size{dlq_topic}              - Gauge: Current DLQ depth
//	event_dlq_processing_attempts{source_topic}  - Histogram: Retry attempts before DLQ
type DeadLetterQueue interface {
	// Send moves a failed event to the dead letter queue.
	// Returns error if the DLQ write fails.
	Send(ctx context.Context, event DLQEvent) error

	// Close gracefully shuts down the DLQ connection.
	Close() error
}

// DLQMetrics defines the interface for recording DLQ metrics.
type DLQMetrics interface {
	// IncDLQTotal increments the counter for events moved to DLQ.
	IncDLQTotal(topic, errType string)

	// IncDLQErrors increments the counter for failed DLQ writes.
	IncDLQErrors(topic string)
}

// NopDLQMetrics is a no-op implementation of DLQMetrics.
type NopDLQMetrics struct{}

func (m NopDLQMetrics) IncDLQTotal(_, _ string) {}
func (m NopDLQMetrics) IncDLQErrors(_ string)   {}

// DLQConfig provides options for DLQ behavior.
type DLQConfig struct {
	// TopicName is the DLQ topic/queue name.
	// If empty, defaults to "{source-topic}.dlq"
	TopicName string

	// AlertThreshold is the number of events before alerting.
	// Default: 100
	AlertThreshold int

	// IncludeStackTrace includes stack trace in DLQEvent.
	// Default: false (for privacy/size)
	IncludeStackTrace bool

	// RetryDelay is the wait time between retries.
	// Default: 1 second
	RetryDelay time.Duration
}

// DefaultDLQConfig returns sensible defaults.
func DefaultDLQConfig() DLQConfig {
	return DLQConfig{
		AlertThreshold:    100,
		IncludeStackTrace: false,
		RetryDelay:        1 * time.Second,
	}
}

// Validate checks the DLQConfig for invalid values.
func (c DLQConfig) Validate() error {
	if c.AlertThreshold < 0 {
		return errors.New("AlertThreshold must be >= 0")
	}
	if c.RetryDelay < 0 {
		return errors.New("RetryDelay must be >= 0")
	}
	return nil
}

// Sentinel errors for DLQ operations.
var (
	// ErrDLQClosed is returned when operations are attempted on a closed DLQ.
	ErrDLQClosed = errors.New("dlq closed")

	// ErrDLQFull is returned when the DLQ has reached capacity.
	ErrDLQFull = errors.New("dlq full")
)

// -----------------------------------------------------------------------------
// DLQHandler Wrapper
// -----------------------------------------------------------------------------

// DLQHandler wraps an EventHandler with retry and DLQ logic.
type DLQHandler struct {
	handler      EventHandler
	dlq          DeadLetterQueue
	maxRetries   int
	retryDelay   time.Duration
	dlqTopic     string
	includeStack bool
	metrics      DLQMetrics
}

// NewDLQHandler creates a handler that retries failures and forwards to DLQ.
//
// CAUTION: The double-retry risk exists if the underlying consumer (e.g., Kafka/RabbitMQ)
// also has a retry mechanism configured. If the consumer retries on error return, and
// DLQHandler also retries, you may get N*M executions.
// RECOMMENDATION: Set consumer's MaxRetries to 0 or 1 when using DLQHandler,
// or ensure this handler returns nil to the consumer (which it does on success or DLQ success).
func NewDLQHandler(handler EventHandler, dlq DeadLetterQueue, config DLQConfig, consumerConfig ConsumerConfig, metrics DLQMetrics) EventHandler {
	if metrics == nil {
		metrics = NopDLQMetrics{}
	}
	h := &DLQHandler{
		handler:      handler,
		dlq:          dlq,
		maxRetries:   consumerConfig.MaxRetries,
		retryDelay:   config.RetryDelay,
		dlqTopic:     config.TopicName,
		includeStack: config.IncludeStackTrace,
		metrics:      metrics,
	}
	return h.Handle
}

// Handle processes the event with retry logic.
func (h *DLQHandler) Handle(ctx context.Context, event Event) error {
	var lastErr error
	for attempt := 0; attempt <= h.maxRetries; attempt++ {
		err := h.handler(ctx, event)
		if err == nil {
			return nil // Success
		}
		lastErr = err

		// If we have retries left, wait and retry
		if attempt < h.maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(h.retryDelay):
				// continue to next attempt
			}
		}
	}

	// Max retries exhausted, send to DLQ
	dlqEvent := DLQEvent{
		OriginalEvent: event,
		ErrorMessage:  lastErr.Error(),
		RetryCount:    h.maxRetries + 1,
		FailedAt:      time.Now().UTC(),
		SourceTopic:   h.dlqTopic,
	}

	if h.includeStack {
		dlqEvent.StackTrace = string(debug.Stack())
	}

	// Send to DLQ
	if err := h.dlq.Send(ctx, dlqEvent); err != nil {
		h.metrics.IncDLQErrors(h.dlqTopic)
		// If DLQ send fails, we must return error so the original message isn't lost/committed
		return fmt.Errorf("failed to send to DLQ: %w", err)
	}

	h.metrics.IncDLQTotal(h.dlqTopic, "processing_failed")
	// Successfully sent to DLQ, return nil to ack original message
	return nil
}

// -----------------------------------------------------------------------------
// NopDeadLetterQueue
// -----------------------------------------------------------------------------

// NopDeadLetterQueue is a no-op DLQ for testing.
type NopDeadLetterQueue struct {
	mu     sync.Mutex
	closed bool
}

// NewNopDeadLetterQueue creates a new NopDeadLetterQueue.
func NewNopDeadLetterQueue() DeadLetterQueue {
	return &NopDeadLetterQueue{}
}

// Send is a no-op.
func (q *NopDeadLetterQueue) Send(_ context.Context, _ DLQEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return ErrDLQClosed
	}
	return nil
}

// Close is a no-op.
func (q *NopDeadLetterQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	return nil
}

// -----------------------------------------------------------------------------
// MockDeadLetterQueue
// -----------------------------------------------------------------------------

// MockDeadLetterQueue is a test double for DLQ.
type MockDeadLetterQueue struct {
	events []DLQEvent
	mu     sync.Mutex
	closed bool
}

// NewMockDeadLetterQueue creates a new MockDeadLetterQueue.
func NewMockDeadLetterQueue() *MockDeadLetterQueue {
	return &MockDeadLetterQueue{
		events: make([]DLQEvent, 0),
	}
}

// Send captures the event.
func (m *MockDeadLetterQueue) Send(_ context.Context, event DLQEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrDLQClosed
	}
	m.events = append(m.events, event)
	return nil
}

// Close marks the DLQ as closed.
func (m *MockDeadLetterQueue) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// Events returns captured DLQ events.
func (m *MockDeadLetterQueue) Events() []DLQEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]DLQEvent, len(m.events))
	copy(result, m.events)
	return result
}

// LastEvent returns the last event sent to DLQ.
func (m *MockDeadLetterQueue) LastEvent() DLQEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.events) == 0 {
		return DLQEvent{}
	}
	return m.events[len(m.events)-1]
}

// Clear resets captured events.
func (m *MockDeadLetterQueue) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = []DLQEvent{}
}
