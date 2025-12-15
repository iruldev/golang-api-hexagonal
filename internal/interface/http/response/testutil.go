package response

import "github.com/iruldev/golang-api-hexagonal/internal/observability"

// TestEnvelopeResponse is the Envelope structure for test assertions.
// Exported for use in middleware tests and other packages that need to
// verify Envelope format in responses.
type TestEnvelopeResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *TestError  `json:"error,omitempty"`
	Meta  *TestMeta   `json:"meta,omitempty"`
}

// TestError represents the error field in an Envelope for test assertions.
type TestError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// TestMeta represents the meta field in an Envelope for test assertions.
type TestMeta struct {
	TraceID string `json:"trace_id"`
}

// MockLogger is a mock implementation of observability.Logger for testing.
// It records the last error message for assertion.
type MockLogger struct {
	ErrorCalled bool
	ErrorMsg    string
	Fields      []observability.Field
}

func (m *MockLogger) Debug(msg string, fields ...observability.Field) {}
func (m *MockLogger) Info(msg string, fields ...observability.Field)  {}
func (m *MockLogger) Warn(msg string, fields ...observability.Field)  {}
func (m *MockLogger) Error(msg string, fields ...observability.Field) {
	m.ErrorCalled = true
	m.ErrorMsg = msg
	m.Fields = fields
}
func (m *MockLogger) With(fields ...observability.Field) observability.Logger { return m }
func (m *MockLogger) Sync() error                                             { return nil }
