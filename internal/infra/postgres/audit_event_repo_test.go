//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
)

func TestAuditEventRepo_Create(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewAuditEventRepo()
	querier := postgres.NewPoolQuerier(&dbAdapter{p: pool})

	id, err := uuid.NewV7()
	require.NoError(t, err)

	entityID, err := uuid.NewV7()
	require.NoError(t, err)

	actorID, err := uuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	event := &domain.AuditEvent{
		ID:         domain.ID(id.String()),
		EventType:  domain.EventUserCreated,
		ActorID:    domain.ID(actorID.String()),
		EntityType: "user",
		EntityID:   domain.ID(entityID.String()),
		Payload:    []byte(`{"email":"[REDACTED]"}`),
		Timestamp:  now,
		RequestID:  "req-123",
	}

	err = repo.Create(ctx, querier, event)
	assert.NoError(t, err)

	// Verify via ListByEntityID
	events, count, err := repo.ListByEntityID(ctx, querier, "user", event.EntityID, domain.ListParams{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	require.Len(t, events, 1)
	assert.Equal(t, event.EventType, events[0].EventType)
	assert.Equal(t, event.ActorID, events[0].ActorID)
	assert.Equal(t, event.EntityType, events[0].EntityType)
	assert.Equal(t, event.EntityID, events[0].EntityID)
	assert.Equal(t, event.RequestID, events[0].RequestID)
}

func TestAuditEventRepo_Create_WithNullActorID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewAuditEventRepo()
	querier := postgres.NewPoolQuerier(&dbAdapter{p: pool})

	id, _ := uuid.NewV7()
	entityID, _ := uuid.NewV7()

	now := time.Now().UTC().Truncate(time.Microsecond)
	event := &domain.AuditEvent{
		ID:         domain.ID(id.String()),
		EventType:  "system.scheduled_task",
		ActorID:    "", // Empty = system event
		EntityType: "job",
		EntityID:   domain.ID(entityID.String()),
		Payload:    []byte(`{"task":"cleanup"}`),
		Timestamp:  now,
		RequestID:  "cron-456",
	}

	err := repo.Create(ctx, querier, event)
	assert.NoError(t, err)

	// Verify ActorID is empty when retrieved
	events, _, err := repo.ListByEntityID(ctx, querier, "job", event.EntityID, domain.ListParams{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	require.Len(t, events, 1)
	assert.Empty(t, events[0].ActorID)
}

func TestAuditEventRepo_ListByEntityID_WithPagination(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewAuditEventRepo()
	querier := postgres.NewPoolQuerier(&dbAdapter{p: pool})

	entityID, _ := uuid.NewV7()
	actorID, _ := uuid.NewV7()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create 25 audit events for the same entity
	for i := 0; i < 25; i++ {
		id, _ := uuid.NewV7()
		event := &domain.AuditEvent{
			ID:         domain.ID(id.String()),
			EventType:  domain.EventUserUpdated,
			ActorID:    domain.ID(actorID.String()),
			EntityType: "user",
			EntityID:   domain.ID(entityID.String()),
			Payload:    []byte(`{"update":"field"}`),
			Timestamp:  now.Add(time.Duration(i) * time.Second),
			RequestID:  "req-" + string(rune('a'+i)),
		}
		err := repo.Create(ctx, querier, event)
		require.NoError(t, err)
	}

	// Test first page
	params := domain.ListParams{Page: 1, PageSize: 10}
	events, totalCount, err := repo.ListByEntityID(ctx, querier, "user", domain.ID(entityID.String()), params)
	assert.NoError(t, err)
	assert.Equal(t, 25, totalCount)
	assert.Len(t, events, 10)

	// Test second page
	params = domain.ListParams{Page: 2, PageSize: 10}
	events, totalCount, err = repo.ListByEntityID(ctx, querier, "user", domain.ID(entityID.String()), params)
	assert.NoError(t, err)
	assert.Equal(t, 25, totalCount)
	assert.Len(t, events, 10)

	// Test third page (partial)
	params = domain.ListParams{Page: 3, PageSize: 10}
	events, totalCount, err = repo.ListByEntityID(ctx, querier, "user", domain.ID(entityID.String()), params)
	assert.NoError(t, err)
	assert.Equal(t, 25, totalCount)
	assert.Len(t, events, 5)
}

func TestAuditEventRepo_ListByEntityID_OrderByTimestampDesc(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewAuditEventRepo()
	querier := postgres.NewPoolQuerier(&dbAdapter{p: pool})

	entityID, _ := uuid.NewV7()
	actorID, _ := uuid.NewV7()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create first event (older)
	id1, _ := uuid.NewV7()
	event1 := &domain.AuditEvent{
		ID:         domain.ID(id1.String()),
		EventType:  domain.EventUserCreated,
		ActorID:    domain.ID(actorID.String()),
		EntityType: "user",
		EntityID:   domain.ID(entityID.String()),
		Payload:    []byte(`{"action":"create"}`),
		Timestamp:  now,
		RequestID:  "req-old",
	}
	require.NoError(t, repo.Create(ctx, querier, event1))

	// Create second event (newer)
	id2, _ := uuid.NewV7()
	event2 := &domain.AuditEvent{
		ID:         domain.ID(id2.String()),
		EventType:  domain.EventUserUpdated,
		ActorID:    domain.ID(actorID.String()),
		EntityType: "user",
		EntityID:   domain.ID(entityID.String()),
		Payload:    []byte(`{"action":"update"}`),
		Timestamp:  now.Add(10 * time.Second), // 10 seconds later
		RequestID:  "req-new",
	}
	require.NoError(t, repo.Create(ctx, querier, event2))

	// List should return newer event first
	params := domain.ListParams{Page: 1, PageSize: 10}
	events, _, err := repo.ListByEntityID(ctx, querier, "user", domain.ID(entityID.String()), params)
	assert.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, domain.EventUserUpdated, events[0].EventType) // Newest first
	assert.Equal(t, domain.EventUserCreated, events[1].EventType)
}

func TestAuditEventRepo_ListByEntityID_FiltersByEntityTypeAndID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewAuditEventRepo()
	querier := postgres.NewPoolQuerier(&dbAdapter{p: pool})

	// Create events for different entities
	entityID1, _ := uuid.NewV7()
	entityID2, _ := uuid.NewV7()
	actorID, _ := uuid.NewV7()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Event for entity 1 (user)
	id1, _ := uuid.NewV7()
	event1 := &domain.AuditEvent{
		ID:         domain.ID(id1.String()),
		EventType:  domain.EventUserCreated,
		ActorID:    domain.ID(actorID.String()),
		EntityType: "user",
		EntityID:   domain.ID(entityID1.String()),
		Payload:    []byte(`{}`),
		Timestamp:  now,
		RequestID:  "req-1",
	}
	require.NoError(t, repo.Create(ctx, querier, event1))

	// Event for entity 2 (order - different type)
	id2, _ := uuid.NewV7()
	event2 := &domain.AuditEvent{
		ID:         domain.ID(id2.String()),
		EventType:  "order.created",
		ActorID:    domain.ID(actorID.String()),
		EntityType: "order",
		EntityID:   domain.ID(entityID2.String()),
		Payload:    []byte(`{}`),
		Timestamp:  now,
		RequestID:  "req-2",
	}
	require.NoError(t, repo.Create(ctx, querier, event2))

	// Event for entity 1 (user) - second event
	id3, _ := uuid.NewV7()
	event3 := &domain.AuditEvent{
		ID:         domain.ID(id3.String()),
		EventType:  domain.EventUserUpdated,
		ActorID:    domain.ID(actorID.String()),
		EntityType: "user",
		EntityID:   domain.ID(entityID1.String()),
		Payload:    []byte(`{}`),
		Timestamp:  now.Add(time.Second),
		RequestID:  "req-3",
	}
	require.NoError(t, repo.Create(ctx, querier, event3))

	// Query only for user entity 1 - should get 2 events
	params := domain.ListParams{Page: 1, PageSize: 10}
	events, count, err := repo.ListByEntityID(ctx, querier, "user", domain.ID(entityID1.String()), params)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, events, 2)

	// Query for order entity 2 - should get 1 event
	events, count, err = repo.ListByEntityID(ctx, querier, "order", domain.ID(entityID2.String()), params)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, events, 1)
	assert.Equal(t, "order.created", events[0].EventType)

	// Query for non-existent entity - should get 0 events
	nonExistent, _ := uuid.NewV7()
	events, count, err = repo.ListByEntityID(ctx, querier, "user", domain.ID(nonExistent.String()), params)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, events)
}

