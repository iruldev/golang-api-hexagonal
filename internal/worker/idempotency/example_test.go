package idempotency_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/worker/idempotency"
)

// Example_basicUsage shows how to create an idempotent handler wrapper.
func Example_basicUsage() {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// Create idempotency store
	store := idempotency.NewRedisStore(client, "idempotency:")

	// Define key extractor - determines what makes a task unique
	keyExtractor := func(t *asynq.Task) string {
		var payload struct {
			OrderID string `json:"order_id"`
		}
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return "" // No idempotency if payload is invalid
		}
		return fmt.Sprintf("order:%s", payload.OrderID)
	}

	// Original handler
	originalHandler := func(ctx context.Context, t *asynq.Task) error {
		fmt.Println("Processing order...")
		return nil
	}

	// Wrap with idempotency
	handler := idempotency.IdempotentHandler(
		store,
		keyExtractor,
		24*time.Hour, // TTL for idempotency key
		originalHandler,
	)

	// Use wrapped handler with asynq server
	_ = handler // srv.HandleFunc("order:create", handler)
}

// Example_customTTL shows how to configure a custom TTL for idempotency keys.
func Example_customTTL() {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	store := idempotency.NewRedisStore(client, "idempotency:")

	keyExtractor := func(t *asynq.Task) string {
		return t.Type() // Use task type as key (simple example)
	}

	handler := func(ctx context.Context, t *asynq.Task) error {
		return nil
	}

	// Short TTL for time-sensitive operations (1 hour)
	shortTTLHandler := idempotency.IdempotentHandler(
		store, keyExtractor, 1*time.Hour, handler,
	)

	// Long TTL for operations that shouldn't repeat (7 days)
	longTTLHandler := idempotency.IdempotentHandler(
		store, keyExtractor, 7*24*time.Hour, handler,
	)

	_, _ = shortTTLHandler, longTTLHandler
}

// Example_failModes shows how to configure fail-open vs fail-closed behavior.
func Example_failModes() {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	logger := zap.NewNop()

	// Fail-open store (default) - process task even if Redis is down
	failOpenStore := idempotency.NewRedisStore(client, "idempotency:",
		idempotency.WithFailMode(idempotency.FailOpen),
		idempotency.WithLogger(logger),
	)

	// Fail-closed store - return error if Redis is down (task will be retried)
	failClosedStore := idempotency.NewRedisStore(client, "idempotency:",
		idempotency.WithFailMode(idempotency.FailClosed),
		idempotency.WithLogger(logger),
	)

	_, _ = failOpenStore, failClosedStore
}

// Example_keyExtractionStrategies shows different ways to extract idempotency keys.
func Example_keyExtractionStrategies() {
	// Strategy 1: Payload-based - extract specific field from payload
	payloadBasedExtractor := func(t *asynq.Task) string {
		var p struct {
			OrderID   string `json:"order_id"`
			ProductID string `json:"product_id"`
		}
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return ""
		}
		return fmt.Sprintf("order:%s:product:%s", p.OrderID, p.ProductID)
	}

	// Strategy 2: Task-type + entity ID
	taskTypeExtractor := func(t *asynq.Task) string {
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return ""
		}
		return fmt.Sprintf("%s:%s", t.Type(), p.ID)
	}

	// Strategy 3: Hash-based - for complex payloads
	hashBasedExtractor := func(t *asynq.Task) string {
		h := sha256.Sum256(t.Payload())
		return fmt.Sprintf("%s:%x", t.Type(), h[:8]) // First 8 bytes of hash
	}

	// Strategy 4: Business key - domain-specific
	invoiceExtractor := func(t *asynq.Task) string {
		var p struct {
			InvoiceNumber string `json:"invoice_number"`
			Year          int    `json:"year"`
		}
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return ""
		}
		return fmt.Sprintf("invoice:%d-%s", p.Year, p.InvoiceNumber)
	}

	_, _, _, _ = payloadBasedExtractor, taskTypeExtractor, hashBasedExtractor, invoiceExtractor
}

// Example_workerIntegration shows complete worker setup with idempotency.
func Example_workerIntegration() {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	logger := zap.NewNop()

	// Create idempotency store
	store := idempotency.NewRedisStore(client, "idempotency:",
		idempotency.WithFailMode(idempotency.FailOpen),
		idempotency.WithLogger(logger),
	)

	// Define handlers
	orderHandler := func(ctx context.Context, t *asynq.Task) error {
		fmt.Println("Processing order...")
		return nil
	}

	paymentHandler := func(ctx context.Context, t *asynq.Task) error {
		fmt.Println("Processing payment...")
		return nil
	}

	// Wrap with idempotency
	idempotentOrderHandler := idempotency.IdempotentHandler(
		store,
		func(t *asynq.Task) string {
			var p struct {
				OrderID string `json:"order_id"`
			}
			json.Unmarshal(t.Payload(), &p)
			return fmt.Sprintf("order:%s", p.OrderID)
		},
		24*time.Hour,
		orderHandler,
		idempotency.WithHandlerLogger(logger),
	)

	idempotentPaymentHandler := idempotency.IdempotentHandler(
		store,
		func(t *asynq.Task) string {
			var p struct {
				PaymentID string `json:"payment_id"`
			}
			json.Unmarshal(t.Payload(), &p)
			return fmt.Sprintf("payment:%s", p.PaymentID)
		},
		24*time.Hour,
		paymentHandler,
		idempotency.WithHandlerLogger(logger),
		idempotency.WithHandlerFailMode(idempotency.FailClosed), // Critical operation
	)

	// Register with asynq server
	// srv.HandleFunc("order:create", idempotentOrderHandler)
	// srv.HandleFunc("payment:process", idempotentPaymentHandler)

	_, _ = idempotentOrderHandler, idempotentPaymentHandler
}

// Example_combineWithAsynqUnique shows how to use both asynq.Unique and idempotency package.
func Example_combineWithAsynqUnique() {
	// For maximum protection, combine both approaches:

	// 1. Use asynq.Unique when enqueuing (prevents duplicate enqueue)
	task := asynq.NewTask("order:create", []byte(`{"order_id": "123"}`))
	_ = task // client.Enqueue(task, asynq.Unique(24*time.Hour))

	// 2. Wrap handler with idempotency (prevents duplicate processing on retry)
	// handler := idempotency.IdempotentHandler(store, keyExtractor, 24*time.Hour, originalHandler)
	// srv.HandleFunc("order:create", handler)

	// This provides protection at both enqueue and processing levels:
	// - asynq.Unique: Prevents same task from being queued twice
	// - idempotency package: Prevents same task from being processed twice (e.g., on retry)
}
