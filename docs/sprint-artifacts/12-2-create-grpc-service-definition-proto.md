# Story 12.2: Create gRPC Service Definition (Proto)

Status: Done

## Story

As a developer,
I want to define service contracts via Protobuf,
So that interfaces are strongly typed and versioned.

## Acceptance Criteria

1. **Given** `proto/note/v1/note.proto` exists
   **When** I run `make gen-proto`
   **Then** Go code interfaces are generated
   **And** generated code compiles without errors

2. **Given** proto file defines NoteService
   **When** code is generated
   **Then** I can implement the server interface in `internal/interface/grpc/note`
   **And** implementation follows hexagonal architecture (calls usecase layer)

3. **Given** proto file is properly structured
   **When** proto file is reviewed
   **Then** it includes proper package naming (`note.v1`)
   **And** includes Go package option pointing to generated location
   **And** includes proper versioning for API evolution

4. **Given** NoteService proto definition
   **When** service is defined
   **Then** it includes CRUD operations: CreateNote, GetNote, ListNotes, UpdateNote, DeleteNote
   **And** request/response messages follow protobuf best practices

## Tasks / Subtasks

- [x] Task 1: Create Proto File Structure (AC: #1, #3)
  - [x] Create `proto/note/v1/note.proto` file
  - [x] Add package declaration: `package note.v1;`
  - [x] Add Go package option: `option go_package = "github.com/iruldev/golang-api-hexagonal/proto/note/v1;notev1";`
  - [x] Add proto3 syntax declaration

- [x] Task 2: Define Message Types (AC: #4)
  - [x] Create `Note` message with fields: id (string), title (string), content (string), created_at (google.protobuf.Timestamp), updated_at (google.protobuf.Timestamp)
  - [x] Create `CreateNoteRequest` message (title, content)
  - [x] Create `CreateNoteResponse` message (Note)
  - [x] Create `GetNoteRequest` message (id)
  - [x] Create `GetNoteResponse` message (Note)
  - [x] Create `ListNotesRequest` message (page_size, page_token)
  - [x] Create `ListNotesResponse` message (repeated Note, next_page_token, total_count)
  - [x] Create `UpdateNoteRequest` message (id, title, content)
  - [x] Create `UpdateNoteResponse` message (Note)
  - [x] Create `DeleteNoteRequest` message (id)
  - [x] Create `DeleteNoteResponse` message (empty)

- [x] Task 3: Define NoteService (AC: #2, #4)
  - [x] Define `service NoteService` block
  - [x] Add `rpc CreateNote(CreateNoteRequest) returns (CreateNoteResponse)`
  - [x] Add `rpc GetNote(GetNoteRequest) returns (GetNoteResponse)`
  - [x] Add `rpc ListNotes(ListNotesRequest) returns (ListNotesResponse)`
  - [x] Add `rpc UpdateNote(UpdateNoteRequest) returns (UpdateNoteResponse)`
  - [x] Add `rpc DeleteNote(DeleteNoteRequest) returns (DeleteNoteResponse)`

- [x] Task 4: Update Makefile for Proto Generation (AC: #1)
  - [x] Verify `make gen-proto` target works with new proto file
  - [x] Ensure generated code goes to `proto/note/v1/` directory
  - [x] Run `make gen-proto` and verify no errors

- [x] Task 5: Implement gRPC Note Handler (AC: #2)
  - [x] Create `internal/interface/grpc/note/handler.go`
  - [x] Implement `NoteServiceServer` interface
  - [x] Inject `usecase.NoteUsecase` dependency (hexagonal pattern)
  - [x] Implement CreateNote calling usecase
  - [x] Implement GetNote calling usecase
  - [x] Implement ListNotes calling usecase
  - [x] Implement UpdateNote calling usecase
  - [x] Implement DeleteNote calling usecase
  - [x] Map domain errors to gRPC status codes

- [x] Task 6: Register gRPC Service with Server (AC: #2)
  - [x] Update `cmd/server/main.go` to register NoteService
  - [x] Call `notev1.RegisterNoteServiceServer(grpcSrv.GRPCServer(), noteHandler)`
  - [x] Ensure handler is only registered when gRPC is enabled and database available

- [x] Task 7: Write Unit Tests (AC: #1-4)
  - [x] Test generated proto compiles (build successful)
  - [x] Test handler with mock usecase (MockRepository pattern)
  - [x] Test error mapping to gRPC status codes (TestMapErrorToStatus)
  - [x] Test all CRUD operations (18 test cases, all passing)

- [x] Task 8: Update Documentation (AC: #1)
  - [x] Update `AGENTS.md` with proto patterns section
  - [x] Document proto file naming conventions
  - [x] Document gRPC handler implementation pattern
  - [x] Document error mapping to gRPC status codes

## Dev Notes

### Architecture Compliance

gRPC handlers MUST call usecase layer, NOT infra layer directly:

```
┌─────────────────────────────────────────────────────────────────┐
│                      Interface Layer                             │
│  ┌─────────────────┐    ┌─────────────────┐                     │
│  │   HTTP Handler  │    │  gRPC Handler   │  ← This Story       │
│  │  note/handler   │    │  note/handler   │                     │
│  └────────┬────────┘    └────────┬────────┘                     │
│           │                      │                              │
│           └──────────┬───────────┘                              │
│                      │                                          │
│  ┌───────────────────▼───────────────────┐                      │
│  │            Usecase Layer              │                      │
│  │        note.NoteUsecase               │  ← SHARED            │
│  └───────────────────┬───────────────────┘                      │
│                      │                                          │
│  ┌───────────────────▼───────────────────┐                      │
│  │            Domain Layer               │                      │
│  │        note.Repository (interface)    │                      │
│  └───────────────────┬───────────────────┘                      │
│                      │                                          │
│  ┌───────────────────▼───────────────────┐                      │
│  │            Infra Layer                │                      │
│  │        postgres/note (adapter)        │                      │
│  └───────────────────────────────────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

### Proto File Structure

```protobuf
syntax = "proto3";

package note.v1;

option go_package = "github.com/iruldev/golang-api-hexagonal/proto/note/v1;notev1";

import "google/protobuf/timestamp.proto";

message Note {
  string id = 1;
  string title = 2;
  string content = 3;
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}

// ... other messages and service definition
```

### gRPC Handler Pattern

```go
// internal/interface/grpc/note/handler.go
package note

import (
    "context"
    
    notev1 "github.com/iruldev/golang-api-hexagonal/proto/note/v1"
    noteuc "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type Handler struct {
    notev1.UnimplementedNoteServiceServer
    usecase *noteuc.Usecase
}

func NewHandler(uc *noteuc.Usecase) *Handler {
    return &Handler{usecase: uc}
}

func (h *Handler) CreateNote(ctx context.Context, req *notev1.CreateNoteRequest) (*notev1.CreateNoteResponse, error) {
    // Call usecase, NOT infra
    note, err := h.usecase.Create(ctx, req.Title, req.Content)
    if err != nil {
        return nil, mapErrorToStatus(err)
    }
    return &notev1.CreateNoteResponse{
        Note: toProto(note),
    }, nil
}

func mapErrorToStatus(err error) error {
    switch {
    case errors.Is(err, notedom.ErrNoteNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, notedom.ErrEmptyTitle):
        return status.Error(codes.InvalidArgument, err.Error())
    default:
        return status.Error(codes.Internal, "internal error")
    }
}
```

### Error Mapping Table

| Domain Error | gRPC Status Code |
|--------------|------------------|
| `ErrNoteNotFound` | `codes.NotFound` |
| `ErrEmptyTitle` | `codes.InvalidArgument` |
| `ErrEmptyContent` | `codes.InvalidArgument` |
| Validation errors | `codes.InvalidArgument` |
| DB errors | `codes.Internal` |
| Unknown | `codes.Internal` |

### Previous Story Intelligence (Story 12.1)

From Story 12.1 implementation:
- gRPC server is in `internal/interface/grpc/server.go`
- Interceptor chain: RequestID → Metrics → Logging → Recovery
- Use `grpcSrv.GRPCServer()` to get underlying `*grpc.Server` for registration
- Config in `internal/config/config.go` - GRPCConfig struct
- Reflection enabled via `GRPC_REFLECTION_ENABLED`
- Test with grpcurl: `grpcurl -plaintext localhost:50051 list`

### Library Versions

| Package | Version | Purpose |
|---------|---------|---------|
| `google.golang.org/grpc` | v1.64.0+ | gRPC framework |
| `google.golang.org/protobuf` | v1.34.0+ | Protocol buffers |

### Testing with grpcurl

```bash
# After implementation, test with:
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 describe note.v1.NoteService

# Create a note
grpcurl -plaintext -d '{"title": "Test", "content": "Hello gRPC"}' \
  localhost:50051 note.v1.NoteService/CreateNote

# Get a note
grpcurl -plaintext -d '{"id": "uuid-here"}' \
  localhost:50051 note.v1.NoteService/GetNote
```

### References

- [Source: docs/epics.md#Story-12.2] - Story requirements
- [Source: docs/architecture.md#Hexagonal-Architecture] - Layer boundaries
- [Source: AGENTS.md#gRPC-Server-Patterns] - gRPC patterns from Story 12.1
- [Source: docs/sprint-artifacts/12-1-add-grpc-server-support.md] - Previous story learnings

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- ✅ Created comprehensive proto file `proto/note/v1/note.proto` with full CRUD operations
- ✅ Generated Go code successfully with `make gen-proto`
- ✅ Implemented `NoteServiceServer` in `internal/interface/grpc/note/handler.go` following hexagonal architecture
- ✅ Created repository adapter `internal/infra/postgres/note/repository.go` to bridge sqlc queries to domain interface
- ✅ Added factory function `postgres.NewNoteRepository()` for clean dependency injection
- ✅ Registered NoteService in `main.go` with conditional database availability check
- ✅ All 18 unit tests passing with table-driven tests covering all CRUD operations and error mapping
- ✅ Updated AGENTS.md with Proto File Patterns section including naming conventions, examples, and best practices
- ⚠️ Note: golangci-lint has an internal panic in printf analyzer (not related to this story's code)

#### Code Review Fixes (2025-12-14)

- ✅ Fixed repository to use sqlc-generated `ListNotes` and `CountNotes` instead of raw SQL (FR21 compliance)
- ✅ Fixed `Delete` method to use `:execrows` instead of extra `GetNote` call (eliminates DB round-trip)
- ✅ Updated `db/queries/note.sql` with `:execrows` annotation for DeleteNote
- ✅ Regenerated sqlc code with `make gen`
- ✅ All 18 tests still passing after fixes

### Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | SM Agent | Story created with comprehensive context for gRPC service definition |
| 2025-12-14 | Dev Agent | Implemented all tasks: proto file, handler, tests, documentation |
| 2025-12-14 | Code Review | Fixed raw SQL usage in List, optimized Delete to use :execrows |

### File List

- [NEW] `proto/note/v1/note.proto` - Protobuf service definition with NoteService CRUD operations
- [NEW] `proto/note/v1/note.pb.go` - Generated protobuf Go code
- [NEW] `proto/note/v1/note_grpc.pb.go` - Generated gRPC service interface
- [NEW] `internal/interface/grpc/note/handler.go` - gRPC handler implementing NoteServiceServer
- [NEW] `internal/interface/grpc/note/handler_test.go` - 18 unit tests for handler
- [NEW] `internal/infra/postgres/note/repository.go` - PostgreSQL repository adapter (uses sqlc-generated queries)
- [NEW] `internal/infra/postgres/repositories.go` - Factory function for note repository
- [MOD] `cmd/server/main.go` - Register NoteService with gRPC server
- [MOD] `AGENTS.md` - Add Proto File Patterns section with conventions and examples
- [MOD] `db/queries/note.sql` - Changed DeleteNote to :execrows for efficient delete check
- [MOD] `internal/infra/postgres/note/note.sql.go` - Regenerated with ListNotes, CountNotes, execrows DeleteNote
