// Package observability provides observability utilities for the application.
package observability

import (
	"time"
)

// Logger defines logging abstraction for swappable implementations.
// Use the default ZapLogger or implement custom loggers (e.g., logrus, slog).
type Logger interface {
	// Debug logs a debug message with optional fields.
	Debug(msg string, fields ...Field)

	// Info logs an info message with optional fields.
	Info(msg string, fields ...Field)

	// Warn logs a warning message with optional fields.
	Warn(msg string, fields ...Field)

	// Error logs an error message with optional fields.
	Error(msg string, fields ...Field)

	// With returns a new Logger with the given fields added to all messages.
	With(fields ...Field) Logger

	// Sync flushes any buffered log entries.
	Sync() error
}

// Field represents a structured log field.
type Field struct {
	Key   string
	Value interface{}
}

// Field constructors for common types.

// String creates a string field.
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

// Int creates an int field.
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Int64 creates an int64 field.
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Duration creates a duration field.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

// Bool creates a boolean field.
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

// Err creates an error field with key "error".
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any creates a field with any value.
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}
