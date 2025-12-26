# Data Models

**Project:** golang-api-hexagonal  
**Database:** PostgreSQL 15+  
**Migration Tool:** Goose v3.26.0  

## Overview

This document describes the database schema and data models used by the application.

## Schema Diagram

```
┌─────────────────────────────────────┐
│            schema_info              │
├─────────────────────────────────────┤
│ id          SERIAL PRIMARY KEY      │
│ version     VARCHAR(50)             │
│ description TEXT                    │
│ initialized_at TIMESTAMPTZ          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│              users                   │
├─────────────────────────────────────┤
│ id          UUID PRIMARY KEY        │
│ email       VARCHAR(255) NOT NULL   │
│ first_name  VARCHAR(100) NOT NULL   │
│ last_name   VARCHAR(100) NOT NULL   │
│ created_at  TIMESTAMPTZ NOT NULL    │
│ updated_at  TIMESTAMPTZ NOT NULL    │
├─────────────────────────────────────┤
│ UNIQUE INDEX: email                 │
│ INDEX: created_at                   │
└─────────────────────────────────────┘
          │
          │ entity_id (FK concept, not enforced)
          ▼
┌─────────────────────────────────────┐
│          audit_events                │
├─────────────────────────────────────┤
│ id          UUID PRIMARY KEY        │
│ event_type  VARCHAR(100) NOT NULL   │
│ actor_id    UUID (nullable)         │
│ entity_type VARCHAR(50) NOT NULL    │
│ entity_id   UUID NOT NULL           │
│ payload     JSONB NOT NULL          │
│ timestamp   TIMESTAMPTZ NOT NULL    │
│ request_id  VARCHAR(50)             │
├─────────────────────────────────────┤
│ INDEX: event_type                   │
│ INDEX: entity_type                  │
│ INDEX: (entity_type, entity_id, timestamp DESC) │
│ INDEX: timestamp DESC               │
└─────────────────────────────────────┘
```

---

## Tables

### schema_info

**Purpose:** Application-level schema versioning (separate from Goose's internal tracking).

**Migration:** `20251216000000_init.sql`

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | SERIAL | No | auto | Primary key |
| `version` | VARCHAR(50) | No | '0.0.1' | Schema version |
| `description` | TEXT | Yes | - | Version description |
| `initialized_at` | TIMESTAMPTZ | No | NOW() | Initialization timestamp |

**Indexes:** None (single row table)

---

### users

**Purpose:** Store user account information.

**Migration:** `20251217000000_create_users.sql`

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | No | - | Primary key (application-generated) |
| `email` | VARCHAR(255) | No | - | User email address |
| `first_name` | VARCHAR(100) | No | - | User first name |
| `last_name` | VARCHAR(100) | No | - | User last name |
| `created_at` | TIMESTAMPTZ | No | - | Creation timestamp |
| `updated_at` | TIMESTAMPTZ | No | - | Last update timestamp |

**Indexes:**

| Name | Columns | Type | Purpose |
|------|---------|------|---------|
| `uniq_users_email` | email | UNIQUE | Prevent duplicate emails |
| `idx_users_created_at` | created_at | B-tree | Sort by creation date |

**Domain Entity:** `internal/domain/user.go`

```go
type User struct {
    ID        ID
    Email     string
    FirstName string
    LastName  string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

---

### audit_events

**Purpose:** Store audit trail for business operations.

**Migration:** `20251219000000_create_audit_events.sql`

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | UUID | No | - | Primary key (application-generated) |
| `event_type` | VARCHAR(100) | No | - | Event type (e.g., "user.created") |
| `actor_id` | UUID | Yes | - | User who performed action (null for system) |
| `entity_type` | VARCHAR(50) | No | - | Type of entity affected (e.g., "user") |
| `entity_id` | UUID | No | - | ID of affected entity |
| `payload` | JSONB | No | - | Event data (PII-redacted) |
| `timestamp` | TIMESTAMPTZ | No | - | When event occurred |
| `request_id` | VARCHAR(50) | Yes | - | Correlation ID from HTTP request |

**Indexes:**

| Name | Columns | Type | Purpose |
|------|---------|------|---------|
| `idx_audit_events_event_type` | event_type | B-tree | Filter by event type |
| `idx_audit_events_entity_type` | entity_type | B-tree | Filter by entity type |
| `idx_audit_events_entity_time` | entity_type, entity_id, timestamp DESC | Composite | Lookup audit history for entity |
| `idx_audit_events_timestamp` | timestamp DESC | B-tree | Time-based queries |

**Domain Entity:** `internal/domain/audit.go`

```go
type AuditEvent struct {
    ID         ID
    EventType  string    // "entity.action" format
    ActorID    ID        // Nullable
    EntityType string
    EntityID   ID
    Payload    []byte    // JSON, PII-redacted
    Timestamp  time.Time
    RequestID  string    // Nullable
}
```

**Event Type Constants:**

```go
const (
    EventUserCreated = "user.created"
    EventUserUpdated = "user.updated"
    EventUserDeleted = "user.deleted"
)
```

---

## Goose Internal Table

> **Note:** This table is managed by Goose, not by application code.

### goose_db_version

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key |
| `version_id` | BIGINT | Migration version number |
| `is_applied` | BOOLEAN | Whether migration is applied |
| `tstamp` | TIMESTAMPTZ | When migration was applied |

---

## Migration Files

| File | Version | Description |
|------|---------|-------------|
| `20251216000000_init.sql` | 0.0.1 | Initial schema setup |
| `20251217000000_create_users.sql` | - | Users table |
| `20251219000000_create_audit_events.sql` | - | Audit events table |

### Migration Commands

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Show migration status
make migrate-status

# Create new migration
make migrate-create name=add_roles_table

# Validate migration syntax
make migrate-validate
```

---

## Data Access Layer

### Repository Implementations

| Repository | File | Methods |
|------------|------|---------|
| UserRepository | `internal/infra/postgres/user_repo.go` | Create, GetByID, List |
| AuditEventRepository | `internal/infra/postgres/audit_repo.go` | Create, ListByEntityID |

### Transaction Support

All repositories accept a `Querier` interface, allowing them to work with both:
- Connection pool (for single operations)
- Transaction (for atomic multi-operation workflows)

```go
// Transaction pattern
txManager.WithTx(ctx, func(tx domain.Querier) error {
    if err := userRepo.Create(ctx, tx, user); err != nil {
        return err // Rollback
    }
    if err := auditRepo.Create(ctx, tx, event); err != nil {
        return err // Rollback
    }
    return nil // Commit
})
```

---

## ID Generation

- **Type:** UUID v4
- **Generator:** `internal/infra/postgres/id_generator.go`
- **Library:** `github.com/google/uuid`

All entity IDs are generated by the application, not the database. This allows:
- Predictable IDs in tests
- ID generation without database round-trip
- Consistency across distributed systems

---

## PII Handling

Audit event payloads undergo PII redaction before storage:

| Field | Redaction Mode | Example |
|-------|---------------|---------|
| Email (full) | `[REDACTED]` | `[REDACTED]` |
| Email (partial) | First 2 chars + domain | `jo***@example.com` |

Configured via `AUDIT_REDACT_EMAIL` environment variable.
