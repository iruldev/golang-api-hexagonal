# Story 7.10: Document Copy Pattern for New Modules

Status: done

## Story

As a developer,
I want documentation on copying example module,
So that I can create new domains quickly.

## Acceptance Criteria

### AC1: Copy pattern documented
**Given** README.md and AGENTS.md exist
**When** I follow the "Adding New Module" section
**Then** step-by-step guide is available
**And** checklist ensures all layers are created
**And** example `cp -r` commands are provided

---

## Tasks / Subtasks

- [x] **Task 1: Create step-by-step guide** (AC: #1)
  - [x] Step 1: Copy source files
  - [x] Step 2: Rename package and types
  - [x] Step 3: Create database migration
  - [x] Step 4: Generate SQLC and wire up

- [x] **Task 2: Add cp -r commands** (AC: #1)
  - [x] Copy domain, usecase, handler layers
  - [x] Include sed find/replace commands
  - [x] Include Linux/macOS notes

- [x] **Task 3: Create checklist** (AC: #1)
  - [x] Domain layer (4 items)
  - [x] Use case layer (2 items)
  - [x] Interface layer (4 items)
  - [x] Infrastructure layer (4 items)
  - [x] Wiring (3 items)

---

## Dev Notes

### Adding to README.md

Add section "Adding New Modules" with:
- Copy commands for each layer
- Find/replace instructions
- Checklist for verification

### Example Commands

```bash
# Copy note module to create "task" module
cp -r internal/domain/note internal/domain/task
cp -r internal/usecase/note internal/usecase/task
cp -r internal/interface/http/note internal/interface/http/task

# Find and replace
find internal/domain/task -type f -name "*.go" -exec sed -i '' 's/note/task/g' {} \;
```

### File List

Files to modify:
- `README.md` - Add "Adding New Modules" section
