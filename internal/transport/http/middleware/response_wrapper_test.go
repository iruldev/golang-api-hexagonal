package middleware

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseWrapper_DefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapper := NewResponseWrapper(rec)

	// Before any WriteHeader call, status should be 200 (default)
	assert.Equal(t, http.StatusOK, wrapper.Status())
}

func TestResponseWrapper_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapper := NewResponseWrapper(rec)

	wrapper.WriteHeader(http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, wrapper.Status())
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestResponseWrapper_Write(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapper := NewResponseWrapper(rec)

	data := []byte("hello world")
	n, err := wrapper.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), wrapper.BytesWritten())
	assert.Equal(t, "hello world", rec.Body.String())
}

func TestResponseWrapper_MultipleWrites(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapper := NewResponseWrapper(rec)

	_, _ = wrapper.Write([]byte("hello"))
	_, _ = wrapper.Write([]byte(" "))
	_, _ = wrapper.Write([]byte("world"))

	assert.Equal(t, 11, wrapper.BytesWritten()) // 5 + 1 + 5
	assert.Equal(t, "hello world", rec.Body.String())
}

func TestResponseWrapper_Unwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapper := NewResponseWrapper(rec)

	underlying := wrapper.Unwrap()
	assert.Equal(t, rec, underlying)
}

func TestResponseWrapper_WriteHeaderNotCalled(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapper := NewResponseWrapper(rec)

	// Write without calling WriteHeader - should use default status
	_, _ = wrapper.Write([]byte("data"))

	// wrapper retains default 200
	assert.Equal(t, http.StatusOK, wrapper.Status())
}

type flushRecorder struct {
	http.ResponseWriter
	flushed bool
}

func (f *flushRecorder) Flush() { f.flushed = true }

func TestResponseWrapper_FlushPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	fr := &flushRecorder{ResponseWriter: rec}
	wrapper := NewResponseWrapper(fr)

	wrapper.Flush()

	assert.True(t, fr.flushed, "flush should delegate to underlying Flusher")
}

type hijackableRecorder struct {
	http.ResponseWriter
	hijacked bool
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacked = true
	return nil, nil, nil
}

func TestResponseWrapper_HijackPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	hr := &hijackableRecorder{ResponseWriter: rec}
	wrapper := NewResponseWrapper(hr)

	_, _, err := wrapper.Hijack()

	assert.NoError(t, err)
	assert.True(t, hr.hijacked, "hijack should delegate to underlying Hijacker")
}

type pushRecorder struct {
	http.ResponseWriter
	pushed []string
	err    error
}

func (p *pushRecorder) Push(target string, _ *http.PushOptions) error {
	p.pushed = append(p.pushed, target)
	return p.err
}

func TestResponseWrapper_PushPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	pr := &pushRecorder{ResponseWriter: rec}
	wrapper := NewResponseWrapper(pr)

	err := wrapper.Push("/foo", nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/foo"}, pr.pushed)
}
