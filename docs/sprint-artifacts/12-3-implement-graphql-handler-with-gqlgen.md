# Story 12.3: Implement GraphQL Handler with GQLGen

Status: done

## Story

As a frontend developer,
I want a GraphQL endpoint,
So that I can fetch flexible data dependencies in one request.

## Acceptance Criteria

1. **Given** `gqlgen.yml` configured and available
   **When** I run `make gen` (or `make gen-gql`)
   **Then** resolvers and server code are generated in `internal/interface/graphql`
   **And** code compiles without errors

2. **Given** schema defined in `api/graphql/schema.graphqls` matches Note domain
   **When** I query the endpoint
   **Then** I can fetch Notes data (Get, List)
   **And** I can mutate Notes data (Create, Update, Delete)

3. **Given** a GraphQL request
   **When** business logic error occurs (e.g., Not Found, Validation)
   **Then** standard GraphQL error format is returned
   **And** errors map correctly to error codes

4. **Given** `internal/interface/graphql` directory
   **When** implementation is reviewed
   **Then** it follows hexagonal architecture (resolvers call usecase layer)
   **And** no direct infrastructure dependencies in resolvers

## Tasks / Subtasks

- [x] Task 1: Initialize GQLGen & Configuration (AC: #1)
  - [x] Add `github.com/99designs/gqlgen` dependency
  - [x] Create `gqlgen.yml` configuration
  - [x] Configure output to `internal/interface/graphql`
  - [x] Configure schema path to `api/graphql/*.graphqls`
  - [x] Update `make gen` to include graphql generation

- [x] Task 2: Define GraphQL Schema (AC: #2)
  - [x] Create `api/graphql/schema.graphqls`
  - [x] Define `Note` type matching domain entity (ID, Title, Content, timestamps)
  - [x] Define `Query` type with `notes` (list) and `note` (server-side filtering if needed, or by ID)
  - [x] Define `Mutation` type with `createNote`, `updateNote`, `deleteNote`
  - [x] Define necessary Input types (`CreateNoteInput`, etc.)

- [x] Task 3: Implement Resolvers (AC: #2, #4)
  - [x] Generate initial resolver shells
  - [x] Inject `usecase.NoteUsecase` into Resolver struct
  - [x] Implement `notes` query resolver calling `usecase.List`
  - [x] Implement `note` query resolver calling `usecase.Get`
  - [x] Implement `createNote` resolver calling `usecase.Create`
  - [x] Implement `updateNote` resolver calling `usecase.Update`
  - [x] Implement `deleteNote` resolver calling `usecase.Delete`

- [x] Task 4: Error Handling & Mapping (AC: #3)
  - [x] Create helper to map `domain.AppError` to `graphql.Error` (or gqlgen primitives)
  - [x] Ensure `NotFound` returns null with error or appropriate GQL pattern

- [x] Task 5: Server Integration (AC: #2)
  - [x] Update `cmd/server/main.go`
  - [x] Initialize GraphQL handler `handler.NewDefaultServer(...)`
  - [x] Register route `/query` (and optional `/playground` if nice to have, though 12.4 is dedicated)
  - [x] Ensure middleware (Auth, Logging) applies to GraphQL endpoint

- [x] Task 6: Tests
  - [x] Test schema generation works
  - [x] Unit test resolvers (mock usecase)
  - [x] Integration test: sending querying to `/query`

- [x] Task 7: Documentation
  - [x] Update `AGENTS.md` with GraphQL patterns
  - [x] Document how to add new GraphQL schemas/resolvers

## Dev Notes

### Architecture Compliance

The GraphQL interface lives in `internal/interface/graphql`.
It acts as an adapter, translating GraphQL requests into Usecase interactions.

**Dependencies:**
`internal/interface/graphql` -> `internal/usecase/note`
`internal/interface/graphql` -> `internal/domain/note` (models)

**Configuration:**
`gqlgen.yml` should attempt to reuse `internal/domain/note` models if they match.
If not, let gqlgen generate models in `internal/interface/graphql/model` and map them manually.

### Schema Design

Follow standard GraphQL naming conventions (camelCase fields).
Map `apple_case` or `snake_case` from DB/JSON to GraphQL `camelCase` if needed, but `gqlgen` handles snake_case struct tags well usually.

### References

- [Source: docs/epics.md#Story-12.3]
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [GQLGen Docs](https://gqlgen.com/)

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- Implemented `internal/interface/graphql/errors.go` for domain error mapping.
- Implemented resolvers for Create, Update, Delete, List, Get.
- Added comprehensive integration tests in `internal/interface/graphql/integration_test.go`.

### File List

- api/graphql/schema.graphqls
- internal/interface/graphql/errors.go
- internal/interface/graphql/resolver.go
- internal/interface/graphql/schema.resolvers.go
- internal/interface/graphql/integration_test.go
- internal/interface/graphql/generated.go
- internal/interface/graphql/model/models_gen.go
- gqlgen.yml
- cmd/server/main.go
- Makefile
- AGENTS.md
