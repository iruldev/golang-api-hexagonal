//go:build integration

package rabbitmq

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
	testhelper "github.com/iruldev/golang-api-hexagonal/internal/testing"
)

func TestRabbitMQPublisher_Integration_Publish(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start RabbitMQ container
	rmqContainer, err := testhelper.NewRabbitMQContainer(ctx)
	require.NoError(t, err, "Failed to start RabbitMQ container")
	defer rmqContainer.Terminate(ctx)

	// Create publisher
	cfg := &config.RabbitMQConfig{
		Enabled:      true,
		URL:          rmqContainer.URL,
		Exchange:     "test-events",
		ExchangeType: "topic",
		Durable:      true,
	}
	logger := observability.NewNopLoggerInterface()

	pub, err := NewRabbitMQPublisher(cfg, logger)
	require.NoError(t, err, "Failed to create publisher")
	defer func() {
		if rmqPub, ok := pub.(*RabbitMQPublisher); ok {
			rmqPub.Close()
		}
	}()

	// Verify it's not a noop publisher
	_, isNop := pub.(*runtimeutil.NopEventPublisher)
	require.False(t, isNop, "Expected real publisher, not NopEventPublisher")

	// Create test event
	event, err := runtimeutil.NewEvent("test.created", map[string]string{
		"id":     "123",
		"action": "test",
	})
	require.NoError(t, err)

	// Act: Publish event
	err = pub.Publish(ctx, "test-events", event)

	// Assert: No error
	assert.NoError(t, err, "Publish should succeed")
}

func TestRabbitMQPublisher_Integration_PublishAsync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start RabbitMQ container
	rmqContainer, err := testhelper.NewRabbitMQContainer(ctx)
	require.NoError(t, err, "Failed to start RabbitMQ container")
	defer rmqContainer.Terminate(ctx)

	// Create publisher
	cfg := &config.RabbitMQConfig{
		Enabled:      true,
		URL:          rmqContainer.URL,
		Exchange:     "test-events",
		ExchangeType: "topic",
		Durable:      true,
	}
	logger := observability.NewNopLoggerInterface()

	pub, err := NewRabbitMQPublisher(cfg, logger)
	require.NoError(t, err, "Failed to create publisher")
	defer func() {
		if rmqPub, ok := pub.(*RabbitMQPublisher); ok {
			rmqPub.Close()
		}
	}()

	// Create test event
	event, err := runtimeutil.NewEvent("test.async", map[string]string{
		"id": "456",
	})
	require.NoError(t, err)

	// Act: Publish event asynchronously
	err = pub.PublishAsync(ctx, "test-events", event)

	// Assert: No error (fire-and-forget)
	assert.NoError(t, err)

	// Give goroutine time to complete
	time.Sleep(100 * time.Millisecond)
}

func TestRabbitMQPublisher_Integration_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start RabbitMQ container
	rmqContainer, err := testhelper.NewRabbitMQContainer(ctx)
	require.NoError(t, err, "Failed to start RabbitMQ container")
	defer rmqContainer.Terminate(ctx)

	// Create publisher
	cfg := &config.RabbitMQConfig{
		Enabled:      true,
		URL:          rmqContainer.URL,
		Exchange:     "health-check-test",
		ExchangeType: "topic",
		Durable:      true,
	}
	logger := observability.NewNopLoggerInterface()

	pub, err := NewRabbitMQPublisher(cfg, logger)
	require.NoError(t, err, "Failed to create publisher")

	rmqPub, ok := pub.(*RabbitMQPublisher)
	require.True(t, ok, "Expected RabbitMQPublisher type")
	defer rmqPub.Close()

	// Act: Health check
	err = rmqPub.HealthCheck(ctx)

	// Assert: Should pass
	assert.NoError(t, err, "Health check should pass on healthy connection")
}

func TestRabbitMQPublisher_Integration_PublishAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start RabbitMQ container
	rmqContainer, err := testhelper.NewRabbitMQContainer(ctx)
	require.NoError(t, err, "Failed to start RabbitMQ container")
	defer rmqContainer.Terminate(ctx)

	// Create publisher
	cfg := &config.RabbitMQConfig{
		Enabled:      true,
		URL:          rmqContainer.URL,
		Exchange:     "integration-test",
		ExchangeType: "topic",
		Durable:      true,
	}
	logger := observability.NewNopLoggerInterface()

	pub, err := NewRabbitMQPublisher(cfg, logger)
	require.NoError(t, err, "Failed to create publisher")

	rmqPub, ok := pub.(*RabbitMQPublisher)
	require.True(t, ok, "Expected RabbitMQPublisher type")
	defer rmqPub.Close()

	// Create consumer connection for verification
	conn, err := amqp.Dial(rmqContainer.URL)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// Declare queue and bind to exchange
	queue, err := ch.QueueDeclare(
		"test-queue", // name
		false,        // durable
		true,         // auto-delete
		false,        // exclusive
		false,        // no-wait
		nil,          // args
	)
	require.NoError(t, err)

	err = ch.QueueBind(
		queue.Name,         // queue
		"#",                // routing key (match all for topic)
		"integration-test", // exchange
		false,              // no-wait
		nil,                // args
	)
	require.NoError(t, err)

	// Start consuming
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer tag
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	require.NoError(t, err)

	// Create and publish test event
	testPayload := map[string]string{
		"id":      "integration-test-1",
		"message": "Hello RabbitMQ!",
	}
	event, err := runtimeutil.NewEvent("test.created", testPayload)
	require.NoError(t, err)

	err = pub.Publish(ctx, "integration-test", event)
	require.NoError(t, err)

	// Wait for message
	select {
	case msg := <-msgs:
		assert.NotEmpty(t, msg.Body)
		assert.Equal(t, "application/json", msg.ContentType)
		assert.Equal(t, event.ID, msg.MessageId)
		t.Logf("Received message: %s", string(msg.Body))
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestRabbitMQPublisher_Integration_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start RabbitMQ container
	rmqContainer, err := testhelper.NewRabbitMQContainer(ctx)
	require.NoError(t, err, "Failed to start RabbitMQ container")
	defer rmqContainer.Terminate(ctx)

	// Create publisher
	cfg := &config.RabbitMQConfig{
		Enabled:      true,
		URL:          rmqContainer.URL,
		Exchange:     "close-test",
		ExchangeType: "topic",
		Durable:      true,
	}
	logger := observability.NewNoopLogger()

	pub, err := NewRabbitMQPublisher(cfg, logger)
	require.NoError(t, err)

	rmqPub, ok := pub.(*RabbitMQPublisher)
	require.True(t, ok)

	// Act: Close
	err = rmqPub.Close()
	assert.NoError(t, err)

	// Assert: Health check should fail after close
	err = rmqPub.HealthCheck(ctx)
	assert.Error(t, err)
}
