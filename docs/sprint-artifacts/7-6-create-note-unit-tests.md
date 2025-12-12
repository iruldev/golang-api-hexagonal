# Story 7.6: Create Note Unit Tests

Status: done

## Story

As a developer,
I want example unit tests,
So that I can understand testing patterns.

## Acceptance Criteria

### AC1: Unit tests implemented
**Given** `*_test.go` files exist in note packages
**When** I run `make test`
**Then** tests pass with ≥70% coverage
**And** table-driven tests with AAA pattern are used
**And** mocks are used for dependencies

---

## Tasks / Subtasks

- [x] **Task 1: Create domain/note tests** (AC: #1)
  - [x] Test Note entity validation
  - [x] Test NewNote constructor
  - [x] Test Update method

- [x] **Task 2: Create usecase/note tests** (AC: #1)
  - [x] Mock Repository interface
  - [x] Test Create (valid, validation error)
  - [x] Test Get (found, not found)
  - [x] Test List (with pagination)
  - [x] Test Update (valid, not found)
  - [x] Test Delete

- [x] **Task 3: Create handler tests** (AC: #1)
  - [x] Mock Usecase via mock repo
  - [x] Test all handlers (Create, Get, List, Update, Delete)
  - [x] Test error responses

- [x] **Task 4: Verify coverage** (AC: #1)
  - [x] Run `go test` - all pass
  - [x] Table-driven tests
  - [x] AAA pattern

---

## Dev Notes

### Testing Patterns

```go
// Table-driven test with AAA pattern
func TestUsecase_Create(t *testing.T) {
    tests := []struct {
        name    string
        title   string
        content string
        wantErr error
    }{
        {"valid", "Title", "Content", nil},
        {"empty title", "", "Content", note.ErrEmptyTitle},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            repo := &MockRepository{}
            uc := NewUsecase(repo)

            // Act
            _, err := uc.Create(context.Background(), tt.title, tt.content)

            // Assert
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### File List

Files to create:
- `internal/domain/note/entity_test.go`
- `internal/usecase/note/usecase_test.go`
- `internal/interface/http/note/handler_test.go`
