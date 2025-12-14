# Story 13.3: Create Event Consumer Interface

Status: done

## Story

As a developer,
I want a standard interface for consuming events,
So that consumer implementations are swappable and testable.

## Acceptance Criteria

1. **Given** `EventConsumer` interface defined in `internal/runtimeutil/events.go`
   **When** implementing a new consumer (Kafka, RabbitMQ, etc.)
   **Then** I only need to implement the standard interface methods
   **And** implementations are interchangeable without code changes

2. **Given** `EventHandler` function type defined
   **When** consuming events from a topic/queue
   **Then** each event is passed to the handler function
   **And** handler returns error to signal processing failure

3. **Given** `ConsumerConfig` options for consumer behavior
   **When** initializing a consumer
   **Then** retry count, concurrency, and timeout are configurable
   **And** defaults are provided for common use cases

4. **Given** any event is consumed
   **When** processing completes or fails
   **Then** structured logs include: topic/queue, event_id, event_type, processing_duration, success/error
   **And** Prometheus metrics track: `event_consumer_messages_total`, `event_consumer_errors_total`, `event_consumer_duration_seconds`

5. **Given** consumer implements graceful shutdown
   **When** `Close()` is called
   **Then** in-flight messages complete processing before shutdown
   **And** consumer unsubscribes from topic/queue

6. **Given** `NopEventConsumer` for testing
   **When** used in unit tests
   **Then** no actual message consumption occurs
   **And** test can verify consumer behavior without broker

## Tasks / Subtasks

