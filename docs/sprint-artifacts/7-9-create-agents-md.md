# Story 7.9: Create AGENTS.md

Status: done

## Story

As a AI assistant user,
I want AGENTS.md as AI contract,
So that AI assistants follow consistent patterns.

## Acceptance Criteria

### AC1: AGENTS.md exists
**Given** `AGENTS.md` exists in project root
**When** AI reads the document
**Then** DO/DON'T patterns are clear
**And** file structure conventions are listed
**And** testing requirements are documented

---

## Tasks / Subtasks

- [x] **Task 1: Document DO patterns** (AC: #1)
  - [x] Architecture (4 patterns)
  - [x] Code Style (5 patterns)
  - [x] Patterns (5 patterns)
  - [x] Testing (4 patterns)

- [x] **Task 2: Document DON'T patterns** (AC: #1)
  - [x] Architecture (4 patterns)
  - [x] Code Style (4 patterns)
  - [x] Patterns (4 patterns)
  - [x] Testing (4 patterns)

- [x] **Task 3: Document file structure conventions** (AC: #1)
  - [x] Per domain structure (directory tree)
  - [x] Naming conventions (table)
  - [x] Database conventions

- [x] **Task 4: Document testing requirements** (AC: #1)
  - [x] Unit test pattern (code example)
  - [x] Coverage requirements (table)
  - [x] Mock pattern
  - [x] Integration test pattern

---

## Dev Notes

### AGENTS.md Structure

```markdown
# AI Assistant Guide (AGENTS.md)

## DO
- Follow hexagonal architecture
- Use existing patterns
- Write tests

## DON'T
- Skip validation
- Bypass layers
- Create new patterns

## File Structure
- domain/ entity, errors, repository interface
- usecase/ business logic
- interface/ HTTP handlers, DTOs
- infra/ database implementations

## Testing Requirements
- Unit tests for all layers
- Table-driven tests
- AAA pattern
- Mock dependencies
```

### File List

Files to create:
- `AGENTS.md` - AI assistant guide and contract
