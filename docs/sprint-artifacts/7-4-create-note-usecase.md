# Story 7.4: Create Note Usecase

Status: done

## Story

As a developer,
I want an example usecase implementation,
So that I can understand business logic patterns.

## Acceptance Criteria

### AC1: Note usecase implemented
**Given** `internal/usecase/note/usecase.go` exists
**When** I review the code
**Then** NoteUsecase depends on NoteRepository interface
**And** methods include: Create, Get, List, Update, Delete
**And** business logic validates before persistence

---

## Tasks / Subtasks

- [x] **Task 1: Define NoteRepository interface** (AC: #1)
  - [x] Define in domain/note package
  - [x] Include Create, Get, List, Update, Delete methods

- [x] **Task 2: Create NoteUsecase** (AC: #1)
  - [x] Create usecase struct with repository dependency
  - [x] Implement Create (validate → persist)
  - [x] Implement Get, List, Update, Delete

- [x] **Task 3: Add business logic validation** (AC: #1)
  - [x] Validate entity before Create/Update
  - [x] Return domain errors

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### NoteRepository Interface

```go
// internal/domain/note/repository.go
type Repository interface {
    Create(ctx context.Context, note *Note) error
    Get(ctx context.Context, id uuid.UUID) (*Note, error)
    List(ctx context.Context, limit, offset int) ([]*Note, error)
    Update(ctx context.Context, note *Note) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### NoteUsecase

```go
// internal/usecase/note/usecase.go
type Usecase struct {
    repo note.Repository
}

func (u *Usecase) Create(ctx context.Context, title, content string) (*note.Note, error) {
    n := note.NewNote(title, content)
    if err := n.Validate(); err != nil {
        return nil, err // domain error
    }
    if err := u.repo.Create(ctx, n); err != nil {
        return nil, err
    }
    return n, nil
}
```

### File List

Files to create:
- `internal/domain/note/repository.go` - Repository interface
- `internal/usecase/note/usecase.go` - Note usecase
