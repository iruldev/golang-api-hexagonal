package runtimeutil

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// =============================================================================
// Event Tests
// =============================================================================

func TestNewEvent(t *testing.T) {
	type testPayload struct {
		UserID string `json:"user_id"`
		Action string `json:"action"`
	}

	tests := []struct {
		name      string
		eventType string
		payload   interface{}
		wantErr   bool
	}{
		{
			name:      "valid struct payload",
			eventType: "user.created",
			payload:   testPayload{UserID: "123", Action: "signup"},
			wantErr:   false,
		},
		{
			name:      "valid map payload",
			eventType: "order.completed",
			payload:   map[string]string{"order_id": "456"},
			wantErr:   false,
		},
		{
			name:      "empty event type",
			eventType: "",
			payload:   map[string]string{"key": "value"},
			wantErr:   false,
		},
		{
			name:      "nil payload",
			eventType: "notification.sent",
			payload:   nil,
			wantErr:   false,
		},
		{
			name:      "unmarshalable payload",
			eventType: "test.event",
			payload:   make(chan int), // channels cannot be marshalled
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewEvent(tt.eventType, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEvent() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewEvent() unexpected error: %v", err)
				return
			}

			if event.ID == "" {
				t.Error("NewEvent() event.ID is empty")
			}
			if event.Type != tt.eventType {
				t.Errorf("NewEvent() event.Type = %v, want %v", event.Type, tt.eventType)
			}
			if event.Timestamp.IsZero() {
				t.Error("NewEvent() event.Timestamp is zero")
			}
		})
	}
}

// =============================================================================
// EventPublisher Tests
// =============================================================================

func TestNopEventPublisher_ImplementsInterface(t *testing.T) {
	var _ EventPublisher = &NopEventPublisher{}
	var _ EventPublisher = NewNopEventPublisher()
}

func TestNopEventPublisher_Publish(t *testing.T) {
	publisher := NewNopEventPublisher()
	event, _ := NewEvent("test.event", map[string]string{"key": "value"})

	err := publisher.Publish(context.Background(), "test-topic", event)
	if err != nil {
		t.Errorf("NopEventPublisher.Publish() error = %v, want nil", err)
	}
}

func TestNopEventPublisher_PublishAsync(t *testing.T) {
	publisher := NewNopEventPublisher()
	event, _ := NewEvent("test.event", map[string]string{"key": "value"})

	err := publisher.PublishAsync(context.Background(), "test-topic", event)
	if err != nil {
		t.Errorf("NopEventPublisher.PublishAsync() error = %v, want nil", err)
	}
}

// =============================================================================
// EventConsumer Interface Tests
// =============================================================================

func TestNopEventConsumer_ImplementsInterface(t *testing.T) {
	var _ EventConsumer = &NopEventConsumer{}
	var _ EventConsumer = NewNopEventConsumer()
}

func TestNopEventConsumer_Subscribe(t *testing.T) {
	consumer := NewNopEventConsumer()
	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	err := consumer.Subscribe(context.Background(), "test-topic", handler)
	if err != nil {
		t.Errorf("NopEventConsumer.Subscribe() error = %v, want nil", err)
	}
}

func TestNopEventConsumer_Close(t *testing.T) {
	consumer := NewNopEventConsumer()

	err := consumer.Close()
	if err != nil {
		t.Errorf("NopEventConsumer.Close() error = %v, want nil", err)
	}
}

func TestNopEventConsumer_SubscribeAfterClose(t *testing.T) {
	consumer := NewNopEventConsumer()
	_ = consumer.Close()

	err := consumer.Subscribe(context.Background(), "test", func(ctx context.Context, event Event) error {
		return nil
	})

	if !errors.Is(err, ErrConsumerClosed) {
		t.Errorf("NopEventConsumer.Subscribe() after Close = %v, want ErrConsumerClosed", err)
	}
}

// =============================================================================
// MockEventConsumer Tests
// =============================================================================

func TestMockEventConsumer_ImplementsInterface(t *testing.T) {
	var _ EventConsumer = &MockEventConsumer{}
	var _ EventConsumer = NewMockEventConsumer()
}

