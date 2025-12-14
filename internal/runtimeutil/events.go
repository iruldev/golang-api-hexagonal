// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"encoding/json"
	"errors"
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
