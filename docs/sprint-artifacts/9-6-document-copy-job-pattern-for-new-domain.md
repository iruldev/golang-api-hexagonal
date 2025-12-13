# Story 9.6: Document Copy Job Pattern for New Domain

Status: done

## Story

As a developer,
I want documentation on adding new job types,
So that I can extend async capabilities consistently.

## Acceptance Criteria

### AC1: AGENTS.md Async Section
**Given** AGENTS.md is updated
**When** I read the async patterns section
**Then** step-by-step guide exists for new job types
**And** guide covers all 4 job patterns (Fire-and-Forget, Scheduled, Fanout, Idempotent)

### AC2: Task Creation Checklist
**Given** I want to create a new async job
**When** I follow the checklist in AGENTS.md
**Then** checklist ensures all components are created:
- Task type constant in `types.go`
- Payload struct in task file
- Task constructor function
- Handler struct with dependencies
- Handle method with validation
- Registration in worker main
- Unit tests

### AC3: Copy Commands Provided
**Given** documentation exists
**When** I want to copy the example job
**Then** example copy commands are provided for:
- Task type definition
- Task handler file
- Test file
**And** placeholder replacement instructions are clear

---

## Tasks / Subtasks

- [x] **Task 1: Add Async Job Patterns section to AGENTS.md** (AC: #1)
  - [x] Add "Adding a New Async Job" section after "Adding a New Domain"
  - [x] Document step-by-step process for creating new job types
  - [x] Reference existing patterns in `internal/worker/patterns/`
  - [x] Reference `docs/async-jobs.md` for detailed documentation

- [x] **Task 2: Create job type creation checklist** (AC: #2)
  - [x] Define checklist items matching existing patterns
  - [x] Include file locations for each component
  - [x] Reference `note_archive.go` as canonical example
  - [x] Include test checklist items

- [x] **Task 3: Add copy commands and templates** (AC: #3)
  - [x] Provide shell commands to copy task files
  - [x] Provide sed/replace commands for placeholders
  - [x] Document queue selection guidance
  - [x] Add pattern selection guidance (when to use which pattern)

- [x] **Task 4: Add job pattern selection guide** (AC: #1)
  - [x] Create decision table for pattern selection
  - [x] Fire-and-Forget: non-critical, best-effort
  - [x] Scheduled: cron-based periodic tasks
  - [x] Fanout: event-driven, multiple handlers
  - [x] Idempotent: prevent duplicate processing

- [x] **Task 5: Verify documentation accuracy** (AC: #1, #2, #3)
  - [x] Cross-reference with existing `docs/async-jobs.md`
  - [x] Ensure file paths are accurate
  - [x] Test copy commands work correctly
  - [x] Run `make lint` + `make test` to verify no regressions

---

## Dev Notes

### Current AGENTS.md Structure

The existing AGENTS.md has:
- ✅ DO / DON'T sections
- 📁 File Structure Conventions
- 🧪 Testing Requirements
- 🔧 Common Tasks → "Adding a New Domain" (step 1-11)
- 📋 Checklist for Code Review

**Add new section after "Adding a New Domain":**
```
## 🔧 Common Tasks

### Adding a New Domain (existing)
...

### Adding a New Async Job (NEW)
...
```

---

### Reference Files

| Component | Example File |
|-----------|-------------|
| Task types | `internal/worker/tasks/types.go` |
| Task handler | `internal/worker/tasks/note_archive.go` |
| Task tests | `internal/worker/tasks/note_archive_test.go` |
| TaskEnqueuer | `internal/worker/tasks/enqueue.go` |
| Fire-and-Forget | `internal/worker/patterns/fireandforget.go` |
| Scheduled | `internal/worker/patterns/scheduled.go` |
| Fanout | `internal/worker/patterns/fanout.go` |
| Idempotency | `internal/worker/idempotency/` |
| Worker main | `cmd/worker/main.go` |
| Scheduler main | `cmd/scheduler/main.go` |
| Full docs | `docs/async-jobs.md` |

---

### Task Type Naming Convention

```go
// internal/worker/tasks/types.go
const (
    TypeNoteArchive = "note:archive"
    // Convention: {domain}:{action}
    // Examples:
    TypeEmailSend      = "email:send"
    TypeReportGenerate = "report:generate"
    TypeUserCleanup    = "user:cleanup"
)
```

---

### Handler Pattern (From note_archive.go)

```go
// 1. Payload struct
type NoteArchivePayload struct {
    NoteID uuid.UUID `json:"note_id"`
}

// 2. Task constructor
func NewNoteArchiveTask(noteID uuid.UUID) (*asynq.Task, error) {
    payload, err := json.Marshal(NoteArchivePayload{NoteID: noteID})
    if err != nil {
        return nil, fmt.Errorf("marshal payload: %w", err)
    }
    return asynq.NewTask(TypeNoteArchive, payload, asynq.MaxRetry(3)), nil
}

// 3. Handler struct
type NoteArchiveHandler struct {
    logger *zap.Logger
    // Add dependencies: repo, usecase, etc.
}

func NewNoteArchiveHandler(logger *zap.Logger) *NoteArchiveHandler {
    return &NoteArchiveHandler{logger: logger}
}

// 4. Handle method with validation
func (h *NoteArchiveHandler) Handle(ctx context.Context, t *asynq.Task) error {
    taskID, _ := asynq.GetTaskID(ctx)
    
    // Unmarshal
    var p NoteArchivePayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("unmarshal: %v: %w", err, asynq.SkipRetry)
    }
    
    // Validate
    if p.NoteID == uuid.Nil {
        return fmt.Errorf("note_id required: %w", asynq.SkipRetry)
    }
    
    // Process
    h.logger.Info("processing", zap.String("task_id", taskID))
    
    return nil
}
```

---

### Copy Commands Template

```bash
# Create new task file
cp internal/worker/tasks/note_archive.go internal/worker/tasks/{name}.go
cp internal/worker/tasks/note_archive_test.go internal/worker/tasks/{name}_test.go

# Replace placeholders
sed -i '' 's/NoteArchive/{Name}/g' internal/worker/tasks/{name}.go
sed -i '' 's/note:archive/{domain}:{action}/g' internal/worker/tasks/{name}.go
sed -i '' 's/NoteID/YourFieldID/g' internal/worker/tasks/{name}.go

# Add type constant to types.go
echo 'Type{Name} = "{domain}:{action}"' >> internal/worker/tasks/types.go

# Register in cmd/worker/main.go
# Add: srv.HandleFunc(tasks.Type{Name}, {name}Handler.Handle)
```

---

### Pattern Selection Decision Table

| Scenario | Pattern | Queue |
|----------|---------|-------|
| Non-critical background (analytics, audit) | Fire-and-Forget | `low` |
| Scheduled cleanup/reports | Scheduled | `default` |
| Event → multiple handlers | Fanout | per-handler |
| Critical operations (payments, orders) | Standard + Idempotency | `critical` |
| User-triggered async (email, notification) | Standard | `default`/`critical` |

---

### Previous Story Patterns

**From Story 9.1 (Fire-and-Forget):**
- Pattern function with default queue override
- Error handling logs but doesn't propagate

**From Story 9.2 (Scheduled):**
- Separate scheduler binary (`cmd/scheduler/main.go`)
- Cron validation helper
- RegisterScheduledJobs pattern

**From Story 9.3 (Fanout):**
- FanoutRegistry for handler registration
- FanoutDispatcher for worker processing
- Per-handler queue assignment

**From Story 9.4 (Idempotency):**
- IdempotentHandler wrapper
- RedisStore with FailOpen/FailClosed modes
- Key extractor pattern

---

### References

- [Source: docs/epics.md#Story-9.6] - Story requirements
- [Source: AGENTS.md] - Current documentation structure
- [Source: docs/async-jobs.md] - Comprehensive async job documentation
- [Source: internal/worker/tasks/note_archive.go] - Reference implementation
- [Source: docs/sprint-artifacts/9-1-implement-fire-and-forget-job-pattern.md] - Pattern learnings
- [Source: docs/sprint-artifacts/9-2-implement-scheduled-job-pattern-with-cron.md] - Scheduler learnings
- [Source: docs/sprint-artifacts/9-3-implement-fanout-job-pattern.md] - Fanout learnings
- [Source: docs/sprint-artifacts/9-4-add-idempotency-key-pattern.md] - Idempotency learnings

---

## Dev Agent Record

### Context Reference

Previous stories:
- `docs/sprint-artifacts/9-5-create-dedicated-job-metrics-dashboard.md`
- `docs/sprint-artifacts/9-4-add-idempotency-key-pattern.md`
- `docs/sprint-artifacts/8-8-document-async-job-patterns.md`

Target file: `AGENTS.md`
Reference docs: `docs/async-jobs.md`

### Agent Model Used

Claude Sonnet 4 (2025-12-13)

### Debug Log References

### Completion Notes List

- ✅ Added "Adding a New Async Job" section to `AGENTS.md` (137 new lines)
- ✅ Section includes 5-step guide for creating new async jobs
- ✅ Job creation checklist with 8 items covering all components
- ✅ Copy commands for macOS with sed placeholder replacement
- ✅ Queue selection guide with priority weights
- ✅ Pattern selection decision table covering all 4 patterns
- ✅ Reference to `docs/async-jobs.md` for comprehensive documentation
- ✅ All tests pass (coverage 93.2%-94.1%)
- ⚠️ Pre-existing lint errors in Go files (errcheck, buildtag) - not related to this documentation story
- ✅ Code review passed with minor fixes applied

### File List

**Modified:**
- `AGENTS.md` - Added "Adding a New Async Job" section with step-by-step guide, checklist, copy commands, and pattern selection decision table; added Linux sed compatibility note
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status

### Senior Developer Review (AI)

**Review Date:** 2025-12-13
**Reviewer:** Claude Sonnet 4 (Code Review Workflow)

**Outcome:** ✅ APPROVED

**Findings Addressed:**
| Severity | Issue | Resolution |
|----------|-------|------------|
| MEDIUM | Pre-existing lint errors | Noted as out-of-scope for this story |
| LOW | Story file untracked in git | Added to git tracking |
| LOW | Copy commands Linux incompatible | Added Linux sed syntax note |
| LOW | Missing AC references | Minor, not blocking |

**Verification:**
- ✅ All 3 ACs implemented and verified
- ✅ All 5 tasks marked complete with evidence
- ✅ Tests pass (94.1% coverage)
- ✅ Cross-reference with `docs/async-jobs.md` valid
- ✅ File paths accurate

### Change Log

| Date | Changes |
|------|---------|
| 2025-12-13 | Story drafted with comprehensive developer context |
| 2025-12-13 | **Implementation Complete:** All 5 tasks implemented, AGENTS.md updated with async job documentation |
| 2025-12-13 | **Code Review Passed:** 1 Medium, 3 Low issues found; fixes applied automatically |

