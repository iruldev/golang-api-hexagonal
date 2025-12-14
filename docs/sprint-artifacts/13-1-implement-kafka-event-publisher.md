# Story 13.1: Implement Kafka Event Publisher

Status: done

## Story

As a developer,
I want to publish events to Kafka,
So that other services can react asynchronously with high throughput.

## Acceptance Criteria

1. **Given** `KafkaPublisher` implements `EventPublisher` interface
   **When** `Publish(ctx, "topic", event)` is called
   **Then** message is sent to Kafka broker synchronously
   **And** returns error if publish fails

2. **Given** `KafkaPublisher` implements `EventPublisher` interface
   **When** `PublishAsync(ctx, "topic", event)` is called
   **Then** message is sent asynchronously (fire-and-forget)
   **And** returns immediately without waiting

3. **Given** a valid Kafka producer configuration
   **When** the application starts with `KAFKA_ENABLED=true`
   **Then** Kafka producer is initialized with connection pooling
   **And** structured logging captures initialization status

4. **Given** any event is published
   **When** publish operation completes or fails
   **Then** structured logs include: topic, event_id, event_type, success/error status

5. **Given** Kafka is configured as event publisher
   **When** `/readyz` is requested
   **Then** Kafka connectivity is checked as part of readiness health check
   **And** 503 is returned if Kafka is unavailable

## Tasks / Subtasks

