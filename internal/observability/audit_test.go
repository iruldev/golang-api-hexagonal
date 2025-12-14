package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaskSensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "no sensitive data",
			input: map[string]any{
				"username": "jdoe",
				"email":    "jdoe@example.com",
			},
			expected: map[string]any{
				"username": "jdoe",
				"email":    "jdoe@example.com",
			},
		},
		{
			name: "sensitive data",
			input: map[string]any{
				"password": "secret123",
				"Token":    "xyz-token",
				"API_KEY":  "12345",
				"other":    "value",
			},
			expected: map[string]any{
				"password": "***MASKED***",
				"Token":    "***MASKED***",
				"API_KEY":  "***MASKED***",
				"other":    "value",
			},
		},
		{
			name: "nested sensitive data",
			input: map[string]any{
				"details": map[string]any{
					"password": "nested-secret",
					"public":   "visible",
				},
			},
			expected: map[string]any{
				"details": map[string]any{
					"password": "***MASKED***",
					"public":   "visible",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskSensitive(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestNewAuditEvent(t *testing.T) {
	ctx := context.Background()

	// With no auth claims (passed explicitly as "anonymous")
	event := NewAuditEvent(ctx, ActionCreate, "note:123", "anonymous", nil)
	assert.Equal(t, ActionCreate, event.Action)
	assert.Equal(t, "note:123", event.Resource)
	assert.Equal(t, "anonymous", event.ActorID)
	assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)

	// With explicit actorID and Status default
	eventWithAuth := NewAuditEvent(ctx, ActionUpdate, "note:456", "user-1", nil)
	assert.Equal(t, "user-1", eventWithAuth.ActorID)
	assert.Equal(t, "success", eventWithAuth.Status) // Default status

	// With metadata masking
	metadata := map[string]any{"password": "pw"}
	eventWithMeta := NewAuditEvent(ctx, ActionLogin, "user:1", "user-1", metadata)
	assert.Equal(t, "***MASKED***", eventWithMeta.Metadata["password"])

	// RequestID is no longer auto-extracted
	eventWithReq := NewAuditEvent(ctx, ActionCreate, "note:789", "user-1", nil)
	assert.Empty(t, eventWithReq.RequestID)
	// Caller sets it manually
	eventWithReq.RequestID = "req-123"
	assert.Equal(t, "req-123", eventWithReq.RequestID)
}
