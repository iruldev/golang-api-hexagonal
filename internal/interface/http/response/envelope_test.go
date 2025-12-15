// Package response tests for HTTP response helpers.
package response

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// TestEnvelopeStructure tests that Envelope serializes correctly.
func TestEnvelopeStructure(t *testing.T) {
	tests := []struct {
		name     string
		envelope Envelope
		want     map[string]any
	}{
		{
			name: "success response with data and meta",
			envelope: Envelope{
				Data: map[string]string{"name": "test"},
				Meta: &Meta{TraceID: "trace-123"},
			},
			want: map[string]any{
				"data": map[string]any{"name": "test"},
				"meta": map[string]any{"trace_id": "trace-123"},
			},
		},
		{
			name: "error response with code and message",
			envelope: Envelope{
				Error: &ErrorBody{Code: "NOT_FOUND", Message: "Resource not found"},
				Meta:  &Meta{TraceID: "trace-456"},
			},
			want: map[string]any{
				"error": map[string]any{"code": "NOT_FOUND", "message": "Resource not found"},
				"meta":  map[string]any{"trace_id": "trace-456"},
			},
		},
		{
			name: "error response with hint",
			envelope: Envelope{
				Error: &ErrorBody{Code: "BAD_REQUEST", Message: "Invalid input", Hint: "Check the field format"},
				Meta:  &Meta{TraceID: "trace-789"},
			},
			want: map[string]any{
				"error": map[string]any{"code": "BAD_REQUEST", "message": "Invalid input", "hint": "Check the field format"},
				"meta":  map[string]any{"trace_id": "trace-789"},
			},
		},
		{
			name: "pagination response",
			envelope: Envelope{
				Data: []string{"item1", "item2"},
				Meta: &Meta{TraceID: "trace-pag", Page: 1, PageSize: 10, Total: 100},
			},
			want: map[string]any{
				"data": []any{"item1", "item2"},
				"meta": map[string]any{"trace_id": "trace-pag", "page": float64(1), "page_size": float64(10), "total": float64(100)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.envelope)
			if err != nil {
				t.Fatalf("failed to marshal envelope: %v", err)
			}

			var got map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to unmarshal envelope: %v", err)
			}

			// Check data field
			if tt.want["data"] != nil {
				if got["data"] == nil {
					t.Error("expected data field to be present")
				}
			}

			// Check error field
			if tt.want["error"] != nil {
				gotErr, ok := got["error"].(map[string]any)
				if !ok {
					t.Error("expected error field to be present")
				}
				wantErr := tt.want["error"].(map[string]any)
				if gotErr["code"] != wantErr["code"] {
					t.Errorf("error.code = %v, want %v", gotErr["code"], wantErr["code"])
				}
				if gotErr["message"] != wantErr["message"] {
					t.Errorf("error.message = %v, want %v", gotErr["message"], wantErr["message"])
				}
			}

			// Check meta field
			gotMeta, ok := got["meta"].(map[string]any)
			if !ok {
				t.Error("expected meta field to be present")
			}
			wantMeta := tt.want["meta"].(map[string]any)
			if gotMeta["trace_id"] != wantMeta["trace_id"] {
				t.Errorf("meta.trace_id = %v, want %v", gotMeta["trace_id"], wantMeta["trace_id"])
			}
		})
	}
}

// TestTraceIDExtraction tests that trace_id is correctly extracted from context.
func TestTraceIDExtraction(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		wantTraceID string
	}{
		{
			name:        "with trace_id in context",
			ctx:         ctxutil.NewRequestIDContext(context.Background(), "test-trace-id"),
			wantTraceID: "test-trace-id",
		},
		{
			name:        "without trace_id in context",
			ctx:         context.Background(),
			wantTraceID: "unknown",
		},
		{
			name:        "with empty trace_id",
			ctx:         ctxutil.NewRequestIDContext(context.Background(), ""),
			wantTraceID: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTraceID(tt.ctx)
			if got != tt.wantTraceID {
				t.Errorf("getTraceID() = %v, want %v", got, tt.wantTraceID)
			}
		})
	}
}

