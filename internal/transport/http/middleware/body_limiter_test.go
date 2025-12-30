//go:build !integration

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

func TestBodyLimiter_AllowsWithinLimitAndRehydratesBody(t *testing.T) {
	t.Parallel()

	limited := BodyLimiter(1024)

	var capturedBody string
	handlerCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		capturedBody = string(bodyBytes)
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	body := bytes.Repeat([]byte("a"), 512)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	rec := httptest.NewRecorder()

	limited(handler).ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(body), capturedBody)
}

func TestBodyLimiter_RejectsWhenContentLengthExceedsLimit(t *testing.T) {
	t.Parallel()

	limited := BodyLimiter(100)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := bytes.Repeat([]byte("a"), 200)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	rec := httptest.NewRecorder()

	limited(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)

	var problem contract.ProblemDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
	assert.Equal(t, app.CodeRequestTooLarge, problem.Code)
}

func TestBodyLimiter_RejectsWhenStreamingExceedsLimit(t *testing.T) {
	t.Parallel()

	limited := BodyLimiter(256)
	handlerCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	body := bytes.Repeat([]byte("b"), 1024)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	// Leave ContentLength unset to mimic chunked/unknown length (defaults to -1).
	req.ContentLength = -1
	rec := httptest.NewRecorder()

	limited(handler).ServeHTTP(rec, req)

	assert.False(t, handlerCalled, "handler should not be called when body exceeds limit")
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)

	var problem contract.ProblemDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
	assert.Equal(t, app.CodeRequestTooLarge, problem.Code)
}

func TestBodyLimiter_SkipsWhenLimitDisabled(t *testing.T) {
	t.Parallel()

	limited := BodyLimiter(0)
	handlerCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	limited(handler).ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBodyLimiter_StreamDoesNotBufferFullBody(t *testing.T) {
	t.Parallel()

	limited := BodyLimiter(1024)
	handlerCalled := false
	var bodyLen int

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyLen = len(data)
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// 900 bytes < limit to ensure streaming path works without buffering full >limit payload.
	body := strings.Repeat("x", 900)
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	rec := httptest.NewRecorder()

	limited(handler).ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, len(body), bodyLen)
}
