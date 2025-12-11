// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"encoding/json"
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