- [x] Task 1: Add Kafka dependency (AC: #3)
  - [x] Add `github.com/IBM/sarama` to go.mod (maintained fork of Shopify/sarama)
  - [x] Run `go mod tidy`

- [x] Task 2: Create Kafka configuration (AC: #3)
  - [x] Add Kafka config struct to `internal/config/config.go`
  - [x] Add environment variables: `KAFKA_ENABLED`, `KAFKA_BROKERS`, `KAFKA_CLIENT_ID`
  - [x] Add optional TLS/SASL config placeholders for production

- [x] Task 3: Implement KafkaPublisher (AC: #1, #2, #4)
  - [x] Create `internal/infra/kafka/publisher.go`
  - [x] Implement `EventPublisher` interface from `internal/runtimeutil/events.go`
  - [x] Implement `Publish()` - synchronous with acknowledgement
  - [x] Implement `PublishAsync()` - fire-and-forget via goroutine
  - [x] Add structured logging using `observability/logger` pattern
  - [x] Add Prometheus metrics: `kafka_publish_total`, `kafka_publish_errors_total`, `kafka_publish_duration_seconds`

- [x] Task 4: Add factory and graceful shutdown (AC: #3)
  - [x] Create `NewKafkaPublisher(cfg *config.Kafka, logger observability.Logger) (EventPublisher, error)`
  - [x] Implement `Close()` method for graceful shutdown
  - [x] Add connection health check method for readiness probe

- [x] Task 4.5: Update Test Helpers
  - [x] Update `internal/testing/containers.go` to include `NewKafkaContainer`
  - [x] Ensure it uses the same standardization as Postgres/Redis helpers

- [x] Task 5: Integrate with main.go (AC: #3, #5)
  - [x] Add conditional initialization when `KAFKA_ENABLED=true`
  - [x] Register Kafka health check with `/readyz` endpoint
  - [x] Handle graceful shutdown in main.go

- [x] Task 6: Update docker-compose (AC: #3)
  - [x] Add Kafka + Zookeeper services to `docker-compose.yaml`
  - [x] Add environment variables to `.env.example`

- [x] Task 7: Write unit tests (AC: #1, #2)
  - [x] Create `internal/infra/kafka/publisher_test.go`
  - [x] Test `Publish()` with mock producer
  - [x] Test `PublishAsync()` behavior
  - [x] Test error handling and logging
  - [x] Use table-driven tests with AAA pattern

- [x] Task 8: Write integration tests (AC: #1, #2, #3)
  - [x] Create `internal/infra/kafka/publisher_integration_test.go`
  - [x] Use `testing.NewKafkaContainer` from helpers
  - [x] Test real message publish/consume flow
  - [x] Add `//go:build integration` tag

- [x] Task 9: Update documentation (AC: all)
  - [x] Update `AGENTS.md` with Kafka publisher patterns
  - [x] Update `README.md` with Kafka setup instructions
  - [x] Update `docs/architecture.md` with event-driven section


## Dev Notes

### Architecture Compliance

**Location:** `internal/infra/kafka/` - This is an infrastructure adapter implementing the port defined in `internal/runtimeutil/events.go`.

**Layer Boundaries:**
- `EventPublisher` interface (port) is in `runtimeutil` package
- `KafkaPublisher` (adapter) goes in `infra/kafka` package
- Usecase layer calls publisher via interface, NOT concrete implementation

### Existing Interface to Implement

```go
// From internal/runtimeutil/events.go
type EventPublisher interface {
    Publish(ctx context.Context, topic string, event Event) error
    PublishAsync(ctx context.Context, topic string, event Event) error
}

type Event struct {
    ID        string          `json:"id"`
    Type      string          `json:"type"`
    Payload   json.RawMessage `json:"payload"`
    Timestamp time.Time       `json:"timestamp"`
}
```

### Kafka Library Choice

**Use:** `github.com/IBM/sarama` (v1.43+)
- This is the maintained fork of `Shopify/sarama`
- Most popular Go Kafka client with robust producer/consumer support
- Supports sync and async producers

**Alternative considered:** `segmentio/kafka-go` - simpler API but less feature-rich

### Implementation Pattern (from architecture.md)

```go
type KafkaPublisher struct {
    producer sarama.SyncProducer
    asyncProd sarama.AsyncProducer
    logger   observability.Logger
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, event Event) error {
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }
    
    msg := &sarama.ProducerMessage{
        Topic: topic,
        Key:   sarama.StringEncoder(event.ID),
        Value: sarama.ByteEncoder(data),
    }
    
    _, _, err = p.producer.SendMessage(msg)
    if err != nil {
        p.logger.Error("kafka publish failed", 
            observability.String("topic", topic),
            observability.String("event_id", event.ID),
            observability.String("error", err.Error()))
        return fmt.Errorf("publish to kafka: %w", err)
    }
    
    p.logger.Info("event published",
        observability.String("topic", topic),
        observability.String("event_id", event.ID),
        observability.String("event_type", event.Type))
    return nil
}
```

### Configuration Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KAFKA_ENABLED` | `false` | Enable Kafka publisher |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated broker addresses |
| `KAFKA_CLIENT_ID` | `golang-api-hexagonal` | Client identifier |
| `KAFKA_PRODUCER_TIMEOUT` | `10s` | Producer timeout |
| `KAFKA_PRODUCER_REQUIRED_ACKS` | `all` | Ack level: `all`, `local`, `none` |

### Docker Compose Services

```yaml
# Add to docker-compose.yaml
zookeeper:
  image: confluentinc/cp-zookeeper:7.5.0
  environment:
    ZOOKEEPER_CLIENT_PORT: 2181
  ports:
    - "2181:2181"

kafka:
  image: confluentinc/cp-kafka:7.5.0
  depends_on:
    - zookeeper
  environment:
    KAFKA_BROKER_ID: 1
    KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
    KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
    KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
  ports:
    - "9092:9092"
```

### Metrics to Expose

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `kafka_publish_total` | Counter | `topic`, `status` | Total publish attempts |
| `kafka_publish_errors_total` | Counter | `topic`, `error_type` | Failed publishes |
| `kafka_publish_duration_seconds` | Histogram | `topic` | Publish latency |

### Testing with Testcontainers

Instead of duplicating container setup, extend `internal/testing/containers.go`:

```go
// Add to internal/testing/containers.go
type KafkaContainer struct {
    Container testcontainers.Container
    Brokers   []string
}

func NewKafkaContainer(ctx context.Context) (*KafkaContainer, error) {
    // Implementation using "confluentinc/cp-kafka:7.5.0"
    // Configure with kraft enabled or zookeeper as needed
    // Return container with broker addresses
}
```

### Previous Story Learnings (Story 12.4)

- Use environment constants from `internal/config/config.go` for consistency
- Add helper methods like `cfg.Kafka.IsEnabled()` for cleaner checks
- Include integration tests for enabled/disabled scenarios
- Update both `AGENTS.md` and `README.md` with new patterns

### Project Structure Notes

Files to create/modify:
```
internal/
├── config/config.go          # Add Kafka config struct
├── infra/
│   └── kafka/
│       ├── publisher.go      # NEW: KafkaPublisher implementation
│       └── publisher_test.go # NEW: Unit tests
cmd/server/main.go            # Wire Kafka publisher conditionally
docker-compose.yaml           # Add Kafka + Zookeeper
.env.example                  # Add Kafka env vars
```

### References

- [Source: docs/epics.md#Story-13.1]
- [Source: docs/architecture.md#EventPublisher-Interface]
- [Source: internal/runtimeutil/events.go] - Interface definition
- [IBM Sarama Documentation](https://github.com/IBM/sarama)
- [Confluent Kafka Docker](https://hub.docker.com/r/confluentinc/cp-kafka)

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

gemini-2.5-pro

### Debug Log References

### Completion Notes List

- Implemented KafkaPublisher using IBM/sarama library with sync and async producers
- Added Prometheus metrics (kafka_publish_total, kafka_publish_errors_total, kafka_publish_duration_seconds)
- Added KafkaHealthChecker adapter for /readyz endpoint
- Used Redpanda in testcontainers for faster integration test startup
- Added TLS/SASL config placeholders for production use

### File List

| File | Action |
|------|--------|
| `internal/infra/kafka/publisher.go` | NEW - KafkaPublisher implementation |
| `internal/infra/kafka/publisher_test.go` | NEW - Unit tests |
| `internal/infra/kafka/publisher_integration_test.go` | NEW - Integration tests |
| `internal/config/config.go` | MODIFIED - Added KafkaConfig struct |
| `internal/testing/containers.go` | MODIFIED - Added NewKafkaContainer |
| `internal/interface/http/handlers/health.go` | MODIFIED - Added WithKafka |
| `internal/interface/http/router.go` | MODIFIED - Added KafkaChecker |
| `cmd/server/main.go` | MODIFIED - Kafka initialization |
| `docker-compose.yaml` | MODIFIED - Added Zookeeper + Kafka |
| `.env.example` | MODIFIED - Added Kafka env vars |
| `AGENTS.md` | MODIFIED - Added Kafka section |
| `README.md` | MODIFIED - Added Kafka docs |

### Change Log

| Date | Change |
|------|--------|
| 2025-12-14 | Story created for Epic 13 - Event-Driven Implementations |
| 2025-12-14 | Implementation complete - All tasks done, tests passing |
| 2025-12-14 | Code Review: Fixed graceful shutdown in main.go, optimized health check in publisher.go |

### Senior Developer Review (AI)

**Reviewed:** 2025-12-14
**Issues Found:** 1 High, 1 Medium, 1 Low
**Status:** ✅ All HIGH and MEDIUM issues fixed

**Fixes Applied:**
1. **[HIGH] Graceful Shutdown** - Added `defer kafkaPub.Close()` in `main.go` to ensure buffered messages are flushed on shutdown.
2. **[MEDIUM] Health Check Optimization** - Added shorter timeouts (3s) for faster failure detection during readiness probes.
3. **[LOW] Test/Docker Mismatch** - Noted: Redpanda in tests vs Confluent Kafka in docker-compose is acceptable (Kafka-compatible).