func TestMockEventConsumer_SimulateEvent(t *testing.T) {
	mock := NewMockEventConsumer()
	var receivedEvent Event
	var handlerCalled bool

	handler := func(ctx context.Context, event Event) error {
		handlerCalled = true
		receivedEvent = event
		return nil
	}

	// Start subscription in goroutine since it blocks
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mock.Subscribe(ctx, "orders", handler)
	}()

	// Give time for goroutine to start
	time.Sleep(10 * time.Millisecond)

	// Simulate an event
	testEvent, _ := NewEvent("order.created", map[string]string{"order_id": "123"})
	err := mock.SimulateEvent(testEvent)
	if err != nil {
		t.Errorf("MockEventConsumer.SimulateEvent() error = %v, want nil", err)
	}

	if !handlerCalled {
		t.Error("MockEventConsumer.SimulateEvent() handler was not called")
	}

	if receivedEvent.ID != testEvent.ID {
		t.Errorf("MockEventConsumer.SimulateEvent() event.ID = %v, want %v", receivedEvent.ID, testEvent.ID)
	}

	cancel()
}

func TestMockEventConsumer_HandlerCalled(t *testing.T) {
	mock := NewMockEventConsumer()

	if mock.HandlerCalled() {
		t.Error("MockEventConsumer.HandlerCalled() = true before any calls, want false")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mock.Subscribe(ctx, "test", func(ctx context.Context, event Event) error {
			return nil
		})
	}()

	time.Sleep(10 * time.Millisecond)

	testEvent, _ := NewEvent("test.event", nil)
	_ = mock.SimulateEvent(testEvent)

	if !mock.HandlerCalled() {
		t.Error("MockEventConsumer.HandlerCalled() = false after SimulateEvent, want true")
	}

	cancel()
}

func TestMockEventConsumer_LastEvent(t *testing.T) {
	mock := NewMockEventConsumer()

	// Before any events
	lastEvent := mock.LastEvent()
	if lastEvent.ID != "" {
		t.Error("MockEventConsumer.LastEvent() should return zero Event before any calls")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mock.Subscribe(ctx, "test", func(ctx context.Context, event Event) error {
			return nil
		})
	}()

	time.Sleep(10 * time.Millisecond)

	event1, _ := NewEvent("event.first", nil)
	event2, _ := NewEvent("event.second", nil)
	_ = mock.SimulateEvent(event1)
	_ = mock.SimulateEvent(event2)

	lastEvent = mock.LastEvent()
	if lastEvent.ID != event2.ID {
		t.Errorf("MockEventConsumer.LastEvent() ID = %v, want %v", lastEvent.ID, event2.ID)
	}

	cancel()
}

func TestMockEventConsumer_Events(t *testing.T) {
	mock := NewMockEventConsumer()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mock.Subscribe(ctx, "test", func(ctx context.Context, event Event) error {
			return nil
		})
	}()

	time.Sleep(10 * time.Millisecond)

	event1, _ := NewEvent("event.1", nil)
	event2, _ := NewEvent("event.2", nil)
	event3, _ := NewEvent("event.3", nil)
	_ = mock.SimulateEvent(event1)
	_ = mock.SimulateEvent(event2)
	_ = mock.SimulateEvent(event3)

	events := mock.Events()
	if len(events) != 3 {
		t.Errorf("MockEventConsumer.Events() len = %d, want 3", len(events))
	}

	cancel()
}

func TestMockEventConsumer_Topic(t *testing.T) {
	mock := NewMockEventConsumer()
	expectedTopic := "my-topic"

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mock.Subscribe(ctx, expectedTopic, func(ctx context.Context, event Event) error {
			return nil
		})
	}()

	time.Sleep(10 * time.Millisecond)

	if mock.Topic() != expectedTopic {
		t.Errorf("MockEventConsumer.Topic() = %v, want %v", mock.Topic(), expectedTopic)
	}

	cancel()
}

func TestMockEventConsumer_SimulateEventWithoutSubscribe(t *testing.T) {
	mock := NewMockEventConsumer()
	testEvent, _ := NewEvent("test.event", nil)

	err := mock.SimulateEvent(testEvent)
	if err == nil {
		t.Error("MockEventConsumer.SimulateEvent() should return error when no handler subscribed")
	}
}

func TestMockEventConsumer_SubscribeAfterClose(t *testing.T) {
	mock := NewMockEventConsumer()
	_ = mock.Close()

	err := mock.Subscribe(context.Background(), "test", func(ctx context.Context, event Event) error {
		return nil
	})

	if !errors.Is(err, ErrConsumerClosed) {
		t.Errorf("MockEventConsumer.Subscribe() after Close = %v, want ErrConsumerClosed", err)
	}
}

