package patterns_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/iruldev/golang-api-hexagonal/internal/worker"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
	"go.uber.org/zap"
)

// Example_basicFanout demonstrates basic fanout usage with multiple handlers.
func Example_basicFanout() {
	// Create a new registry
	registry := patterns.NewFanoutRegistry()

	// Register handlers for the "user:created" event
	_ = registry.Register("user:created", "welcome-email", func(ctx context.Context, event patterns.FanoutEvent) error {
		fmt.Println("Sending welcome email")
		return nil
	})

	_ = registry.Register("user:created", "default-settings", func(ctx context.Context, event patterns.FanoutEvent) error {
		fmt.Println("Creating default settings")
		return nil
	})

	_ = registry.Register("user:created", "notify-admin", func(ctx context.Context, event patterns.FanoutEvent) error {
		fmt.Println("Notifying admin")
		return nil
	})

	// Check registered handlers
	handlers := registry.Handlers("user:created")
	fmt.Printf("Registered %d handlers for user:created\n", len(handlers))
	// Output: Registered 3 handlers for user:created
}

// Example_fanoutWithQueuePriorities demonstrates using different queues for handlers.
func Example_fanoutWithQueuePriorities() {
	registry := patterns.NewFanoutRegistry()

	// Critical handler - use critical queue
	_ = registry.RegisterWithQueue("order:completed", "send-receipt",
		func(ctx context.Context, event patterns.FanoutEvent) error {
			return nil
		},
		worker.QueueCritical,
	)

	// Normal handler - use default queue
	_ = registry.Register("order:completed", "update-inventory",
		func(ctx context.Context, event patterns.FanoutEvent) error {
			return nil
		},
	)

	// Low priority handler - use low queue
	_ = registry.RegisterWithQueue("order:completed", "analytics",
		func(ctx context.Context, event patterns.FanoutEvent) error {
			return nil
		},
		worker.QueueLow,
	)

	handlers := registry.Handlers("order:completed")
	for _, h := range handlers {
		fmt.Printf("Handler %s uses queue: %s\n", h.ID, h.Queue)
	}
	// Output:
	// Handler send-receipt uses queue: critical
	// Handler update-inventory uses queue: default
	// Handler analytics uses queue: low
}

// Example_fanoutEvent demonstrates creating and publishing fanout events.
func Example_fanoutEvent() {
	// Create an event with typed payload
	userPayload, _ := json.Marshal(map[string]interface{}{
		"user_id": "123",
		"email":   "user@example.com",
		"name":    "John Doe",
	})

	event := patterns.FanoutEvent{
		Type:    "user:created",
		Payload: userPayload,
		Metadata: map[string]string{
			"trace_id": "abc-123",
			"source":   "api",
		},
		// Timestamp is auto-set if zero
	}

	fmt.Printf("Event type: %s\n", event.Type)
	fmt.Printf("Has metadata: %v\n", len(event.Metadata) > 0)
	// Output:
	// Event type: user:created
	// Has metadata: true
}

// Example_workerRegistration demonstrates how to register fanout handlers in the worker.
func Example_workerRegistration() {
	// In cmd/worker/main.go:

	// 1. Create shared registry (same instance used by API and Worker)
	registry := patterns.NewFanoutRegistry()
	logger := zap.NewNop()

	// 2. Register handlers
	_ = registry.Register("note:archived", "update-search", func(ctx context.Context, event patterns.FanoutEvent) error {
		// Update search index
		return nil
	})

	_ = registry.Register("note:archived", "notify-owner", func(ctx context.Context, event patterns.FanoutEvent) error {
		// Send notification
		return nil
	})

	// 3. Create dispatcher
	dispatcher := patterns.NewFanoutDispatcher(registry, logger)

	// 4. The dispatcher.Handle method would be registered with the worker server
	// In real code: srv.HandleFunc(patterns.TaskTypePrefix()+"*", dispatcher.Handle)
	fmt.Printf("Dispatcher created with %d handlers\n", len(registry.Handlers("note:archived")))
	fmt.Printf("Task type prefix: %s\n", patterns.TaskTypePrefix())

	_ = dispatcher // Prevent unused variable error
	// Output:
	// Dispatcher created with 2 handlers
	// Task type prefix: fanout:
}

// Example_dynamicUnregister demonstrates dynamically removing handlers.
func Example_dynamicUnregister() {
	registry := patterns.NewFanoutRegistry()

	// Register handlers
	_ = registry.Register("test:event", "handler-a", func(ctx context.Context, event patterns.FanoutEvent) error { return nil })
	_ = registry.Register("test:event", "handler-b", func(ctx context.Context, event patterns.FanoutEvent) error { return nil })
	_ = registry.Register("test:event", "handler-c", func(ctx context.Context, event patterns.FanoutEvent) error { return nil })

	fmt.Printf("Before unregister: %d handlers\n", len(registry.Handlers("test:event")))

	// Unregister handler-b
	registry.Unregister("test:event", "handler-b")

	handlers := registry.Handlers("test:event")
	fmt.Printf("After unregister: %d handlers\n", len(handlers))
	for _, h := range handlers {
		fmt.Printf("- %s\n", h.ID)
	}
	// Output:
	// Before unregister: 3 handlers
	// After unregister: 2 handlers
	// - handler-a
	// - handler-c
}

// Example_handlerWithOptions demonstrates adding retry and timeout options to handlers.
func Example_handlerWithOptions() {
	registry := patterns.NewFanoutRegistry()

	// Handler with custom retry and timeout
	_ = registry.Register("email:send", "smtp-delivery",
		func(ctx context.Context, event patterns.FanoutEvent) error {
			return nil
		},
		asynq.MaxRetry(5),
		asynq.Timeout(30_000_000_000), // 30 seconds in nanoseconds
	)

	handlers := registry.Handlers("email:send")
	fmt.Printf("Handler has %d options\n", len(handlers[0].Opts))
	// Output: Handler has 2 options
}
