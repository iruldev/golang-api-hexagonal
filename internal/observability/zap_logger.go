package observability

import (
	"time"

	"go.uber.org/zap"
)

// ZapLogger wraps zap.Logger to implement Logger interface.
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger from a zap.Logger.
func NewZapLogger(logger *zap.Logger) Logger {
	return &ZapLogger{logger: logger}
}

// Debug logs a debug message with optional fields.
func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, toZapFields(fields)...)
}

// Info logs an info message with optional fields.
func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, toZapFields(fields)...)
}

// Warn logs a warning message with optional fields.
func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, toZapFields(fields)...)
}

// Error logs an error message with optional fields.
func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, toZapFields(fields)...)
}

// With returns a new Logger with the given fields added to all messages.
func (z *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{logger: z.logger.With(toZapFields(fields)...)}
}

// Sync flushes any buffered log entries.
func (z *ZapLogger) Sync() error {
	return z.logger.Sync()
}

// Underlying returns the underlying zap.Logger.
// Useful when you need to access zap-specific functionality.
func (z *ZapLogger) Underlying() *zap.Logger {
	return z.logger
}

// toZapFields converts Field slice to zap.Field slice.
func toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, toZapField(f))
	}
	return zapFields
}

// toZapField converts a Field to a zap.Field.
func toZapField(f Field) zap.Field {
	switch v := f.Value.(type) {
	case string:
		return zap.String(f.Key, v)
	case int:
		return zap.Int(f.Key, v)
	case int64:
		return zap.Int64(f.Key, v)
	case float64:
		return zap.Float64(f.Key, v)
	case bool:
		return zap.Bool(f.Key, v)
	case time.Duration:
		return zap.Duration(f.Key, v)
	case error:
		return zap.Error(v)
	default:
		return zap.Any(f.Key, v)
	}
}

// NopLogger is a no-op logger for testing.
type NopLogger struct{}

// NewNopLoggerInterface creates a new NopLogger that implements Logger.
func NewNopLoggerInterface() Logger {
	return &NopLogger{}
}

func (n *NopLogger) Debug(_ string, _ ...Field) {}
func (n *NopLogger) Info(_ string, _ ...Field)  {}
func (n *NopLogger) Warn(_ string, _ ...Field)  {}
func (n *NopLogger) Error(_ string, _ ...Field) {}
func (n *NopLogger) With(_ ...Field) Logger     { return n }
func (n *NopLogger) Sync() error                { return nil }
