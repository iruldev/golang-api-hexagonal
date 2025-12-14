// Package testing provides test helpers for integration tests.
package testing

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a testcontainers PostgreSQL container.
type PostgresContainer struct {
	Container testcontainers.Container
	DSN       string
}

// NewPostgresContainer starts a PostgreSQL container for testing.
// Returns the container wrapper with DSN for connecting.
// Caller must call Terminate() when done.
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	return &PostgresContainer{
		Container: container,
		DSN:       dsn,
	}, nil
}

// Terminate stops and removes the PostgreSQL container.
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

// RedisContainer wraps a testcontainers Redis container.
type RedisContainer struct {
	Container testcontainers.Container
	Addr      string
}

// NewRedisContainer starts a Redis container for testing.
// Returns the container wrapper with address for connecting.
// Caller must call Terminate() when done.
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	container, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start redis container: %w", err)
	}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get endpoint: %w", err)
	}

	return &RedisContainer{
		Container: container,
		Addr:      endpoint,
	}, nil
}

// Terminate stops and removes the Redis container.
func (c *RedisContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

// KafkaContainer wraps a testcontainers Kafka container.
type KafkaContainer struct {
	Container testcontainers.Container
	Brokers   []string
}

// NewKafkaContainer starts a Kafka container for testing.
// Returns the container wrapper with broker addresses for connecting.
// Caller must call Terminate() when done.
// Uses Redpanda (Kafka-compatible) for faster startup and simpler config.
func NewKafkaContainer(ctx context.Context) (*KafkaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redpandadata/redpanda:v24.1.1",
		ExposedPorts: []string{"9092/tcp"},
		Cmd: []string{
			"redpanda", "start",
			"--smp", "1",
			"--memory", "512M",
			"--reserve-memory", "0M",
			"--overprovisioned",
			"--node-id", "0",
			"--kafka-addr", "PLAINTEXT://0.0.0.0:9092",
			"--advertise-kafka-addr", "PLAINTEXT://localhost:9092",
			"--set", "redpanda.auto_create_topics_enabled=true",
		},
		WaitingFor: wait.ForListeningPort("9092/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("start kafka container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "9092")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get port: %w", err)
	}

	broker := fmt.Sprintf("%s:%s", host, port.Port())

	return &KafkaContainer{
		Container: container,
		Brokers:   []string{broker},
	}, nil
}

// Terminate stops and removes the Kafka container.
func (c *KafkaContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

// RabbitMQContainer wraps a testcontainers RabbitMQ container.
type RabbitMQContainer struct {
	Container testcontainers.Container
	URL       string // AMQP connection URL
}

// NewRabbitMQContainer starts a RabbitMQ container for testing.
// Returns the container wrapper with AMQP URL for connecting.
// Caller must call Terminate() when done.
func NewRabbitMQContainer(ctx context.Context) (*RabbitMQContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.13-management",
		ExposedPorts: []string{"5672/tcp", "15672/tcp"},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": "guest",
			"RABBITMQ_DEFAULT_PASS": "guest",
		},
		WaitingFor: wait.ForLog("Server startup complete").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("start rabbitmq container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5672")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get port: %w", err)
	}

	url := fmt.Sprintf("amqp://guest:guest@%s:%s/", host, port.Port())

	return &RabbitMQContainer{
		Container: container,
		URL:       url,
	}, nil
}

// Terminate stops and removes the RabbitMQ container.
func (c *RabbitMQContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}
