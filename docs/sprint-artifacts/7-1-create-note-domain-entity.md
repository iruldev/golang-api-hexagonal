# Story 7.1: Create Note Domain Entity

Status: done

## Story

As a developer,
I want an example entity with validation,
So that I can understand how to create domain entities.

## Acceptance Criteria

### AC1: Note entity defined
**Given** `internal/domain/note/entity.go` exists
**When** I review the code
**Then** Note entity has ID, Title, Content, CreatedAt, UpdatedAt
**And** validation method returns domain errors
**And** entity is documented with comments

---

## Tasks / Subtasks

- [x] **Task 1: Create Note entity** (AC: #1)
  - [x] Create Note struct with ID, Title, Content, CreatedAt, UpdatedAt
  - [x] Add JSON and DB tags

- [x] **Task 2: Add validation method** (AC: #1)
  - [x] Create Validate() method
  - [x] Return domain errors for invalid data

- [x] **Task 3: Add errors.go** (AC: #1)
  - [x] Create domain-specific errors (ErrEmptyTitle, etc.)

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Note Entity

```go
// internal/domain/note/entity.go
package note

import (
    "time"

    "github.com/google/uuid"
)

// Note represents a note in the system.
type Note struct {
    ID        uuid.UUID `json:"id" db:"id"`
    Title     string    `json:"title" db:"title"`
    Content   string    `json:"content" db:"content"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate validates the Note entity.
func (n *Note) Validate() error {
    if n.Title == "" {
        return ErrEmptyTitle
    }
    if len(n.Title) > 255 {
        return ErrTitleTooLong
    }
    return nil
}
```

### Domain Errors

```go
// internal/domain/note/errors.go
var (
    ErrEmptyTitle   = errors.New("note: title cannot be empty")
    ErrTitleTooLong = errors.New("note: title exceeds maximum length")
    ErrNoteNotFound = errors.New("note: not found")
)
```

### File List

Files to create:
- `internal/domain/note/entity.go` - Note entity
- `internal/domain/note/errors.go` - Domain errors