func TestMockEventConsumer_Close(t *testing.T) {
	mock := NewMockEventConsumer()
	ctx := context.Background()
	done := make(chan struct{})

	go func() {
		_ = mock.Subscribe(ctx, "test", func(ctx context.Context, event Event) error {
			return nil
		})
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)

	err := mock.Close()
	if err != nil {
		t.Errorf("MockEventConsumer.Close() error = %v, want nil", err)
	}

	select {
	case <-done:
		// Subscribe returned after Close - success
	case <-time.After(100 * time.Millisecond):
		t.Error("MockEventConsumer.Close() did not unblock Subscribe")
	}
}

func TestMockEventConsumer_HandlerError(t *testing.T) {
	mock := NewMockEventConsumer()
	expectedErr := errors.New("handler error")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mock.Subscribe(ctx, "test", func(ctx context.Context, event Event) error {
			return expectedErr
		})
	}()

	time.Sleep(10 * time.Millisecond)

	testEvent, _ := NewEvent("test.event", nil)
	err := mock.SimulateEvent(testEvent)

	if !errors.Is(err, expectedErr) {
		t.Errorf("MockEventConsumer.SimulateEvent() error = %v, want %v", err, expectedErr)
	}

	cancel()
}

// =============================================================================
// ConsumerConfig Tests
// =============================================================================

func TestDefaultConsumerConfig(t *testing.T) {
	config := DefaultConsumerConfig()

	if config.MaxRetries != 3 {
		t.Errorf("DefaultConsumerConfig().MaxRetries = %d, want 3", config.MaxRetries)
	}

	if config.Concurrency != 1 {
		t.Errorf("DefaultConsumerConfig().Concurrency = %d, want 1", config.Concurrency)
	}

	if config.ProcessingTimeout != 30*time.Second {
		t.Errorf("DefaultConsumerConfig().ProcessingTimeout = %v, want 30s", config.ProcessingTimeout)
	}

	if !config.AutoAck {
		t.Error("DefaultConsumerConfig().AutoAck = false, want true")
	}

	if config.GroupID != "" {
		t.Errorf("DefaultConsumerConfig().GroupID = %v, want empty", config.GroupID)
	}
}

func TestConsumerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ConsumerConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConsumerConfig(),
			wantErr: false,
		},
		{
			name:    "valid custom config",
			config:  ConsumerConfig{MaxRetries: 5, Concurrency: 10, ProcessingTimeout: time.Minute},
			wantErr: false,
		},
		{
			name:    "invalid negative MaxRetries",
			config:  ConsumerConfig{MaxRetries: -1, Concurrency: 1},
			wantErr: true,
		},
		{
			name:    "invalid zero Concurrency",
			config:  ConsumerConfig{MaxRetries: 3, Concurrency: 0},
			wantErr: true,
		},
		{
			name:    "invalid negative ProcessingTimeout",
			config:  ConsumerConfig{MaxRetries: 3, Concurrency: 1, ProcessingTimeout: -time.Second},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConsumerConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Sentinel Error Tests
// =============================================================================

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrConsumerClosed",
			err:  ErrConsumerClosed,
			want: "consumer closed",
		},
		{
			name: "ErrProcessingTimeout",
			err:  ErrProcessingTimeout,
			want: "processing timeout exceeded",
		},
		{
			name: "ErrMaxRetriesExceeded",
			err:  ErrMaxRetriesExceeded,
			want: "max retries exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %v, want %v", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	if errors.Is(ErrConsumerClosed, ErrProcessingTimeout) {
		t.Error("ErrConsumerClosed should not equal ErrProcessingTimeout")
	}
	if errors.Is(ErrConsumerClosed, ErrMaxRetriesExceeded) {
		t.Error("ErrConsumerClosed should not equal ErrMaxRetriesExceeded")
	}
	if errors.Is(ErrProcessingTimeout, ErrMaxRetriesExceeded) {
		t.Error("ErrProcessingTimeout should not equal ErrMaxRetriesExceeded")
	}
}

func TestSentinelErrorsCanBeWrapped(t *testing.T) {
	wrappedErr := fmt.Errorf("wrapper: %w", ErrConsumerClosed)
	if !errors.Is(wrappedErr, ErrConsumerClosed) {
		t.Error("Wrapped error should match original via errors.Is")
	}
}
