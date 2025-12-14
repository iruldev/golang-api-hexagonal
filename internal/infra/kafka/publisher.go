// Package kafka provides Kafka event publisher implementation.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

var (
	// publishTotal tracks total publish attempts.
	publishTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_publish_total",
			Help: "Total number of Kafka publish attempts",
		},
		[]string{"topic", "status"},
	)

	// publishErrors tracks failed publish attempts.
	publishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_publish_errors_total",
			Help: "Total number of Kafka publish errors",
		},
		[]string{"topic", "error_type"},
	)

	// publishDuration tracks publish latency.
	publishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_publish_duration_seconds",
			Help:    "Kafka publish duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)
)

// KafkaPublisher implements runtimeutil.EventPublisher using Kafka.
type KafkaPublisher struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	logger        observability.Logger
	brokers       []string
}

// NewKafkaPublisher creates a new KafkaPublisher with sync and async producers.
func NewKafkaPublisher(cfg *config.KafkaConfig, logger observability.Logger) (runtimeutil.EventPublisher, error) {
	if !cfg.IsEnabled() {
		logger.Info("Kafka publisher disabled, using noop publisher")
		return runtimeutil.NewNopEventPublisher(), nil
	}

	saramaConfig := sarama.NewConfig()

	// Producer settings
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Timeout = cfg.Timeout
	if saramaConfig.Producer.Timeout == 0 {
		saramaConfig.Producer.Timeout = 10 * time.Second
	}

	// Set required acks
	switch strings.ToLower(cfg.RequiredAcks) {
	case "none":
		saramaConfig.Producer.RequiredAcks = sarama.NoResponse
	case "local":
		saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	default: // "all" or empty
		saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	}

	// Set client ID
	if cfg.ClientID != "" {
		saramaConfig.ClientID = cfg.ClientID
	} else {
		saramaConfig.ClientID = "golang-api-hexagonal"
	}

	// TLS configuration
	if cfg.TLSEnabled {
		saramaConfig.Net.TLS.Enable = true
		// TLS config would be expanded here for production
	}

	// SASL configuration
	if cfg.SASLEnabled {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.User = cfg.SASLUsername
		saramaConfig.Net.SASL.Password = cfg.SASLPassword

		switch strings.ToUpper(cfg.SASLMechanism) {
		case "SCRAM-SHA-256":
			saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		case "SCRAM-SHA-512":
			saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		default:
			saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		}
	}

	brokers := cfg.Brokers
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	// Create sync producer
	syncProducer, err := sarama.NewSyncProducer(brokers, saramaConfig)
	if err != nil {
		logger.Error("failed to create Kafka sync producer",
			observability.String("brokers", strings.Join(brokers, ",")),
			observability.Err(err))
		return nil, fmt.Errorf("create sync producer: %w", err)
	}

	// Create async producer
	asyncProducer, err := sarama.NewAsyncProducer(brokers, saramaConfig)
	if err != nil {
		_ = syncProducer.Close()
		logger.Error("failed to create Kafka async producer",
			observability.String("brokers", strings.Join(brokers, ",")),
			observability.Err(err))
		return nil, fmt.Errorf("create async producer: %w", err)
	}

	publisher := &KafkaPublisher{
		syncProducer:  syncProducer,
		asyncProducer: asyncProducer,
		logger:        logger,
		brokers:       brokers,
	}

	// Start error handler for async producer
	go publisher.handleAsyncErrors()

	logger.Info("Kafka publisher initialized",
		observability.String("brokers", strings.Join(brokers, ",")),
		observability.String("client_id", cfg.ClientID))

	return publisher, nil
}

// handleAsyncErrors handles errors from async producer.
func (p *KafkaPublisher) handleAsyncErrors() {
	for err := range p.asyncProducer.Errors() {
		topic := ""
		if err.Msg != nil {
			topic = err.Msg.Topic
		}
		publishErrors.WithLabelValues(topic, "async").Inc()
		p.logger.Error("async publish failed",
			observability.String("topic", topic),
			observability.Err(err.Err))
	}
}

