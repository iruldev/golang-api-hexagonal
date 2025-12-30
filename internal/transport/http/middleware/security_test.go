package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureHeaders_BasicHeaders(t *testing.T) {
	// Arrange
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	// Assert
	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify all required security headers are present
	assert.Equal(t, nosniff, resp.Header.Get(headerXContentTypeOptions), "X-Content-Type-Options should be nosniff")
	assert.Equal(t, deny, resp.Header.Get(headerXFrameOptions), "X-Frame-Options should be DENY")
	assert.Equal(t, xssBlock, resp.Header.Get(headerXXSSProtection), "X-XSS-Protection should be 1; mode=block")
	assert.Equal(t, cspDefaultNone, resp.Header.Get(headerCSP), "Content-Security-Policy should be restrictive")
	assert.Equal(t, referrerStrictOrigin, resp.Header.Get(headerReferrerPolicy), "Referrer-Policy should be strict")
}

func TestSecureHeaders_NoHSTSWithoutHTTPS(t *testing.T) {
	// Arrange
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	// Assert - HSTS should NOT be present when not over HTTPS
	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Empty(t, resp.Header.Get(headerHSTS), "HSTS should not be set without HTTPS")
}

func TestSecureHeaders_HSTSWithTLS(t *testing.T) {
	// Arrange
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Simulate direct TLS connection
	req.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()

	// Act
	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	// Assert - HSTS should be present when over TLS
	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, hstsMaxAgeWithSubdoms, resp.Header.Get(headerHSTS), "HSTS should be set when TLS is active")
}

func TestSecureHeaders_HSTSWithXForwardedProto(t *testing.T) {
	// Arrange
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(headerXForwardedProto, "https")
	rec := httptest.NewRecorder()

	// Act
	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	// Assert - HSTS should be present when behind HTTPS proxy
	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, hstsMaxAgeWithSubdoms, resp.Header.Get(headerHSTS), "HSTS should be set when X-Forwarded-Proto is https")
}

func TestSecureHeaders_HSTSWithUppercaseXForwardedProto(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(headerXForwardedProto, "HTTPS")
	rec := httptest.NewRecorder()

	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, hstsMaxAgeWithSubdoms, resp.Header.Get(headerHSTS), "HSTS should be set when X-Forwarded-Proto is HTTPS (case-insensitive)")
}

func TestSecureHeaders_HSTSWithURLScheme(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.URL = &url.URL{Scheme: "https", Path: "/test"}
	rec := httptest.NewRecorder()

	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, hstsMaxAgeWithSubdoms, resp.Header.Get(headerHSTS), "HSTS should be set when URL scheme is https")
}

func TestSecureHeaders_ErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		setupReq   func(*http.Request)
	}{
		{"bad request 400", http.StatusBadRequest, func(r *http.Request) {}},
		{"not found 404", http.StatusNotFound, func(r *http.Request) {}},
		{"internal error 500", http.StatusInternalServerError, func(r *http.Request) {}},
		{"service unavailable 503", http.StatusServiceUnavailable, func(r *http.Request) {}},
		{"internal error 500 with TLS", http.StatusInternalServerError, func(r *http.Request) {
			r.TLS = &tls.ConnectionState{}
		}},
		{"internal error 500 with forwarded https", http.StatusInternalServerError, func(r *http.Request) {
			r.Header.Set(headerXForwardedProto, "https")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupReq(req)
			rec := httptest.NewRecorder()

			// Act
			handler := SecureHeaders(next)
			handler.ServeHTTP(rec, req)

			// Assert - All security headers should be present even on errors
			resp := rec.Result()
			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tt.statusCode, resp.StatusCode)
			assert.Equal(t, nosniff, resp.Header.Get(headerXContentTypeOptions))
			assert.Equal(t, deny, resp.Header.Get(headerXFrameOptions))
			assert.Equal(t, xssBlock, resp.Header.Get(headerXXSSProtection))
			assert.Equal(t, cspDefaultNone, resp.Header.Get(headerCSP))
			assert.Equal(t, referrerStrictOrigin, resp.Header.Get(headerReferrerPolicy))
			if tt.name == "internal error 500 with TLS" || tt.name == "internal error 500 with forwarded https" {
				assert.Equal(t, hstsMaxAgeWithSubdoms, resp.Header.Get(headerHSTS))
			}
		})
	}
}

func TestSecureHeaders_HeadersSetBeforeNextHandler(t *testing.T) {
	// Arrange - verify headers are available to downstream handlers
	var capturedHeaders http.Header

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture headers that have been set at this point
		capturedHeaders = w.Header().Clone()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	// Assert - Headers should already be set when next handler runs
	assert.Equal(t, nosniff, capturedHeaders.Get(headerXContentTypeOptions))
	assert.Equal(t, deny, capturedHeaders.Get(headerXFrameOptions))
	assert.Equal(t, xssBlock, capturedHeaders.Get(headerXXSSProtection))
	assert.Equal(t, cspDefaultNone, capturedHeaders.Get(headerCSP))
	assert.Equal(t, referrerStrictOrigin, capturedHeaders.Get(headerReferrerPolicy))
}

func TestSecureHeaders_HTTPMethodsSupported(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			handler := SecureHeaders(next)
			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, nosniff, resp.Header.Get(headerXContentTypeOptions))
		})
	}
}

func TestSecureHeaders_HSTSForcedEnabled(t *testing.T) {
	t.Setenv(envHSTSEnabled, "true")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, hstsMaxAgeWithSubdoms, resp.Header.Get(headerHSTS), "HSTS should be set when explicitly enabled via env")
}

func TestSecureHeaders_HSTSForcedDisabled(t *testing.T) {
	t.Setenv(envHSTSEnabled, "false")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Simulate TLS connection before middleware runs
	req.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()

	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, hstsMaxAgeWithSubdoms, resp.Header.Get(headerHSTS), "HSTS should still be set on HTTPS even if env disables it")
}

func TestSecureHeaders_HSTSForcedDisabledOnPlainHTTP(t *testing.T) {
	t.Setenv(envHSTSEnabled, "false")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	rec := httptest.NewRecorder()

	handler := SecureHeaders(next)
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	assert.Empty(t, resp.Header.Get(headerHSTS), "HSTS should not be set when explicitly disabled and request is plain HTTP")
}

func TestIsHTTPS(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func(*http.Request)
		want     bool
	}{
		{
			name:     "no TLS, no proxy header",
			setupReq: func(r *http.Request) {},
			want:     false,
		},
		{
			name: "direct TLS connection",
			setupReq: func(r *http.Request) {
				r.TLS = &tls.ConnectionState{}
			},
			want: true,
		},
		{
			name: "X-Forwarded-Proto https",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXForwardedProto, "https")
			},
			want: true,
		},
		{
			name: "X-Forwarded-Proto http",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXForwardedProto, "http")
			},
			want: false,
		},
		{
			name: "X-Forwarded-Proto HTTPS uppercase",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXForwardedProto, "HTTPS")
			},
			want: true,
		},
		{
			name: "X-Forwarded-Proto comma-separated https first",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXForwardedProto, "https,http")
			},
			want: true,
		},
		{
			name: "URL scheme https",
			setupReq: func(r *http.Request) {
				r.URL = &url.URL{Scheme: "https", Path: "/test"}
			},
			want: true,
		},
		{
			name: "both TLS and proxy header",
			setupReq: func(r *http.Request) {
				r.TLS = &tls.ConnectionState{}
				r.Header.Set(headerXForwardedProto, "https")
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupReq(req)

			got := isHTTPS(req)
			assert.Equal(t, tt.want, got)
		})
	}
}
