# Story 6.4: Define EventPublisher Interface

Status: done

## Story

As a developer,
I want an EventPublisher interface abstraction,
So that I can publish events to Kafka, RabbitMQ, or NATS.

## Acceptance Criteria

### AC1: EventPublisher interface defined
**Given** `internal/runtimeutil/events.go` exists
**When** I review the interface
**Then** methods include: Publish(topic, event), PublishAsync(topic, event)
**And** event struct has ID, Type, Payload, Timestamp

---

## Tasks / Subtasks

- [x] **Task 1: Define Event struct** (AC: #1)
  - [x] Create Event struct with ID, Type, Payload, Timestamp
  - [x] Add NewEvent constructor

- [x] **Task 2: Define EventPublisher interface** (AC: #1)
  - [x] Create Publish(ctx, topic, event) method
  - [x] Create PublishAsync(ctx, topic, event) method

- [x] **Task 3: Add NopEventPublisher for testing** (AC: #1)
  - [x] Create no-op implementation

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Event Struct

```go
// Event represents a domain event for publishing.
type Event struct {
    ID        string          // Unique event ID
    Type      string          // Event type (e.g., "user.created")
    Payload   json.RawMessage // Event data
    Timestamp time.Time       // When the event occurred
}
```

### EventPublisher Interface

```go
// EventPublisher defines event publishing abstraction.
type EventPublisher interface {
    // Publish sends an event synchronously.
    Publish(ctx context.Context, topic string, event Event) error

    // PublishAsync sends an event asynchronously.
    // Returns immediately, event is sent in background.
    PublishAsync(ctx context.Context, topic string, event Event) error
}
```

### Architecture Compliance

**Layer:** `internal/runtimeutil/`
**Pattern:** Interface abstraction for message brokers
**Benefit:** Swappable backends (Kafka, RabbitMQ, NATS)

### File List

Files to create:
- `internal/runtimeutil/events.go` - EventPublisher interface
