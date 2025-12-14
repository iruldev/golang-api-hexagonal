// Package rabbitmq provides RabbitMQ event publisher implementation.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

var (
	// publishTotal tracks total publish attempts.
	publishTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_publish_total",
			Help: "Total number of RabbitMQ publish attempts",
		},
		[]string{"exchange", "routing_key", "status"},
	)

	// publishErrors tracks failed publish attempts.
	publishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_publish_errors_total",
			Help: "Total number of RabbitMQ publish errors",
		},
		[]string{"exchange", "error_type"},
	)

	// publishDuration tracks publish latency.
	publishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rabbitmq_publish_duration_seconds",
			Help:    "RabbitMQ publish duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"exchange"},
	)
)

// RabbitMQPublisher implements runtimeutil.EventPublisher using RabbitMQ.
type RabbitMQPublisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	logger       observability.Logger
	exchange     string
	exchangeType string
	durable      bool
	mu           sync.RWMutex
}

// NewRabbitMQPublisher creates a new RabbitMQPublisher with publisher confirms.
func NewRabbitMQPublisher(cfg *config.RabbitMQConfig, logger observability.Logger) (runtimeutil.EventPublisher, error) {
	if !cfg.IsEnabled() {
		logger.Info("RabbitMQ publisher disabled, using noop publisher")
		return runtimeutil.NewNopEventPublisher(), nil
	}

	url := cfg.URL
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	exchange := cfg.Exchange
	if exchange == "" {
		exchange = "events"
	}

	exchangeType := cfg.ExchangeType
	if exchangeType == "" {
		exchangeType = "topic"
	}

	durable := cfg.Durable
	// Default to true if not explicitly set
	if !cfg.Durable && cfg.URL == "" {
		durable = true
	}

	// Create connection
	conn, err := amqp.Dial(url)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ",
			observability.String("url", sanitizeURL(url)),
			observability.Err(err))
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		logger.Error("failed to open RabbitMQ channel",
			observability.Err(err))
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Enable publisher confirms
	if err := channel.Confirm(false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		logger.Error("failed to enable publisher confirms",
			observability.Err(err))
		return nil, fmt.Errorf("enable confirms: %w", err)
	}

	// Declare exchange
	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		durable,      // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		logger.Error("failed to declare exchange",
			observability.String("exchange", exchange),
			observability.String("type", exchangeType),
			observability.Err(err))
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	publisher := &RabbitMQPublisher{
		conn:         conn,
		channel:      channel,
		logger:       logger,
		exchange:     exchange,
		exchangeType: exchangeType,
		durable:      durable,
	}

	logger.Info("RabbitMQ publisher initialized",
		observability.String("url", sanitizeURL(url)),
		observability.String("exchange", exchange),
		observability.String("exchange_type", exchangeType),
		observability.Bool("durable", durable))

	return publisher, nil
}

// sanitizeURL removes password from URL for logging.
func sanitizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "[invalid-url]"
	}
	if parsed.User != nil {
		if _, hasPass := parsed.User.Password(); hasPass {
			parsed.User = url.UserPassword(parsed.User.Username(), "***")
		}
	}
	return parsed.String()
}

// Publish sends an event synchronously and waits for confirmation.
func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, event runtimeutil.Event) error {
	start := time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		publishErrors.WithLabelValues(topic, "marshal").Inc()
		return fmt.Errorf("marshal event: %w", err)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.channel == nil {
		publishErrors.WithLabelValues(topic, "channel_closed").Inc()
		return fmt.Errorf("channel is closed")
	}

	// Use topic parameter as exchange name (following Kafka pattern)
	// Use event.Type as routing key for topic exchanges
	exchange := topic
	routingKey := event.Type

	// Publish with deferred confirmation
	confirmation, err := p.channel.PublishWithDeferredConfirmWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		true,       // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.ID,
			Timestamp:    event.Timestamp,
			Body:         data,
		},
	)
	if err != nil {
		duration := time.Since(start)
		publishDuration.WithLabelValues(exchange).Observe(duration.Seconds())
		publishTotal.WithLabelValues(exchange, routingKey, "error").Inc()
		publishErrors.WithLabelValues(exchange, "publish").Inc()
		p.logger.Error("rabbitmq publish failed",
			observability.String("exchange", exchange),
			observability.String("routing_key", routingKey),
			observability.String("event_id", event.ID),
			observability.String("event_type", event.Type),
			observability.Err(err))
		return fmt.Errorf("publish to rabbitmq: %w", err)
	}

	// Wait for confirmation
	confirmed := confirmation.Wait()
	duration := time.Since(start)
	publishDuration.WithLabelValues(exchange).Observe(duration.Seconds())

	if !confirmed {
		publishTotal.WithLabelValues(exchange, routingKey, "nack").Inc()
		publishErrors.WithLabelValues(exchange, "nack").Inc()
		p.logger.Error("message not confirmed by broker",
			observability.String("exchange", exchange),
			observability.String("routing_key", routingKey),
			observability.String("event_id", event.ID))
		return fmt.Errorf("message not confirmed by broker")
	}

	publishTotal.WithLabelValues(exchange, routingKey, "success").Inc()
	p.logger.Info("event published",
		observability.String("exchange", exchange),
		observability.String("routing_key", routingKey),
		observability.String("event_id", event.ID),
		observability.String("event_type", event.Type),
		observability.Duration("duration", duration))

	return nil
}

