package wrapper

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockHTTPClient implements HTTPClient for testing
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

func TestDoRequest_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	// Use a mock RoundTripper to capture the request context
	transport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			expected := time.Now().Add(DefaultHTTPTimeout)
			if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
				t.Errorf("deadline %v not within expected range", deadline)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	}

	client := &http.Client{Transport: transport}
	ctx := context.Background()
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := DoRequest(ctx, client, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestDoRequest_WithDeadline_PreservesExisting(t *testing.T) {
	t.Parallel()

	existingDeadline := time.Now().Add(5 * time.Second)

	transport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Error("expected deadline to be preserved")
			}
			// Should be approximately the existing deadline (within 1 second tolerance)
			if deadline.Before(existingDeadline.Add(-time.Second)) || deadline.After(existingDeadline.Add(time.Second)) {
				t.Errorf("deadline %v should be close to existing %v", deadline, existingDeadline)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	}

	client := &http.Client{Transport: transport}
	ctx, cancel := context.WithDeadline(context.Background(), existingDeadline)
	defer cancel()

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := DoRequest(ctx, client, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestDoRequest_CancelledContext_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	transport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			t.Error("request should not reach transport with cancelled context")
			return nil, nil
		},
	}

	client := &http.Client{Transport: transport}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	_, err := DoRequest(ctx, client, req)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDoRequest_DeadlineExceeded_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	transport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			t.Error("request should not reach transport with exceeded deadline")
			return nil, nil
		},
	}

	client := &http.Client{Transport: transport}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	_, err := DoRequest(ctx, client, req)
	if err == nil {
		t.Error("expected error for exceeded deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestDoRequestWithClient_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			expected := time.Now().Add(DefaultHTTPTimeout)
			if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
				t.Errorf("deadline %v not within expected range", deadline)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	}

	ctx := context.Background()
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := DoRequestWithClient(ctx, mock, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

func TestDoRequestWithClient_CancelledContext_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			t.Error("request should not be made with cancelled context")
			return nil, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	_, err := DoRequestWithClient(ctx, mock, req)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestTimeoutTransport_NoDeadline_AddsTimeout(t *testing.T) {
	t.Parallel()

	underlyingTransport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			expected := time.Now().Add(5 * time.Second)
			if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
				t.Errorf("deadline %v not within expected range", deadline)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	}

	transport := &TimeoutTransport{
		Transport: underlyingTransport,
		Timeout:   5 * time.Second,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestTimeoutTransport_WithDeadline_PreservesExisting(t *testing.T) {
	t.Parallel()

	existingDeadline := time.Now().Add(2 * time.Second)

	underlyingTransport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Error("expected deadline to be preserved")
			}
			// Deadline should be approximately the existing one
			if deadline.Before(existingDeadline.Add(-time.Second)) || deadline.After(existingDeadline.Add(time.Second)) {
				t.Errorf("deadline %v should be close to existing %v", deadline, existingDeadline)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	}

	transport := &TimeoutTransport{
		Transport: underlyingTransport,
		Timeout:   30 * time.Second,
	}

	ctx, cancel := context.WithDeadline(context.Background(), existingDeadline)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestTimeoutTransport_CancelledContext_ReturnsError(t *testing.T) {
	t.Parallel()

	transport := &TimeoutTransport{
		Transport: http.DefaultTransport,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)

	_, err := transport.RoundTrip(req)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestTimeoutTransport_DefaultTimeout_WhenZero(t *testing.T) {
	t.Parallel()

	underlyingTransport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			expected := time.Now().Add(DefaultHTTPTimeout)
			if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
				t.Errorf("deadline %v not within expected range, expected ~%v", deadline, expected)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	}

	transport := &TimeoutTransport{
		Transport: underlyingTransport,
		Timeout:   0, // should use DefaultHTTPTimeout
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestDoRequest_IntegrationWithRealServer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))
	defer server.Close()

	ctx := context.Background()
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := DoRequest(ctx, server.Client(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
