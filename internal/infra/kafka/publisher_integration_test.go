//go:build integration

package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
	inttesting "github.com/iruldev/golang-api-hexagonal/internal/testing"
)

// =============================================================================
// Integration Tests (require Docker)
// =============================================================================
// Run with: go test -v -tags=integration ./internal/infra/kafka/...
// =============================================================================

// testIntegrationLogger returns a logger for integration tests.
func testIntegrationLogger() observability.Logger {
	return observability.NewZapLogger(observability.NewNopLogger())
}

// =============================================================================
// Kafka Publisher Integration Tests
// =============================================================================

func TestKafkaPublisher_Integration_PublishAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Start Kafka container
	kafkaContainer, err := inttesting.NewKafkaContainer(ctx)
	require.NoError(t, err, "Failed to start Kafka container")
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Kafka container: %v", err)
		}
	}()

	// Wait for Kafka to be ready
	time.Sleep(5 * time.Second)

	// Configure publisher
	cfg := &config.KafkaConfig{
		Enabled:      true,
		Brokers:      kafkaContainer.Brokers,
		ClientID:     "integration-test-client",
		Timeout:      10 * time.Second,
		RequiredAcks: "all",
	}
	logger := testIntegrationLogger()

	// Create publisher
	publisher, err := NewKafkaPublisher(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, publisher)

	kafkaPub, ok := publisher.(*KafkaPublisher)
	require.True(t, ok, "Expected *KafkaPublisher")
	defer kafkaPub.Close()

	// Create test event
	topic := "integration-test-topic"
	event, err := runtimeutil.NewEvent("test.event.created", map[string]string{
		"message":   "Hello from integration test",
		"timestamp": time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	// Act: Publish event synchronously
	err = publisher.Publish(ctx, topic, event)
	require.NoError(t, err)

	// Verify: Consume the published message
	consumerConfig := sarama.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(cfg.Brokers, consumerConfig)
	require.NoError(t, err)
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	require.NoError(t, err)
	defer partitionConsumer.Close()

	// Wait for message with timeout
	select {
	case msg := <-partitionConsumer.Messages():
		// Verify message content
		var receivedEvent runtimeutil.Event
		err := json.Unmarshal(msg.Value, &receivedEvent)
		require.NoError(t, err)

		assert.Equal(t, event.ID, receivedEvent.ID)
		assert.Equal(t, event.Type, receivedEvent.Type)
		t.Logf("Successfully consumed message: ID=%s, Type=%s", receivedEvent.ID, receivedEvent.Type)

	case err := <-partitionConsumer.Errors():
		t.Fatalf("Consumer error: %v", err)

	case <-time.After(30 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestKafkaPublisher_Integration_PublishAsync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Start Kafka container
	kafkaContainer, err := inttesting.NewKafkaContainer(ctx)
	require.NoError(t, err, "Failed to start Kafka container")
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Kafka container: %v", err)
		}
	}()

	// Wait for Kafka to be ready
	time.Sleep(5 * time.Second)

	// Configure publisher
	cfg := &config.KafkaConfig{
		Enabled:      true,
		Brokers:      kafkaContainer.Brokers,
		ClientID:     "async-test-client",
		Timeout:      10 * time.Second,
		RequiredAcks: "all",
	}
	logger := testIntegrationLogger()

	// Create publisher
	publisher, err := NewKafkaPublisher(cfg, logger)
	require.NoError(t, err)

	kafkaPub, ok := publisher.(*KafkaPublisher)
	require.True(t, ok)
	defer kafkaPub.Close()

	// Create test event
	topic := "async-test-topic"
	event, err := runtimeutil.NewEvent("async.event", map[string]string{
		"async": "true",
	})
	require.NoError(t, err)

	// Act: Publish event asynchronously
	err = publisher.PublishAsync(ctx, topic, event)
	require.NoError(t, err)

	// Wait a bit for async delivery
	time.Sleep(2 * time.Second)

	// Verify: Check message was delivered
	consumerConfig := sarama.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(cfg.Brokers, consumerConfig)
	require.NoError(t, err)
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	require.NoError(t, err)
	defer partitionConsumer.Close()

	// Wait for message
	select {
	case msg := <-partitionConsumer.Messages():
		var receivedEvent runtimeutil.Event
		err := json.Unmarshal(msg.Value, &receivedEvent)
		require.NoError(t, err)
		assert.Equal(t, "async.event", receivedEvent.Type)
		t.Logf("Successfully received async message: ID=%s", receivedEvent.ID)

	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for async message")
	}
}

func TestKafkaPublisher_Integration_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Start Kafka container
	kafkaContainer, err := inttesting.NewKafkaContainer(ctx)
	require.NoError(t, err, "Failed to start Kafka container")
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Kafka container: %v", err)
		}
	}()

	// Wait for Kafka to be ready
	time.Sleep(5 * time.Second)

	// Configure publisher
	cfg := &config.KafkaConfig{
		Enabled:      true,
		Brokers:      kafkaContainer.Brokers,
		ClientID:     "health-test-client",
		Timeout:      10 * time.Second,
		RequiredAcks: "all",
	}
	logger := testIntegrationLogger()

	// Create publisher
	publisher, err := NewKafkaPublisher(cfg, logger)
	require.NoError(t, err)

	kafkaPub, ok := publisher.(*KafkaPublisher)
	require.True(t, ok)
	defer kafkaPub.Close()

	// Act: Perform health check
	err = kafkaPub.HealthCheck(ctx)

	// Assert
	assert.NoError(t, err, "Health check should succeed when Kafka is available")
}

func TestKafkaPublisher_Integration_DisabledReturnsNoop(t *testing.T) {
	// Arrange
	cfg := &config.KafkaConfig{
		Enabled: false,
	}
	logger := testIntegrationLogger()

	// Act
	publisher, err := NewKafkaPublisher(cfg, logger)

	// Assert
	require.NoError(t, err)
	_, isNop := publisher.(*runtimeutil.NopEventPublisher)
	assert.True(t, isNop, "Expected NopEventPublisher when Kafka is disabled")
}
