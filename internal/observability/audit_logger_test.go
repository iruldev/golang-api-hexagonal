package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogAudit(t *testing.T) {
	// Create an observer core to capture logs
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	ctx := context.Background()
	event := AuditEvent{
		Action:   ActionCreate,
		Resource: "note:123",
		ActorID:  "user:1",
		Metadata: map[string]any{"key": "value"},
	}

	// Act
	LogAudit(ctx, logger, event)

	// Assert
	entries := observedLogs.All()
	assert.Len(t, entries, 1)
	entry := entries[0]

	assert.Equal(t, "Audit Event", entry.Message)

	// Check fields
	fields := entry.ContextMap()
	assert.Equal(t, "audit", fields["event_type"])
	assert.Equal(t, "create", fields["audit_action"])
	assert.Equal(t, "note:123", fields["audit_resource"])
	assert.Equal(t, "user:1", fields["audit_actor_id"])
	// Status won't be present in the struct literal unless initialized,
	// but LogAudit logs event.Status which is empty string if not set.
	assert.Equal(t, "", fields["audit_status"])

	// Metadata might be nested map
	// Metadata might be nested map
	meta := fields["audit_metadata"].(map[string]interface{})
	// zaptest observer might flatten or keep as map, depending on encoding.
	// In-memory core usually keeps implementation types.
	assert.Equal(t, "value", meta["key"])
}

func TestLogAudit_NilLogger(t *testing.T) {
	ctx := context.Background()
	event := AuditEvent{Action: ActionCreate}

	// Should not panic
	LogAudit(ctx, nil, event)
}