func TestAuditEventRepo_ListByEntityID_OrderByIDDescWhenTimestampEqual(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewAuditEventRepo()
	querier := postgres.NewPoolQuerier(&dbAdapter{p: pool})

	entityID, _ := uuid.NewV7()
	actorID, _ := uuid.NewV7()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create two events with the same timestamp to verify the id DESC tie-breaker
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	event1 := &domain.AuditEvent{
		ID:         domain.ID(id1.String()),
		EventType:  domain.EventUserCreated,
		ActorID:    domain.ID(actorID.String()),
		EntityType: "user",
		EntityID:   domain.ID(entityID.String()),
		Payload:    []byte(`{"tie":"1"}`),
		Timestamp:  now,
		RequestID:  "req-tie-1",
	}
	require.NoError(t, repo.Create(ctx, querier, event1))

	event2 := &domain.AuditEvent{
		ID:         domain.ID(id2.String()),
		EventType:  domain.EventUserUpdated,
		ActorID:    domain.ID(actorID.String()),
		EntityType: "user",
		EntityID:   domain.ID(entityID.String()),
		Payload:    []byte(`{"tie":"2"}`),
		Timestamp:  now, // Same timestamp
		RequestID:  "req-tie-2",
	}
	require.NoError(t, repo.Create(ctx, querier, event2))

	params := domain.ListParams{Page: 1, PageSize: 10}
	events, _, err := repo.ListByEntityID(ctx, querier, "user", domain.ID(entityID.String()), params)
	require.NoError(t, err)
	require.Len(t, events, 2)

	// Same timestamp, so id DESC should put id2 before id1
	assert.Equal(t, domain.ID(id2.String()), events[0].ID)
	assert.Equal(t, domain.ID(id1.String()), events[1].ID)
}

func TestAuditEventRepo_ListByEntityID_InvalidID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewAuditEventRepo()
	querier := postgres.NewPoolQuerier(&dbAdapter{p: pool})

	_, _, err := repo.ListByEntityID(ctx, querier, "user", "invalid-uuid", domain.ListParams{Page: 1, PageSize: 10})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse entityID")
}
