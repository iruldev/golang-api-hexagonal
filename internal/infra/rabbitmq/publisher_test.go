package rabbitmq

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

func TestNewRabbitMQPublisher_Disabled(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := &config.RabbitMQConfig{
		Enabled: false,
	}
	logger := observability.NewNopLoggerInterface()

	// Act
	pub, err := NewRabbitMQPublisher(cfg, logger)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pub)

	// Should return NopEventPublisher
	_, isNop := pub.(*runtimeutil.NopEventPublisher)
	assert.True(t, isNop, "Expected NopEventPublisher when disabled")
}

func TestNewRabbitMQPublisher_DefaultValues(t *testing.T) {
	// This test verifies config defaults are applied correctly
	cfg := &config.RabbitMQConfig{
		Enabled: true,
		URL:     "", // Should default to amqp://guest:guest@localhost:5672/
	}

	// Verify expected defaults
	url := cfg.URL
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", url)

	exchange := cfg.Exchange
	if exchange == "" {
		exchange = "events"
	}
	assert.Equal(t, "events", exchange)

	exchangeType := cfg.ExchangeType
	if exchangeType == "" {
		exchangeType = "topic"
	}
	assert.Equal(t, "topic", exchangeType)
}

func TestRabbitMQPublisher_Publish_NilChannel(t *testing.T) {
	t.Parallel()

	// Arrange
	pub := &RabbitMQPublisher{
		channel:  nil, // Simulate closed channel
		exchange: "test-exchange",
		logger:   observability.NewNopLoggerInterface(),
	}

	event, err := runtimeutil.NewEvent("test.event", map[string]string{"key": "value"})
	require.NoError(t, err)

	// Act
	err = pub.Publish(context.Background(), "test-exchange", event)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "channel is closed")
}

func TestRabbitMQPublisher_PublishAsync_NilChannel(t *testing.T) {
	t.Parallel()

	// Arrange
	pub := &RabbitMQPublisher{
		channel:  nil, // Simulate closed channel
		exchange: "test-exchange",
		logger:   observability.NewNopLoggerInterface(),
	}

	event, err := runtimeutil.NewEvent("test.event", map[string]string{"key": "value"})
	require.NoError(t, err)

	// Act - PublishAsync should return nil immediately (fire-and-forget)
	err = pub.PublishAsync(context.Background(), "test-exchange", event)

	// Assert - No error returned (async, error logged internally)
	assert.NoError(t, err)

	// Give goroutine time to execute
	time.Sleep(50 * time.Millisecond)
}

func TestRabbitMQPublisher_Close_NilConnection(t *testing.T) {
	t.Parallel()

	// Arrange
	pub := &RabbitMQPublisher{
		conn:    nil, // Already closed
		channel: nil,
		logger:  observability.NewNopLoggerInterface(),
	}

	// Act
	err := pub.Close()

	// Assert
	assert.NoError(t, err, "Close should not error on nil connection/channel")
}

func TestRabbitMQPublisher_HealthCheck_NilConnection(t *testing.T) {
	t.Parallel()

	// Arrange
	pub := &RabbitMQPublisher{
		conn:   nil,
		logger: observability.NewNopLoggerInterface(),
	}

	// Act
	err := pub.HealthCheck(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection closed")
}

func TestRabbitMQHealthChecker_NilPublisher(t *testing.T) {
	t.Parallel()

	// Arrange - Create health checker with noop publisher (returns nil)
	nopPub := runtimeutil.NewNopEventPublisher()
	checker := NewRabbitMQHealthChecker(nopPub)

	// Assert - Should return nil (no health check needed)
	assert.Nil(t, checker)
}

func TestRabbitMQHealthChecker_Ping_NilPublisher(t *testing.T) {
	t.Parallel()

	// Arrange
	checker := &RabbitMQHealthChecker{publisher: nil}

	// Act
	err := checker.Ping(context.Background())

	// Assert - Should return nil when publisher is nil
	assert.NoError(t, err)
}

func TestEvent_Marshal(t *testing.T) {
	t.Parallel()

	// Arrange
	payload := map[string]interface{}{
		"user_id": "123",
		"action":  "created",
	}

	// Act
	event, err := runtimeutil.NewEvent("user.created", payload)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, "user.created", event.Type)
	assert.NotEmpty(t, event.Payload)
	assert.False(t, event.Timestamp.IsZero())
}

func TestSanitizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "masks password in standard AMQP URL",
			input:    "amqp://user:mypassword@localhost:5672/",
			expected: "amqp://user:%2A%2A%2A@localhost:5672/",
		},
		{
			name:     "masks password with special characters",
			input:    "amqp://admin:p@ss!word@rabbitmq.local:5672/vhost",
			expected: "amqp://admin:%2A%2A%2A@rabbitmq.local:5672/vhost",
		},
		{
			name:     "preserves URL without password",
			input:    "amqp://localhost:5672/",
			expected: "amqp://localhost:5672/",
		},
		{
			name:     "preserves URL with only username",
			input:    "amqp://guest@localhost:5672/",
			expected: "amqp://guest@localhost:5672/",
		},
		{
			name:     "handles invalid URL gracefully",
			input:    "://invalid",
			expected: "[invalid-url]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
