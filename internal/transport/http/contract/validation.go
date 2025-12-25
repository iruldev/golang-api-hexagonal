package contract

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// DecodeAndValidateJSON decodes a JSON body into dst and runs validation tags.
// Uses strict decoding that rejects unknown fields.
func DecodeAndValidateJSON[T any](r io.Reader, dst *T) ([]ValidationError, error) {
	if err := DecodeJSONStrict(r, dst); err != nil {
		// Check if it's a JSONDecodeError for better error messages
		if jsonErr, ok := err.(*JSONDecodeError); ok {
			switch jsonErr.Kind {
			case JSONDecodeErrorKindUnknownField:
				return []ValidationError{{
					Field:   jsonErr.Field,
					Message: "unknown field",
				}}, err
			case JSONDecodeErrorKindTypeMismatch:
				return []ValidationError{{
					Field:   jsonErr.Field,
					Message: "invalid type",
				}}, err
			case JSONDecodeErrorKindTrailingData:
				return []ValidationError{{
					Field:   "body",
					Message: "request body contains trailing data",
				}}, err
			}
		}
		return []ValidationError{{
			Field:   "body",
			Message: "invalid request body",
		}}, err
	}
	return Validate(*dst), nil
}

// ValidateRequestBody is a convenience for HTTP handlers to decode and validate payloads.
func ValidateRequestBody[T any](r *http.Request, dst *T) []ValidationError {
	errs, err := DecodeAndValidateJSON(r.Body, dst)
	if err != nil {
		return errs
	}
	return errs
}

// Validate validates a struct using go-playground/validator rules.
// Returns a slice of ValidationError with camelCase field names.
func Validate(v any) []ValidationError {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		result := make([]ValidationError, len(validationErrors))
		for i, fe := range validationErrors {
			fieldName := fe.Field()
			if fieldName == "" {
				fieldName = fe.StructField()
			}
			result[i] = ValidationError{
				Field:   toLowerCamelCase(fieldName),
				Message: validationMessage(fe),
			}
		}
		return result
	}

	return []ValidationError{{
		Field:   "body",
		Message: "invalid request body",
	}}
}

// validationMessage returns a human-readable message for the field error.
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	default:
		return "is invalid"
	}
}

// toLowerCamelCase converts PascalCase to camelCase.
func toLowerCamelCase(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