- [x] Task 1: Define EventConsumer interface (AC: #1, #2)
  - [x] Create `EventConsumer` interface in `internal/runtimeutil/events.go`
  - [x] Define `Subscribe(ctx, topic, handler) error` method
  - [x] Define `Close() error` method for graceful shutdown
  - [x] Create `EventHandler` function type: `func(ctx, Event) error`

- [x] Task 2: Define ConsumerConfig struct (AC: #3)
  - [x] Create `ConsumerConfig` struct with common options
  - [x] Add `MaxRetries int` (default: 3)
  - [x] Add `Concurrency int` (default: 1 - sequential processing)
  - [x] Add `ProcessingTimeout time.Duration` (default: 30s)
  - [x] Add `GroupID string` (consumer group for Kafka)
  - [x] Add `AutoAck bool` (auto-acknowledge after handler success)

- [x] Task 3: Create NopEventConsumer (AC: #6)
  - [x] Implement `NopEventConsumer` that does nothing
  - [x] Add `NewNopEventConsumer()` factory function
  - [x] Subscribe returns nil immediately (no blocking)
  - [x] Close returns nil immediately

- [x] Task 4: Create MockEventConsumer for testing (AC: #6)
  - [x] Create `MockEventConsumer` with event capture
  - [x] Add `SimulateEvent(event)` to trigger handler
  - [x] Add `HandlerCalled() bool` for test assertions
  - [x] Add `LastEvent() Event` for verification

- [x] Task 5: Define consumer metrics (AC: #4)
  - [x] Define metric names in interface documentation
  - [x] `event_consumer_messages_total{topic, status}` - Counter
  - [x] `event_consumer_errors_total{topic, error_type}` - Counter
  - [x] `event_consumer_duration_seconds{topic}` - Histogram

- [x] Task 6: Define consumer errors (AC: #1, #2)
  - [x] Create `ErrConsumerClosed` sentinel error
  - [x] Create `ErrProcessingTimeout` sentinel error
  - [x] Create `ErrMaxRetriesExceeded` sentinel error
  - [x] Document error handling patterns in godoc

- [x] Task 7: Write unit tests for interface contract
  - [x] Create `internal/runtimeutil/events_test.go` (extend existing)
  - [x] Test `NopEventConsumer` implements interface
  - [x] Test `MockEventConsumer` event simulation
  - [x] Test `ConsumerConfig` defaults
  - [x] Test sentinel errors

- [x] Task 8: Update documentation (AC: all)
  - [x] Update `AGENTS.md` with EventConsumer patterns
  - [x] Update `README.md` with consumer usage examples
  - [x] Update `docs/architecture.md` with consumer section
  - [x] Add godoc examples for interface usage

## Dev Notes

### Architecture Compliance

**Location:** `internal/runtimeutil/events.go` - This is the PORT definition for event consumption.

**Layer Boundaries:**
- `EventConsumer` interface (port) is in `runtimeutil` package
- Future implementations (KafkaConsumer, RabbitMQConsumer) go in `infra/kafka/consumer.go`, `infra/rabbitmq/consumer.go`
- Usecase layer calls consumer via interface, NOT concrete implementation

### Existing EventPublisher Pattern (Follow This)

From `internal/runtimeutil/events.go`:

```go
// EventPublisher defines event publishing abstraction for swappable implementations.
type EventPublisher interface {
    Publish(ctx context.Context, topic string, event Event) error
    PublishAsync(ctx context.Context, topic string, event Event) error
}

// NopEventPublisher is a no-op event publisher for testing.
type NopEventPublisher struct{}
```

**Apply same pattern for EventConsumer:**
- Interface with clear contract
- NopConsumer for testing
- MockConsumer for behavior verification

### Proposed EventConsumer Interface

```go
// EventHandler processes a consumed event.
// Return nil for success, error to signal processing failure.
type EventHandler func(ctx context.Context, event Event) error

// EventConsumer defines event consumption abstraction for swappable implementations.
// Implement this interface for Kafka, RabbitMQ, NATS, or other message brokers.
//
// Usage Example:
//
//  // Create handler function
//  handler := func(ctx context.Context, event Event) error {
//      log.Info("received event", "id", event.ID, "type", event.Type)
//      // Process event...
//      return nil
//  }
//
//  // Subscribe to topic
//  err := consumer.Subscribe(ctx, "orders", handler)
//
// Implementing Kafka Consumer:
//
//  type KafkaConsumer struct {
//      consumerGroup sarama.ConsumerGroup
//  }
//
//  func (c *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
//      return c.consumerGroup.Consume(ctx, []string{topic}, &consumerHandler{handler})
//  }
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
```

### ConsumerConfig Structure

```go
// ConsumerConfig provides options for consumer behavior.
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

// DefaultConsumerConfig returns sensible defaults.
func DefaultConsumerConfig() ConsumerConfig {
    return ConsumerConfig{
        MaxRetries:        3,
        Concurrency:       1,
        ProcessingTimeout: 30 * time.Second,
        AutoAck:           true,
    }
}
```

### MockEventConsumer for Testing

```go
// MockEventConsumer is a test double that captures handler calls.
type MockEventConsumer struct {
    handler     EventHandler
    events      []Event
    mu          sync.Mutex
    closed      bool
}

// Subscribe stores the handler for later simulation.
func (m *MockEventConsumer) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
    m.handler = handler
    <-ctx.Done() // Block until cancelled
    return nil
}

// SimulateEvent triggers the handler with a fake event.
func (m *MockEventConsumer) SimulateEvent(event Event) error {
    if m.handler == nil {
        return errors.New("no handler subscribed")
    }
    m.mu.Lock()
    m.events = append(m.events, event)
    m.mu.Unlock()
    return m.handler(context.Background(), event)
}

// Close marks consumer as closed.
func (m *MockEventConsumer) Close() error {
    m.closed = true
    return nil
}
```

### Sentinel Errors

```go
var (
    // ErrConsumerClosed is returned when operations are attempted on a closed consumer.
    ErrConsumerClosed = errors.New("consumer closed")

    // ErrProcessingTimeout is returned when event processing exceeds the configured timeout.
    ErrProcessingTimeout = errors.New("processing timeout exceeded")

    // ErrMaxRetriesExceeded is returned when all retry attempts are exhausted.
    ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)
```

### Metrics to Define (for future implementations)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `event_consumer_messages_total` | Counter | `topic`, `status` | Total messages consumed |
| `event_consumer_errors_total` | Counter | `topic`, `error_type` | Failed processing attempts |
| `event_consumer_duration_seconds` | Histogram | `topic` | Processing latency |
| `event_consumer_retry_total` | Counter | `topic` | Retry attempts |
| `event_consumer_lag` | Gauge | `topic`, `partition` | Consumer lag (broker-specific) |

### Previous Story Learnings (Story 13.1 & 13.2)

- ✅ Used `Close()` for graceful shutdown with `defer` in main.go
- ✅ Implemented both sync and async patterns
- ✅ Added comprehensive Prometheus metrics with labels
- ✅ Created `NopEventPublisher` for testing
- ✅ Used table-driven tests with AAA pattern
- ✅ Added health check capability

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| `Subscribe` blocks until ctx cancelled | Matches Kafka ConsumerGroup and RabbitMQ Consume patterns |
| `EventHandler` returns error | Allows retry logic in consumer implementation |
| `ConsumerConfig` separate from interface | Implementations may extend with broker-specific options |
| `MockEventConsumer` for simulation | Enables handler testing without broker |
| No built-in retry in interface | Retry logic is implementation-specific (Kafka offset, RabbitMQ nack) |

### Future Implementations (Not in this story)

| Story | Implementation |
|-------|----------------|
| 13.4 (DLQ) | Uses `EventConsumer` for failed event routing |
| Future | `KafkaConsumer` in `internal/infra/kafka/consumer.go` |
| Future | `RabbitMQConsumer` in `internal/infra/rabbitmq/consumer.go` |

### Project Structure Notes

Files to create/modify:
```
internal/
├── runtimeutil/
│   ├── events.go              # MODIFY: Add EventConsumer interface
│   └── events_test.go         # MODIFY: Add consumer tests
docs/
├── architecture.md            # MODIFY: Add consumer section
AGENTS.md                      # MODIFY: Add consumer patterns
README.md                      # MODIFY: Add consumer examples
```

### References

- [Source: docs/epics.md#Story-13.3]
- [Source: docs/architecture.md#Extension-Interfaces]
- [Source: internal/runtimeutil/events.go] - EventPublisher pattern
- [Source: internal/infra/kafka/publisher.go] - Implementation reference
- [Source: internal/infra/rabbitmq/publisher.go] - Implementation reference

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

### Completion Notes List

- ✅ Implemented `EventConsumer` interface with `Subscribe` and `Close` methods
- ✅ Created `EventHandler` function type for event processing
- ✅ Defined `ConsumerConfig` struct with sensible defaults via `DefaultConsumerConfig()`
- ✅ Implemented `NopEventConsumer` for testing (non-blocking, no-op)
- ✅ Implemented `MockEventConsumer` with full simulation API (`SimulateEvent`, `HandlerCalled`, `LastEvent`, `Events`, `Topic`)
- ✅ Created 3 sentinel errors: `ErrConsumerClosed`, `ErrProcessingTimeout`, `ErrMaxRetriesExceeded`
- ✅ Added comprehensive documentation in godoc with usage examples
- ✅ Created 25 unit tests for all components
- ✅ Updated `AGENTS.md` with EventConsumer section (130+ lines)
- ✅ Updated `README.md` with Event Consuming section
- ✅ Updated `docs/architecture.md` with EventConsumer interface documentation

### File List

**New Files:**
- `internal/runtimeutil/events_test.go` - 25 unit tests for events.go

**Modified Files:**
- `internal/runtimeutil/events.go` - Added EventConsumer interface, EventHandler, ConsumerConfig, NopEventConsumer, MockEventConsumer, sentinel errors (240+ lines)
- `AGENTS.md` - Added EventConsumer Interface section
- `README.md` - Added Event Consuming (V3) section
- `docs/architecture.md` - Added EventConsumer Interface section
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status

## Senior Developer Review (AI)

**Reviewer:** Code Review Workflow (2025-12-14)
**Outcome:** ✅ APPROVED with fixes applied

### Issues Found and Fixed

| Severity | Issue | Resolution |
|----------|-------|------------|
| MEDIUM | `NopEventConsumer.Subscribe` allowed calls after `Close()` | Added `closed` state tracking; returns `ErrConsumerClosed` after close |
| LOW | `ConsumerConfig` lacked validation | Added `Validate()` method to check field constraints |
| LOW | Error wrapping test used string concat instead of `%w` | Fixed to use `fmt.Errorf(..., %w)` with `errors.Is` |
| LOW | Test coverage for NopConsumer state | Added `TestNopEventConsumer_SubscribeAfterClose` |

### Files Modified in Review

- `internal/runtimeutil/events.go`
  - `NopEventConsumer` now tracks `closed` state with mutex
  - Added `ConsumerConfig.Validate()` method
- `internal/runtimeutil/events_test.go`
  - Added `TestNopEventConsumer_SubscribeAfterClose`
  - Added `TestConsumerConfig_Validate` (5 sub-tests)
  - Fixed `TestSentinelErrorsCanBeWrapped` to use proper error wrapping

### Test Results

✅ All 29+ tests pass (`go test -v ./internal/runtimeutil/...`)

### Change Log

| Date | Action | Details |
|------|--------|---------|
| 2025-12-14 | Code Review | Fixed 1 Medium, 3 Low issues; all tests pass |

