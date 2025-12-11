# Story 6.1: Define Logger Interface

Status: done

## Story

As a developer,
I want a Logger interface abstraction,
So that I can swap logging implementations.

## Acceptance Criteria

### AC1: Logger interface defined
**Given** `internal/observability/logger.go` exists
**When** I implement Logger interface
**Then** methods include: Debug, Info, Warn, Error, With(fields)
**And** default implementation uses zap

---

## Tasks / Subtasks

- [x] **Task 1: Define Logger interface** (AC: #1)
  - [x] Create Logger interface with Debug, Info, Warn, Error methods
  - [x] Add With(fields) for structured logging
  - [x] Add common field types (String, Int, Duration, Error)

- [x] **Task 2: Implement ZapLogger wrapper** (AC: #1)
  - [x] Wrap *zap.Logger to implement Logger interface
  - [x] Map interface methods to zap methods

- [x] **Task 3: Update existing code** (AC: #1)
  - [x] Added NopLogger for testing that implements Logger
  - [x] Backward compatible - original logger.go unchanged

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Logger Interface

```go
// internal/observability/logger_interface.go
package observability

// Logger defines logging abstraction for swappable implementations.
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    With(fields ...Field) Logger
}

// Field represents a structured log field.
type Field struct {
    Key   string
    Value interface{}
}

// Field constructors
func String(key, val string) Field { return Field{Key: key, Value: val} }
func Int(key string, val int) Field { return Field{Key: key, Value: val} }
func Duration(key string, val time.Duration) Field { return Field{Key: key, Value: val} }
func Err(err error) Field { return Field{Key: "error", Value: err} }
```

### ZapLogger Wrapper

```go
// ZapLogger wraps zap.Logger to implement Logger interface.
type ZapLogger struct {
    logger *zap.Logger
}

func NewZapLogger(logger *zap.Logger) Logger {
    return &ZapLogger{logger: logger}
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
    z.logger.Info(msg, toZapFields(fields)...)
}
```

### Architecture Compliance

**Layer:** `internal/observability/`
**Pattern:** Interface abstraction + default implementation
**Benefit:** Swappable logging (e.g., zap → logrus → slog)

### References

- [Source: docs/epics.md#Story-6.1]
- [Story 3.3 - Logging Middleware](file:///docs/sprint-artifacts/3-3-implement-logging-middleware.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
First story in Epic 6: Extension Interfaces.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/observability/logger_interface.go` - Logger interface and Field types
- `internal/observability/zap_logger.go` - ZapLogger wrapper

Files to modify:
- `internal/observability/logger.go` - Update NewLogger return type
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
