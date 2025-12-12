# Story 7.5: Create Note HTTP Handler

Status: done

## Story

As a developer,
I want an example HTTP handler,
So that I can understand handler patterns.

## Acceptance Criteria

### AC1: Note HTTP handler implemented
**Given** `internal/interface/http/note/handler.go` exists
**When** I review the code
**Then** NoteHandler depends on NoteUsecase interface
**And** routes: POST, GET, GET/:id, PUT/:id, DELETE/:id
**And** uses response envelope pattern

---

## Tasks / Subtasks

- [x] **Task 1: Create NoteHandler** (AC: #1)
  - [x] Define handler struct with usecase dependency
  - [x] Implement CreateNote handler (POST /api/v1/notes)
  - [x] Implement GetNote handler (GET /api/v1/notes/:id)
  - [x] Implement ListNotes handler (GET /api/v1/notes)
  - [x] Implement UpdateNote handler (PUT /api/v1/notes/:id)
  - [x] Implement DeleteNote handler (DELETE /api/v1/notes/:id)

- [x] **Task 2: Register routes** (AC: #1)
  - [x] Add note routes to router via Routes() method
  - [x] Chi router pattern

- [x] **Task 3: Use response envelope** (AC: #1)
  - [x] Use response.Success for success responses
  - [x] Use response.Error for error responses

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Handler Structure

```go
// internal/interface/http/note/handler.go
type Handler struct {
    usecase *note.Usecase
}

func NewHandler(u *note.Usecase) *Handler {
    return &Handler{usecase: u}
}

func (h *Handler) Routes(r chi.Router) {
    r.Route("/notes", func(r chi.Router) {
        r.Post("/", h.Create)
        r.Get("/", h.List)
        r.Get("/{id}", h.Get)
        r.Put("/{id}", h.Update)
        r.Delete("/{id}", h.Delete)
    })
}
```

### Request/Response DTOs

```go
type CreateNoteRequest struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

type NoteResponse struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### File List

Files to create:
- `internal/interface/http/note/handler.go` - Note HTTP handler
- `internal/interface/http/note/dto.go` - Request/Response DTOs