// TestSuccessEnvelope tests the SuccessEnvelope function.
func TestSuccessEnvelope(t *testing.T) {
	ctx := ctxutil.NewRequestIDContext(context.Background(), "success-trace-id")
	w := httptest.NewRecorder()

	SuccessEnvelope(w, ctx, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", ct)
	}

	var envelope Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if envelope.Data == nil {
		t.Error("expected data to be present")
	}
	if envelope.Error != nil {
		t.Error("expected error to be nil")
	}
	if envelope.Meta == nil {
		t.Error("expected meta to be present")
	}
	if envelope.Meta.TraceID != "success-trace-id" {
		t.Errorf("meta.trace_id = %v, want success-trace-id", envelope.Meta.TraceID)
	}
}

// TestSuccessEnvelopeWithPagination tests the SuccessEnvelopeWithPagination function.
func TestSuccessEnvelopeWithPagination(t *testing.T) {
	ctx := ctxutil.NewRequestIDContext(context.Background(), "pagination-trace-id")
	w := httptest.NewRecorder()

	SuccessEnvelopeWithPagination(w, ctx, []string{"a", "b", "c"}, Pagination{
		Page:     2,
		PageSize: 10,
		Total:    25,
	})

	if w.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v", w.Code, http.StatusOK)
	}

	var envelope Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if envelope.Meta == nil {
		t.Fatal("expected meta to be present")
	}
	if envelope.Meta.TraceID != "pagination-trace-id" {
		t.Errorf("meta.trace_id = %v, want pagination-trace-id", envelope.Meta.TraceID)
	}
	if envelope.Meta.Page != 2 {
		t.Errorf("meta.page = %v, want 2", envelope.Meta.Page)
	}
	if envelope.Meta.PageSize != 10 {
		t.Errorf("meta.page_size = %v, want 10", envelope.Meta.PageSize)
	}
	if envelope.Meta.Total != 25 {
		t.Errorf("meta.total = %v, want 25", envelope.Meta.Total)
	}
}

// TestErrorEnvelope tests the ErrorEnvelope function.
func TestErrorEnvelope(t *testing.T) {
	ctx := ctxutil.NewRequestIDContext(context.Background(), "error-trace-id")
	w := httptest.NewRecorder()

	ErrorEnvelope(w, ctx, http.StatusNotFound, "NOT_FOUND", "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status code = %v, want %v", w.Code, http.StatusNotFound)
	}

	var envelope Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if envelope.Data != nil {
		t.Error("expected data to be nil")
	}
	if envelope.Error == nil {
		t.Error("expected error to be present")
	}
	if envelope.Error.Code != "NOT_FOUND" {
		t.Errorf("error.code = %v, want NOT_FOUND", envelope.Error.Code)
	}
	if envelope.Error.Message != "Resource not found" {
		t.Errorf("error.message = %v, want Resource not found", envelope.Error.Message)
	}
	if envelope.Meta == nil {
		t.Error("expected meta to be present")
	}
	if envelope.Meta.TraceID != "error-trace-id" {
		t.Errorf("meta.trace_id = %v, want error-trace-id", envelope.Meta.TraceID)
	}
}

// TestErrorCodeFormat tests that error codes use UPPER_SNAKE format.
func TestErrorCodeFormat(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		wantFormat   bool
		isDeprecated bool
	}{
		{"CodeBadRequest", CodeBadRequest, true, false},
		{"CodeUnauthorized", CodeUnauthorized, true, false},
		{"CodeForbidden", CodeForbidden, true, false},
		{"CodeNotFound", CodeNotFound, true, false},
		{"CodeConflict", CodeConflict, true, false},
		{"CodeValidation", CodeValidation, true, false},
		{"CodeInternalServer", domainerrors.CodeInternalError, true, false},
		{"CodeTimeout", CodeTimeout, true, false},
		{"CodeServiceUnavailable", CodeServiceUnavailable, true, false},
		// Deprecated codes should still work
		{"ErrBadRequest (deprecated)", ErrBadRequest, false, true},
		{"ErrNotFound (deprecated)", ErrNotFound, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantFormat {
				// Check for UPPER_SNAKE format without ERR_ prefix
				if len(tt.code) > 4 && tt.code[:4] == "ERR_" {
					t.Errorf("code %v should not have ERR_ prefix", tt.code)
				}
			}
			if tt.isDeprecated {
				// Deprecated codes should have ERR_ prefix
				if len(tt.code) < 4 || tt.code[:4] != "ERR_" {
					t.Errorf("deprecated code %v should have ERR_ prefix", tt.code)
				}
			}
		})
	}
}

