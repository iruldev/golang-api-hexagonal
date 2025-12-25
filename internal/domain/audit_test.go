package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuditEvent_Fields(t *testing.T) {
	timestamp := time.Now()
	payload := []byte(`{"email":"[REDACTED]","name":"John Doe"}`)

	event := AuditEvent{
		ID:         ID("audit-event-123"),
		EventType:  "user.created",
		ActorID:    ID("actor-456"),
		EntityType: "user",
		EntityID:   ID("entity-789"),
		Payload:    payload,
		Timestamp:  timestamp,
		RequestID:  "req-abc-123",
	}

	// Verify all fields are properly assigned
	assert.Equal(t, ID("audit-event-123"), event.ID)
	assert.Equal(t, "user.created", event.EventType)
	assert.Equal(t, ID("actor-456"), event.ActorID)
	assert.Equal(t, "user", event.EntityType)
	assert.Equal(t, ID("entity-789"), event.EntityID)
	assert.Equal(t, payload, event.Payload)
	assert.Equal(t, timestamp, event.Timestamp)
	assert.Equal(t, "req-abc-123", event.RequestID)
}

func TestAuditEvent_Fields_EmptyActorID(t *testing.T) {
	// ActorID is optional (empty for system/unauthenticated events)
	event := AuditEvent{
		ID:         ID("audit-event-123"),
		EventType:  "user.created",
		ActorID:    ID(""), // Empty - system-initiated event
		EntityType: "user",
		EntityID:   ID("entity-789"),
		Payload:    []byte(`{}`),
		Timestamp:  time.Now(),
		RequestID:  "req-abc-123",
	}

	assert.True(t, event.ActorID.IsEmpty())
	// Should still be valid since ActorID is optional
	assert.NoError(t, event.Validate())
}

func TestAuditEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   AuditEvent
		wantErr error
	}{
		{
			name: "valid event with all fields",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "user.created",
				ActorID:    ID("actor-id"),
				EntityType: "user",
				EntityID:   ID("entity-id"),
				Payload:    []byte(`{"key":"value"}`),
				Timestamp:  time.Now(),
				RequestID:  "req-123",
			},
			wantErr: nil,
		},
		{
			name: "valid event with empty actor (system event)",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "user.created",
				ActorID:    ID(""), // Empty is valid for system events
				EntityType: "user",
				EntityID:   ID("entity-id"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid event with minimal fields",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "order.placed",
				EntityType: "order",
				EntityID:   ID("order-123"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "missing ID",
			event: AuditEvent{
				ID:         ID(""),
				EventType:  "user.created",
				EntityType: "user",
				EntityID:   ID("entity-id"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "missing event type",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "",
				EntityType: "user",
				EntityID:   ID("123"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: ErrInvalidEventType,
		},
		{
			name: "whitespace-only event type",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "   ",
				EntityType: "user",
				EntityID:   ID("123"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: ErrInvalidEventType,
		},
		{
			name: "missing entity type",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "user.created",
				EntityType: "",
				EntityID:   ID("123"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: ErrInvalidEntityType,
		},
		{
			name: "whitespace-only entity type",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "user.created",
				EntityType: "   ",
				EntityID:   ID("123"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: ErrInvalidEntityType,
		},
		{
			name: "missing entity ID",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "user.created",
				EntityType: "user",
				EntityID:   ID(""),
				Payload:    []byte(`{}`),
				Timestamp:  time.Now(),
			},
			wantErr: ErrInvalidEntityID,
		},
		{
			name: "nil payload",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "user.created",
				EntityType: "user",
				EntityID:   ID("entity-id"),
				Payload:    nil,
				Timestamp:  time.Now(),
			},
			wantErr: ErrInvalidPayload,
		},
		{
			name: "zero timestamp",
			event: AuditEvent{
				ID:         ID("test-id"),
				EventType:  "user.created",
				EntityType: "user",
				EntityID:   ID("entity-id"),
				Payload:    []byte(`{}`),
				Timestamp:  time.Time{}, // Zero time
			},
			wantErr: ErrInvalidTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEventTypeConstants(t *testing.T) {
	// Verify event type constants follow "entity.action" pattern
	assert.Equal(t, "user.created", EventUserCreated)
	assert.Equal(t, "user.updated", EventUserUpdated)
	assert.Equal(t, "user.deleted", EventUserDeleted)

	// Verify they can be used in AuditEvent
	event := AuditEvent{
		ID:         ID("test-id"),
		EventType:  EventUserCreated,
		EntityType: "user",
		EntityID:   ID("user-123"),
		Payload:    []byte(`{}`),
		Timestamp:  time.Now(),
	}
	assert.NoError(t, event.Validate())
}

func TestAuditEvent_PayloadBytes(t *testing.T) {
	// Verify Payload can store arbitrary JSON bytes
	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "simple JSON object",
			payload: []byte(`{"key": "value"}`),
		},
		{
			name:    "JSON with nested objects",
			payload: []byte(`{"user": {"email": "[REDACTED]", "name": "John"}}`),
		},
		{
			name:    "JSON array",
			payload: []byte(`["item1", "item2"]`),
		},
		{
			name:    "empty JSON object",
			payload: []byte(`{}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := AuditEvent{
				ID:         ID("test-id"),
				EventType:  EventUserCreated,
				EntityType: "user",
				EntityID:   ID("user-123"),
				Payload:    tt.payload,
				Timestamp:  time.Now(),
			}
			assert.Equal(t, tt.payload, event.Payload)
			assert.NoError(t, event.Validate())
		})
	}
}

func TestAuditEvent_RequestID(t *testing.T) {
	// RequestID is a string (not domain.ID) since it comes from transport layer
	event := AuditEvent{
		ID:         ID("test-id"),
		EventType:  EventUserCreated,
		EntityType: "user",
		EntityID:   ID("user-123"),
		Payload:    []byte(`{}`),
		Timestamp:  time.Now(),
		RequestID:  "request-correlation-id-abc123",
	}

	assert.Equal(t, "request-correlation-id-abc123", event.RequestID)
	assert.NoError(t, event.Validate())

	// Empty RequestID is also valid (not all events have request correlation)
	eventNoReqID := AuditEvent{
		ID:         ID("test-id"),
		EventType:  EventUserCreated,
		EntityType: "user",
		EntityID:   ID("user-123"),
		Payload:    []byte(`{}`),
		Timestamp:  time.Now(),
		RequestID:  "",
	}
	assert.NoError(t, eventNoReqID.Validate())
}

func TestAuditEvent_RequestID_Validation(t *testing.T) {
	// RequestID max length is 64
	longRequestID := "this-request-id-is-way-too-long-and-should-fail-validation-because-it-exceeds-64-chars-now-verifying-that-it-fails"
	event := AuditEvent{
		ID:         ID("test-id"),
		EventType:  EventUserCreated,
		EntityType: "user",
		EntityID:   ID("user-123"),
		Payload:    []byte(`{}`),
		Timestamp:  time.Now(),
		RequestID:  longRequestID,
	}

	assert.ErrorIs(t, event.Validate(), ErrInvalidRequestID)
}