// PublishAsync sends an event asynchronously (fire-and-forget).
func (p *RabbitMQPublisher) PublishAsync(ctx context.Context, topic string, event runtimeutil.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		publishErrors.WithLabelValues(topic, "marshal").Inc()
		return fmt.Errorf("marshal event: %w", err)
	}

	// Fire-and-forget in goroutine
	go func() {
		p.mu.RLock()
		defer p.mu.RUnlock()

		if p.channel == nil {
			publishErrors.WithLabelValues(topic, "channel_closed").Inc()
			p.logger.Error("async publish failed: channel closed",
				observability.String("exchange", topic),
				observability.String("event_id", event.ID))
			return
		}

		exchange := topic
		routingKey := event.Type

		err := p.channel.PublishWithContext(
			context.Background(), // Use detached context for fire-and-forget
			exchange,             // exchange
			routingKey,           // routing key
			false,                // mandatory (false for async - don't wait for routing confirmation)
			false,                // immediate
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				MessageId:    event.ID,
				Timestamp:    event.Timestamp,
				Body:         data,
			},
		)
		if err != nil {
			publishTotal.WithLabelValues(exchange, routingKey, "async_error").Inc()
			publishErrors.WithLabelValues(exchange, "async").Inc()
			p.logger.Error("async publish failed",
				observability.String("exchange", exchange),
				observability.String("routing_key", routingKey),
				observability.String("event_id", event.ID),
				observability.Err(err))
			return
		}

		publishTotal.WithLabelValues(exchange, routingKey, "async_sent").Inc()
		p.logger.Debug("event queued for async publish",
			observability.String("exchange", exchange),
			observability.String("routing_key", routingKey),
			observability.String("event_id", event.ID),
			observability.String("event_type", event.Type))
	}()

	return nil
}

// Close gracefully closes the RabbitMQ connection and channel.
func (p *RabbitMQPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close channel: %w", err))
		}
		p.channel = nil
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close connection: %w", err))
		}
		p.conn = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("close rabbitmq publisher: %v", errs)
	}

	p.logger.Info("RabbitMQ publisher closed")
	return nil
}

// HealthCheck checks RabbitMQ connectivity for readiness probe.
func (p *RabbitMQPublisher) HealthCheck(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection closed")
	}

	if p.channel == nil {
		return fmt.Errorf("rabbitmq channel closed")
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

// Ensure RabbitMQPublisher implements required interfaces.
var (
	_ runtimeutil.EventPublisher = (*RabbitMQPublisher)(nil)
	_ Closeable                  = (*RabbitMQPublisher)(nil)
	_ HealthChecker              = (*RabbitMQPublisher)(nil)
)

// RabbitMQHealthChecker adapts RabbitMQPublisher to the DBHealthChecker interface
// used by the /readyz endpoint.
type RabbitMQHealthChecker struct {
	publisher *RabbitMQPublisher
}

// NewRabbitMQHealthChecker creates a health checker wrapper for RabbitMQ.
func NewRabbitMQHealthChecker(pub runtimeutil.EventPublisher) *RabbitMQHealthChecker {
	if rmqPub, ok := pub.(*RabbitMQPublisher); ok {
		return &RabbitMQHealthChecker{publisher: rmqPub}
	}
	return nil // NopEventPublisher returns nil (no health check needed)
}

// Ping implements the DBHealthChecker interface for RabbitMQ.
func (c *RabbitMQHealthChecker) Ping(ctx context.Context) error {
	if c.publisher == nil {
		return nil
	}
	return c.publisher.HealthCheck(ctx)
}
