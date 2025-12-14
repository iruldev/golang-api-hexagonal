# Story 13.4: Implement Dead Letter Queue Pattern

Status: done

## Story

As a SRE,
I want failed events to go to a DLQ,
So that no data is lost during processing failures.

## Acceptance Criteria

- [x] **AC #1**: Standard `DeadLetterQueue` interface defined in `events.go` with `Send` and `Close` methods.
- [x] **AC #2**: `DLQEvent` struct includes original event, error, retry count, failure timestamp, and source topic.
- [x] **AC #3**: `DLQHandler` wrapper implements retry logic (N times) and moves to DLQ on exhaustion.
- [x] **AC #4**: Metrics defined for DLQ operations: `event_dlq_total`, `event_dlq_errors_total`, `event_dlq_queue_size`.
- [x] **AC #5**: `MockDeadLetterQueue` available for testing DLQ interactions.
- [x] **AC #6**: Documentation updated (AGENTS.md, README.md) with DLQ usage patterns.

## Tasks / Subtasks

- [x] Task 1: Define DeadLetterQueue interface (AC: #1, #2)
  - [x] Create `DeadLetterQueue` interface in `internal/runtimeutil/events.go`
  - [x] Define `Send(ctx, dlqEvent DLQEvent) error` method
  - [x] Define `Close() error` method for graceful shutdown
  - [x] Create `DLQEvent` struct with original event + failure metadata

- [x] Task 2: Define DLQConfig struct (AC: #3)
  - [x] Create `DLQConfig` struct with common options
  - [x] Add `TopicName string` (default: "{original-topic}.dlq")
  - [x] Add `MaxRetries int` (from ConsumerConfig, used for retry threshold)
  - [x] Add `AlertThreshold int` (default: 100 - events before alerting)
  - [x] Add `IncludeStackTrace bool` (default: false for privacy)
  - [x] Add `DefaultDLQConfig()` factory function

- [x] Task 3: Create DLQEvent struct (AC: #2)
  - [x] Define `DLQEvent` with all required fields
  - [x] Add `OriginalEvent Event` (the failed event)
  - [x] Add `ErrorMessage string` (formatted error)
  - [x] Add `RetryCount int` (attempts made)
  - [x] Add `FailedAt time.Time` (timestamp of final failure)
  - [x] Add `SourceTopic string` (original topic/queue)
  - [x] Add `StackTrace string` (optional, if configured)
  - [x] Add `Headers map[string]string` (for custom metadata)

- [x] Task 4: Create NopDeadLetterQueue (AC: #6)
  - [x] Implement `NopDeadLetterQueue` that does nothing
  - [x] Add `NewNopDeadLetterQueue()` factory function
  - [x] Send returns nil immediately
  - [x] Close returns nil immediately

- [x] Task 5: Create MockDeadLetterQueue for testing (AC: #7)
  - [x] Create `MockDeadLetterQueue` with event capture
  - [x] Add `Events() []DLQEvent` for inspection
  - [x] Add `EventCount() int` for quick checks
  - [x] Add `LastEvent() DLQEvent` for verification
  - [x] Add `Clear()` to reset captured events

- [x] Task 6: Create DLQHandler wrapper (AC: #5)
  - [x] Create `DLQHandler` struct that wraps `EventHandler`
  - [x] Add `NewDLQHandler(handler, dlq, config) EventHandler`
  - [x] Implement retry logic with configurable attempts
  - [x] Forward to DLQ after max retries exceeded
  - [x] Preserve original error context in DLQEvent

- [x] Task 7: Define DLQ metrics (AC: #4)
  - [x] Define metric names in interface documentation
  - [x] `event_dlq_total{topic, error_type}` - Counter
  - [x] `event_dlq_errors_total{topic}` - Counter (DLQ write failures)
  - [x] `event_dlq_queue_size{topic}` - Gauge (current queue depth)
  - [x] Document alerting patterns for DLQ thresholds

- [x] Task 8: Define DLQ sentinel errors (AC: #1)
  - [x] Create `ErrDLQClosed` sentinel error
  - [x] Create `ErrDLQFull` sentinel error (if bounded queue)
  - [x] Document error handling patterns in godoc

- [x] Task 9: Write unit tests for DLQ components
  - [x] Extend `internal/runtimeutil/events_test.go`
  - [x] Test `NopDeadLetterQueue` implements interface
  - [x] Test `MockDeadLetterQueue` event capture
  - [x] Test `DLQHandler` retry logic
  - [x] Test `DLQHandler` forwards after max retries
  - [x] Test `DLQEvent` struct serialization
  - [x] Test `DLQConfig` defaults and validation

- [x] Task 10: Update documentation (AC: all)
  - [x] Update `AGENTS.md` with DLQ patterns
  - [x] Update `README.md` with DLQ usage examples
  - [x] Update `docs/architecture.md` with DLQ section
  - [x] Add godoc examples for interface usage

## Dev Notes

### Architecture Compliance

**Location:** `internal/runtimeutil/events.go` - This is the PORT definition for DLQ operations.

**Layer Boundaries:**
- `DeadLetterQueue` interface (port) is in `runtimeutil` package
- Future implementations (KafkaDLQ, RabbitMQDLQ) go in `infra/kafka/dlq.go`, `infra/rabbitmq/dlq.go`
- DLQHandler wrapper lives in `runtimeutil` as it's broker-agnostic

### Existing Patterns to Follow (Story 13.1, 13.2, 13.3)

From `internal/runtimeutil/events.go`:

```go
// EventConsumer interface pattern
type EventConsumer interface {
    Subscribe(ctx context.Context, topic string, handler EventHandler) error
    Close() error
}

// Sentinel errors pattern
var (
    ErrConsumerClosed = errors.New("consumer closed")
    ErrProcessingTimeout = errors.New("processing timeout exceeded")
    ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)

// NopImplementation pattern
type NopEventConsumer struct {
    mu     sync.Mutex
    closed bool
}
```

**Apply same patterns for DLQ:**
- Interface with clear contract
- NopDLQ for testing
- MockDLQ for behavior verification  
- Sentinel errors for common failure modes

### Proposed DeadLetterQueue Interface

```go
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
//  // Configure DLQ
//  cfg := runtimeutil.DefaultDLQConfig()
//  cfg.TopicName = "orders.dlq"
//  
//  // Wrap handler with DLQ support
//  handler := runtimeutil.NewDLQHandler(myHandler, dlq, cfg)
//  consumer.Subscribe(ctx, "orders", handler)
//
// Metrics (implemented by concrete DLQ implementations):
//
//  event_dlq_total{topic, error_type}    - Counter: Events moved to DLQ
//  event_dlq_errors_total{topic}         - Counter: DLQ write failures
//  event_dlq_queue_size{topic}           - Gauge: Current DLQ depth
type DeadLetterQueue interface {
    // Send moves a failed event to the dead letter queue.
    // Returns error if the DLQ write fails.
    Send(ctx context.Context, event DLQEvent) error
    
    // Close gracefully shuts down the DLQ connection.
    Close() error
}
```

### DLQConfig Structure

```go
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
```

### DLQHandler Wrapper Pattern

```go
// DLQHandler wraps an EventHandler with retry and DLQ logic.
type DLQHandler struct {
    handler    EventHandler
    dlq        DeadLetterQueue
    maxRetries int
    retryDelay time.Duration
    dlqTopic   string
    includeStack bool
}

// NewDLQHandler creates a handler that retries failures and forwards to DLQ.
func NewDLQHandler(handler EventHandler, dlq DeadLetterQueue, config DLQConfig, consumerConfig ConsumerConfig) EventHandler {
    h := &DLQHandler{
        handler:      handler,
        dlq:          dlq,
        maxRetries:   consumerConfig.MaxRetries,
        retryDelay:   config.RetryDelay,
        dlqTopic:     config.TopicName,
        includeStack: config.IncludeStackTrace,
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
        
        if attempt < h.maxRetries {
            time.Sleep(h.retryDelay)
        }
    }
    
    // Max retries exceeded, move to DLQ
    dlqEvent := DLQEvent{
        OriginalEvent: event,
        ErrorMessage:  lastErr.Error(),
        RetryCount:    h.maxRetries + 1,
        FailedAt:      time.Now().UTC(),
        SourceTopic:   h.dlqTopic, // Set by consumer
    }
    
    if h.includeStack {
        dlqEvent.StackTrace = string(debug.Stack())
    }
    
    return h.dlq.Send(ctx, dlqEvent)
}
```

### Sentinel Errors

```go
var (
    // ErrDLQClosed is returned when operations are attempted on a closed DLQ.
    ErrDLQClosed = errors.New("dlq closed")
    
    // ErrDLQFull is returned when the DLQ has reached capacity.
    ErrDLQFull = errors.New("dlq full")
)
```

### Metrics to Define

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `event_dlq_total` | Counter | `source_topic`, `error_type` | Events moved to DLQ |
| `event_dlq_errors_total` | Counter | `source_topic` | Failed DLQ writes |
| `event_dlq_queue_size` | Gauge | `dlq_topic` | Current DLQ depth |
| `event_dlq_processing_attempts` | Histogram | `source_topic` | Retry attempts before DLQ |

### Alerting Pattern

```yaml
# deploy/prometheus/alerts.yaml addition
- alert: DLQThresholdExceeded
  expr: |
    sum(increase(event_dlq_total[1h])) > 100
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High DLQ rate detected"
    description: "More than 100 events moved to DLQ in the last hour"
    runbook_url: "docs/runbook/dlq-threshold.md"
```

### Previous Story Learnings (Stories 13.1, 13.2, 13.3)

- ✅ Used `Close()` for graceful shutdown with `defer` in main.go
- ✅ Added `mu sync.Mutex` for thread-safe state tracking
- ✅ Created comprehensive Prometheus metrics with labels
- ✅ Created `NopEventPublisher`, `NopEventConsumer` for testing
- ✅ Used table-driven tests with AAA pattern (Arrange-Act-Assert)
- ✅ Added `Validate()` methods for config structs
- ✅ Used `errors.New()` for sentinel errors, `errors.Is()` for checking

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| DLQ interface separate from Consumer | Allows different DLQ backends than consumer |
| DLQHandler wraps EventHandler | Transparent retry/DLQ for any handler |
| DLQEvent includes SourceTopic | Enables replay from DLQ to original topic |
| Configurable stack trace | Privacy by default, debug when needed |
| Alert threshold in config | Customizable per-topic alerting |
| RetryDelay configurable | Allows backoff strategies |

### Future Implementations (Not in this story)

| Component | Implementation Location |
|-----------|------------------------|
| KafkaDLQ | `internal/infra/kafka/dlq.go` |
| RabbitMQDLQ | `internal/infra/rabbitmq/dlq.go` |
| DLQ Consumer | Separate service for processing DLQ |
| DLQ Dashboard | Grafana panel for DLQ monitoring |

### Project Structure Notes

Files to create/modify:
```
internal/
├── runtimeutil/
│   ├── events.go              # MODIFY: Add DLQ interface, DLQEvent, DLQHandler
│   └── events_test.go         # MODIFY: Add DLQ tests
docs/
├── architecture.md            # MODIFY: Add DLQ section
AGENTS.md                      # MODIFY: Add DLQ patterns
README.md                      # MODIFY: Add DLQ examples
```

### Testing Strategy

1. **Unit Tests:**
   - `TestNopDeadLetterQueue_Send` - Returns nil
   - `TestNopDeadLetterQueue_Close` - Returns nil
   - `TestMockDeadLetterQueue_Capture` - Events captured
   - `TestDLQHandler_Success` - No DLQ on success
   - `TestDLQHandler_RetryAndFail` - DLQ after max retries
   - `TestDLQEvent_Serialization` - JSON round-trip
   - `TestDLQConfig_Defaults` - Sensible defaults
   - `TestDLQConfig_Validate` - Validation errors

2. **Table-Driven Test Pattern:**
```go
func TestDLQHandler(t *testing.T) {
    tests := []struct {
        name           string
        handlerErr     error
        maxRetries     int
        wantDLQCalled  bool
        wantReturnErr  error
    }{
        {"success", nil, 3, false, nil},
        {"fail_all_retries", errors.New("fail"), 3, true, nil},
        {"fail_some_retries", errors.New("fail"), 0, true, nil},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange -> Act -> Assert
        })
    }
}
```

### References

- [Source: docs/epics.md#Story-13.4]
- [Source: docs/architecture.md#Extension-Interfaces]
- [Source: internal/runtimeutil/events.go] - EventConsumer/EventPublisher pattern
- [Source: docs/sprint-artifacts/13-3-create-event-consumer-interface.md] - Previous story patterns

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List

## Senior Developer Review (AI)

**Date:** 2025-12-14
**Reviewer:** Senior Dev Agent

### Findings & Fixes
- **CRITICAL**: Added  interface and implementation in  to satisfy AC #4.
- **MEDIUM**: Added documentation warning about double-retry risks in .
- **LOW**: Added  to verify context cancellation behavior.

**Outcome:** Approved with Fixes