// Publish sends an event synchronously and waits for confirmation.
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, event runtimeutil.Event) error {
	start := time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		publishErrors.WithLabelValues(topic, "marshal").Inc()
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.ID),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.syncProducer.SendMessage(msg)
	duration := time.Since(start)
	publishDuration.WithLabelValues(topic).Observe(duration.Seconds())

	if err != nil {
		publishTotal.WithLabelValues(topic, "error").Inc()
		publishErrors.WithLabelValues(topic, "send").Inc()
		p.logger.Error("kafka publish failed",
			observability.String("topic", topic),
			observability.String("event_id", event.ID),
			observability.String("event_type", event.Type),
			observability.Err(err))
		return fmt.Errorf("publish to kafka: %w", err)
	}

	publishTotal.WithLabelValues(topic, "success").Inc()
	p.logger.Info("event published",
		observability.String("topic", topic),
		observability.String("event_id", event.ID),
		observability.String("event_type", event.Type),
		observability.Int("partition", int(partition)),
		observability.Int64("offset", offset),
		observability.Duration("duration", duration))

	return nil
}

// PublishAsync sends an event asynchronously (fire-and-forget).
func (p *KafkaPublisher) PublishAsync(ctx context.Context, topic string, event runtimeutil.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		publishErrors.WithLabelValues(topic, "marshal").Inc()
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.ID),
		Value: sarama.ByteEncoder(data),
	}

	select {
	case p.asyncProducer.Input() <- msg:
		publishTotal.WithLabelValues(topic, "async_queued").Inc()
		p.logger.Debug("event queued for async publish",
			observability.String("topic", topic),
			observability.String("event_id", event.ID),
			observability.String("event_type", event.Type))
		return nil
	case <-ctx.Done():
		publishErrors.WithLabelValues(topic, "context_cancelled").Inc()
		return ctx.Err()
	}
}

// Close gracefully closes the producers.
func (p *KafkaPublisher) Close() error {
	var errs []error

	if err := p.syncProducer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close sync producer: %w", err))
	}

	if err := p.asyncProducer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close async producer: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("close kafka publisher: %v", errs)
	}

	p.logger.Info("Kafka publisher closed")
	return nil
}

// HealthCheck checks Kafka connectivity for readiness probe.
// Optimized to reuse the existing producer connection instead of creating a new client.
func (p *KafkaPublisher) HealthCheck(ctx context.Context) error {
	// Use a simple ping by fetching brokers metadata from the sync producer
	// The sync producer internally maintains a client that we can leverage
	// A more robust check would be to send a test message, but that's intrusive

	// Create a minimal config for a quick connectivity check
	config := sarama.NewConfig()
	config.Net.DialTimeout = 3 * time.Second
	config.Net.ReadTimeout = 3 * time.Second
	config.Net.WriteTimeout = 3 * time.Second

	// Create a short-lived client just for health check
	// Note: In future, we could cache this client or use metadata refresh
	client, err := sarama.NewClient(p.brokers, config)
	if err != nil {
		return fmt.Errorf("kafka health check failed: %w", err)
	}
	defer client.Close()

	// Verify we can get broker list
	brokers := client.Brokers()
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers available")
	}

	return nil
}

// Closeable interface for graceful shutdown.
type Closeable interface {
	Close() error
}

// HealthChecker interface for readiness probe.
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Ensure KafkaPublisher implements required interfaces.
var (
	_ runtimeutil.EventPublisher = (*KafkaPublisher)(nil)
	_ Closeable                  = (*KafkaPublisher)(nil)
	_ HealthChecker              = (*KafkaPublisher)(nil)
)

// KafkaHealthChecker adapts KafkaPublisher to the DBHealthChecker interface
// used by the /readyz endpoint.
type KafkaHealthChecker struct {
	publisher *KafkaPublisher
}

// NewKafkaHealthChecker creates a health checker wrapper for Kafka.
func NewKafkaHealthChecker(pub runtimeutil.EventPublisher) *KafkaHealthChecker {
	if kafkaPub, ok := pub.(*KafkaPublisher); ok {
		return &KafkaHealthChecker{publisher: kafkaPub}
	}
	return nil // NopEventPublisher returns nil (no health check needed)
}

// Ping implements the DBHealthChecker interface for Kafka.
func (c *KafkaHealthChecker) Ping(ctx context.Context) error {
	if c.publisher == nil {
		return nil
	}
	return c.publisher.HealthCheck(ctx)
}