// TestMetaAlwaysPresent verifies that meta field is always present in responses.
func TestMetaAlwaysPresent(t *testing.T) {
	tests := []struct {
		name    string
		handler func(http.ResponseWriter, context.Context)
	}{
		{
			name: "SuccessEnvelope",
			handler: func(w http.ResponseWriter, ctx context.Context) {
				SuccessEnvelope(w, ctx, nil)
			},
		},
		{
			name: "SuccessEnvelopeWithStatus",
			handler: func(w http.ResponseWriter, ctx context.Context) {
				SuccessEnvelopeWithStatus(w, http.StatusCreated, ctx, nil)
			},
		},
		{
			name: "ErrorEnvelope",
			handler: func(w http.ResponseWriter, ctx context.Context) {
				ErrorEnvelope(w, ctx, http.StatusBadRequest, "BAD_REQUEST", "error")
			},
		},
		{
			name: "BadRequestCtx",
			handler: func(w http.ResponseWriter, ctx context.Context) {
				BadRequestCtx(w, ctx, "bad request")
			},
		},
		{
			name: "NotFoundCtx",
			handler: func(w http.ResponseWriter, ctx context.Context) {
				NotFoundCtx(w, ctx, "not found")
			},
		},
		{
			name: "UnauthorizedCtx",
			handler: func(w http.ResponseWriter, ctx context.Context) {
				UnauthorizedCtx(w, ctx, "unauthorized")
			},
		},
		{
			name: "InternalServerErrorCtx",
			handler: func(w http.ResponseWriter, ctx context.Context) {
				InternalServerErrorCtx(w, ctx, "internal error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ctxutil.NewRequestIDContext(context.Background(), "test-trace")
			w := httptest.NewRecorder()

			tt.handler(w, ctx)

			var envelope Envelope
			if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if envelope.Meta == nil {
				t.Error("expected meta field to be present")
			}
			if envelope.Meta.TraceID == "" {
				t.Error("expected meta.trace_id to be non-empty")
			}
		})
	}
}

// TestConvenienceFunctions tests all ctx convenience functions.
func TestConvenienceFunctions(t *testing.T) {
	ctx := ctxutil.NewRequestIDContext(context.Background(), "conv-trace")

	tests := []struct {
		name       string
		handler    func(http.ResponseWriter, context.Context)
		wantStatus int
		wantCode   string
	}{
		{"BadRequestCtx", func(w http.ResponseWriter, ctx context.Context) { BadRequestCtx(w, ctx, "msg") }, http.StatusBadRequest, CodeBadRequest},
		{"UnauthorizedCtx", func(w http.ResponseWriter, ctx context.Context) { UnauthorizedCtx(w, ctx, "msg") }, http.StatusUnauthorized, CodeUnauthorized},
		{"ForbiddenCtx", func(w http.ResponseWriter, ctx context.Context) { ForbiddenCtx(w, ctx, "msg") }, http.StatusForbidden, CodeForbidden},
		{"NotFoundCtx", func(w http.ResponseWriter, ctx context.Context) { NotFoundCtx(w, ctx, "msg") }, http.StatusNotFound, CodeNotFound},
		{"ConflictCtx", func(w http.ResponseWriter, ctx context.Context) { ConflictCtx(w, ctx, "msg") }, http.StatusConflict, CodeConflict},
		{"ValidationErrorCtx", func(w http.ResponseWriter, ctx context.Context) { ValidationErrorCtx(w, ctx, "msg") }, http.StatusUnprocessableEntity, CodeValidation},
		{"InternalServerErrorCtx", func(w http.ResponseWriter, ctx context.Context) { InternalServerErrorCtx(w, ctx, "msg") }, http.StatusInternalServerError, domainerrors.CodeInternalError},
		{"ServiceUnavailableCtx", func(w http.ResponseWriter, ctx context.Context) { ServiceUnavailableCtx(w, ctx, "msg") }, http.StatusServiceUnavailable, CodeServiceUnavailable},
		{"TimeoutCtx", func(w http.ResponseWriter, ctx context.Context) { TimeoutCtx(w, ctx, "msg") }, http.StatusGatewayTimeout, CodeTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.handler(w, ctx)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", w.Code, tt.wantStatus)
			}

			var envelope Envelope
			if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if envelope.Error == nil {
				t.Fatal("expected error to be present")
			}
			if envelope.Error.Code != tt.wantCode {
				t.Errorf("error.code = %v, want %v", envelope.Error.Code, tt.wantCode)
			}
			if envelope.Meta == nil || envelope.Meta.TraceID != "conv-trace" {
				t.Error("expected trace_id to be present")
			}
		})
	}
}
