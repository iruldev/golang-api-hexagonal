package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// JSONDecodeError represents an error during JSON decoding with additional context.
type JSONDecodeError struct {
	// Kind indicates the type of error (unknown_field, syntax, type_mismatch, etc.)
	Kind string
	// Field is the field name that caused the error (if applicable)
	Field string
	// Message is a human-readable description of the error
	Message string
	// Err is the underlying error
	Err error
}

func (e *JSONDecodeError) Error() string {
	return e.Message
}

func (e *JSONDecodeError) Unwrap() error {
	return e.Err
}

const (
	// JSONDecodeErrorKindUnknownField indicates an unknown field was present in the JSON.
	JSONDecodeErrorKindUnknownField = "unknown_field"
	// JSONDecodeErrorKindSyntax indicates a JSON syntax error.
	JSONDecodeErrorKindSyntax = "syntax"
	// JSONDecodeErrorKindTypeMismatch indicates a type mismatch during decoding.
	JSONDecodeErrorKindTypeMismatch = "type_mismatch"
	// JSONDecodeErrorKindEOF indicates unexpected end of input.
	JSONDecodeErrorKindEOF = "eof"
	// JSONDecodeErrorKindTrailingData indicates trailing data after JSON object.
	JSONDecodeErrorKindTrailingData = "trailing_data"
	// JSONDecodeErrorKindOther indicates an unclassified error.
	JSONDecodeErrorKindOther = "other"
)

// DecodeJSONStrict decodes JSON from r into dst with strict mode enabled.
// Unknown fields in the JSON will cause an error.
// Returns a *JSONDecodeError with classified error information.
func DecodeJSONStrict(r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return classifyJSONError(err)
	}

	// Check for trailing data after JSON object
	if dec.More() {
		return &JSONDecodeError{
			Kind:    JSONDecodeErrorKindTrailingData,
			Message: "trailing data after JSON object",
			Err:     errors.New("trailing data after JSON object"),
		}
	}

	return nil
}

// classifyJSONError converts a json decoding error into a JSONDecodeError with context.
func classifyJSONError(err error) *JSONDecodeError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for unknown field error (e.g., "json: unknown field \"usernmae\"")
	if strings.HasPrefix(errMsg, "json: unknown field") {
		field := extractFieldFromUnknownFieldError(errMsg)
		return &JSONDecodeError{
			Kind:    JSONDecodeErrorKindUnknownField,
			Field:   field,
			Message: fmt.Sprintf("unknown field: %s", field),
			Err:     err,
		}
	}

	// Check for syntax error
	var syntaxErr *json.SyntaxError
	if ok := isType(err, &syntaxErr); ok {
		return &JSONDecodeError{
			Kind:    JSONDecodeErrorKindSyntax,
			Message: "invalid JSON syntax",
			Err:     err,
		}
	}

	// Check for type mismatch
	var unmarshalErr *json.UnmarshalTypeError
	if ok := isType(err, &unmarshalErr); ok {
		return &JSONDecodeError{
			Kind:    JSONDecodeErrorKindTypeMismatch,
			Field:   unmarshalErr.Field,
			Message: fmt.Sprintf("invalid type for field %s: expected %s", unmarshalErr.Field, unmarshalErr.Type.String()),
			Err:     err,
		}
	}

	// Check for EOF
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return &JSONDecodeError{
			Kind:    JSONDecodeErrorKindEOF,
			Message: "unexpected end of JSON input",
			Err:     err,
		}
	}

	// Default case
	return &JSONDecodeError{
		Kind:    JSONDecodeErrorKindOther,
		Message: "invalid request body",
		Err:     err,
	}
}

// extractFieldFromUnknownFieldError extracts the field name from the error message.
// Example input: "json: unknown field \"usernmae\""
func extractFieldFromUnknownFieldError(errMsg string) string {
	// Find the quote-delimited field name
	start := strings.Index(errMsg, "\"")
	if start == -1 {
		return "unknown"
	}
	end := strings.LastIndex(errMsg, "\"")
	if end <= start {
		return "unknown"
	}
	return errMsg[start+1 : end]
}

// isType is a helper to check if err is of a specific type.
func isType[T error](err error, target *T) bool {
	var t T
	if as, ok := err.(T); ok {
		*target = as
		return true
	}
	_ = t // silence unused variable warning
	return false
}
