package resilience

// Error codes for resilience operations.
// These codes are STABLE and should not be changed once published.
const (
	// ErrCodeCircuitOpen indicates that the circuit breaker is open and rejecting requests.
	ErrCodeCircuitOpen = "RES-001"

	// ErrCodeBulkheadFull indicates that the bulkhead capacity has been reached.
	ErrCodeBulkheadFull = "RES-002"

	// ErrCodeTimeoutExceeded indicates that an operation has exceeded its timeout.
	ErrCodeTimeoutExceeded = "RES-003"

	// ErrCodeMaxRetriesExceeded indicates that the maximum retry attempts have been exhausted.
	ErrCodeMaxRetriesExceeded = "RES-004"
)

// ResilienceError represents a resilience-related error with a stable code.
type ResilienceError struct {
	// Code is the stable error code for this error type.
	Code string
	// Message is a human-readable description of the error.
	Message string
	// Err is the underlying error that caused this error, if any.
	Err error
}

// Error returns the error message with code prefix.
func (e *ResilienceError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

// Unwrap returns the underlying error for error chain traversal.
func (e *ResilienceError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is matching by comparing error codes.
func (e *ResilienceError) Is(target error) bool {
	t, ok := target.(*ResilienceError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Sentinel errors for comparison using errors.Is.
var (
	// ErrCircuitOpen is returned when the circuit breaker is in open state.
	ErrCircuitOpen = &ResilienceError{
		Code:    ErrCodeCircuitOpen,
		Message: "circuit breaker is open",
	}

	// ErrBulkheadFull is returned when the bulkhead has reached its capacity.
	ErrBulkheadFull = &ResilienceError{
		Code:    ErrCodeBulkheadFull,
		Message: "bulkhead capacity exceeded",
	}

	// ErrTimeoutExceeded is returned when an operation times out.
	ErrTimeoutExceeded = &ResilienceError{
		Code:    ErrCodeTimeoutExceeded,
		Message: "timeout exceeded",
	}

	// ErrMaxRetriesExceeded is returned when all retry attempts have been exhausted.
	ErrMaxRetriesExceeded = &ResilienceError{
		Code:    ErrCodeMaxRetriesExceeded,
		Message: "maximum retry attempts exceeded",
	}
)

// NewCircuitOpenError creates a new circuit open error with an optional underlying error.
func NewCircuitOpenError(err error) error {
	return &ResilienceError{
		Code:    ErrCodeCircuitOpen,
		Message: "circuit breaker is open",
		Err:     err,
	}
}

// NewBulkheadFullError creates a new bulkhead full error with an optional underlying error.
func NewBulkheadFullError(err error) error {
	return &ResilienceError{
		Code:    ErrCodeBulkheadFull,
		Message: "bulkhead capacity exceeded",
		Err:     err,
	}
}

// NewTimeoutExceededError creates a new timeout exceeded error with an optional underlying error.
func NewTimeoutExceededError(err error) error {
	return &ResilienceError{
		Code:    ErrCodeTimeoutExceeded,
		Message: "timeout exceeded",
		Err:     err,
	}
}

// NewMaxRetriesExceededError creates a new max retries exceeded error with an optional underlying error.
func NewMaxRetriesExceededError(err error) error {
	return &ResilienceError{
		Code:    ErrCodeMaxRetriesExceeded,
		Message: "maximum retry attempts exceeded",
		Err:     err,
	}
}
