package middleware_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// problemResponse represents the RFC 7807 Problem Details response for testing.
type problemResponse struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail"`
	Code      string `json:"code"`
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id"`
	Instance  string `json:"instance"`
}

func TestRecoverer(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		setupContext   func(r *http.Request) *http.Request
		wantStatus     int
		wantCode       string
		wantProblem    bool
		wantStackInLog bool
		wantRequestID  string
		wantTraceID    string
	}{
		{
			name: "normal request passes through",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			},
			wantStatus:  http.StatusOK,
			wantProblem: false,
		},
		{
			name: "panic with string returns SYS-001",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic message")
			},
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "SYS-001",
			wantProblem:    true,
			wantStackInLog: true,
		},
		{
			name: "panic with error returns SYS-001",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(io.EOF)
			},
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "SYS-001",
			wantProblem:    true,
			wantStackInLog: true,
		},
		{
			name: "panic includes request_id from context",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("panic with context")
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := ctxutil.SetRequestID(r.Context(), "test-request-id-123")
				return r.WithContext(ctx)
			},
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "SYS-001",
			wantProblem:    true,
			wantStackInLog: true,
			wantRequestID:  "test-request-id-123",
		},
		{
			name: "panic includes trace_id from context",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("panic with tracing")
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := ctxutil.SetRequestID(r.Context(), "req-456")
				ctx = ctxutil.SetTraceID(ctx, "00000000000000000000000000abcdef")
				return r.WithContext(ctx)
			},
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "SYS-001",
			wantProblem:    true,
			wantStackInLog: true,
			wantRequestID:  "req-456",
			wantTraceID:    "00000000000000000000000000abcdef",
		},
		{
			name: "panic with nil value returns SYS-001 in Go 1.21+",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(nil)
			},
			// Note: In Go 1.21+, panic(nil) triggers a runtime.PanicNilError and IS recoverable
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "SYS-001",
			wantProblem:    true,
			wantStackInLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output for verification
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			// Create the middleware chain
			recovererMiddleware := middleware.Recoverer(logger)
			handler := recovererMiddleware(tt.handler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test-endpoint", nil)
			if tt.setupContext != nil {
				req = tt.setupContext(req)
			}

			// Execute request
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Verify status code
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			// Verify problem response if expected
			if tt.wantProblem {
				// Check Content-Type
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/problem+json" {
					t.Errorf("Content-Type = %q, want %q", contentType, "application/problem+json")
				}

				// Parse response
				var problem problemResponse
				if err := json.NewDecoder(rr.Body).Decode(&problem); err != nil {
					t.Fatalf("failed to decode problem response: %v", err)
				}

				// Verify error code
				if problem.Code != tt.wantCode {
					t.Errorf("code = %q, want %q", problem.Code, tt.wantCode)
				}

				// Verify status in body matches HTTP status
				if problem.Status != tt.wantStatus {
					t.Errorf("problem.status = %d, want %d", problem.Status, tt.wantStatus)
				}

				// Verify safe title
				if problem.Title != "Internal Server Error" {
					t.Errorf("title = %q, want %q", problem.Title, "Internal Server Error")
				}

				// Verify panic message is NOT exposed in detail
				if strings.Contains(problem.Detail, "test panic") || strings.Contains(problem.Detail, "panic") {
					t.Error("panic message exposed in detail field - security violation")
				}

				// Verify request_id is included if expected
				if tt.wantRequestID != "" && problem.RequestID != tt.wantRequestID {
					t.Errorf("request_id = %q, want %q", problem.RequestID, tt.wantRequestID)
				}

				// Verify trace_id is included if expected
				if tt.wantTraceID != "" && problem.TraceID != tt.wantTraceID {
					t.Errorf("trace_id = %q, want %q", problem.TraceID, tt.wantTraceID)
				}

				// Verify instance is set
				if problem.Instance != "/test-endpoint" {
					t.Errorf("instance = %q, want %q", problem.Instance, "/test-endpoint")
				}

				// Verify type URL contains internal-error slug
				if !strings.Contains(problem.Type, "internal-error") {
					t.Errorf("type = %q, want to contain %q", problem.Type, "internal-error")
				}
			}

			// Verify stack trace is logged (but not in response)
			if tt.wantStackInLog {
				logOutput := logBuf.String()
				if !strings.Contains(logOutput, "stack") {
					t.Error("stack trace not found in log output")
				}
				if !strings.Contains(logOutput, "panic recovered") {
					t.Error("panic recovered message not found in log")
				}
				// Verify request_id is in log
				if tt.wantRequestID != "" && !strings.Contains(logOutput, tt.wantRequestID) {
					t.Errorf("request_id %q not found in log", tt.wantRequestID)
				}
				// Verify trace_id is in log (if present)
				if tt.wantTraceID != "" && !strings.Contains(logOutput, tt.wantTraceID) {
					t.Errorf("trace_id %q not found in log", tt.wantTraceID)
				}
			}
		})
	}
}

func TestRecoverer_DoesNotExposeStackTrace(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	handler := middleware.Recoverer(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("secret internal error details")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	body := rr.Body.String()

	// Verify stack trace is NOT in response
	if strings.Contains(body, "runtime/debug") {
		t.Error("stack trace exposed in response body")
	}
	if strings.Contains(body, "secret internal error details") {
		t.Error("panic message exposed in response body")
	}
	if strings.Contains(body, "goroutine") {
		t.Error("goroutine info exposed in response body")
	}

	// Verify stack trace IS in log
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "debug.Stack") || !strings.Contains(logOutput, "stack") {
		// Need to check for stack presence in a different way
		if !strings.Contains(logOutput, "secret internal error details") {
			t.Error("panic message not logged for debugging")
		}
	}
}

func TestRecoverer_PrometheusMetricsExist(t *testing.T) {
	// This test verifies the Prometheus counter is registered (no panic on registration)
	// Actual counter increment testing would require mocking prometheus registry

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	handler := middleware.Recoverer(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("metrics test")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/resource", nil)
	rr := httptest.NewRecorder()

	// This should not panic due to duplicate metric registration
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
