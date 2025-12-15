package wrapper

import (
	"context"
	"io"
	"net/http"
	"time"
)

// DoRequest wraps http.Client.Do with context timeout enforcement.
// If ctx has no deadline, DefaultHTTPTimeout is applied.
// Returns immediately if context is already cancelled.
//
// The request is cloned with the provided context using req.WithContext(ctx).
// The returned response Body is wrapped to ensure the timeout context is cancelled
// only when the Body is closed.
func DoRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Add timeout only if no deadline is set
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, DefaultHTTPTimeout)
	}

	// Clone request with the (possibly timeout-wrapped) context
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	// If we created a cancel function, wrap the body to ensure it's called on Close
	if cancel != nil {
		resp.Body = &cancelBody{
			ReadCloser: resp.Body,
			cancel:     cancel,
		}
	}

	return resp, nil
}

// HTTPClient interface defines the methods used by wrappers.
// This allows easier testing with mock implementations.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DoRequestWithClient wraps HTTPClient.Do with context timeout enforcement.
// This variant accepts an interface for easier testing.
// If ctx has no deadline, DefaultHTTPTimeout is applied.
// Returns immediately if context is already cancelled.
func DoRequestWithClient(ctx context.Context, client HTTPClient, req *http.Request) (*http.Response, error) {
	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Add timeout only if no deadline is set
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, DefaultHTTPTimeout)
	}

	// Clone request with the (possibly timeout-wrapped) context
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	// If we created a cancel function, wrap the body to ensure it's called on Close
	if cancel != nil {
		resp.Body = &cancelBody{
			ReadCloser: resp.Body,
			cancel:     cancel,
		}
	}

	return resp, nil
}

// TimeoutTransport wraps an http.RoundTripper to ensure context timeout.
// This can be used when you want to enforce timeouts at the transport level.
type TimeoutTransport struct {
	Transport http.RoundTripper
	Timeout   time.Duration
}

// RoundTrip implements http.RoundTripper.
func (t *TimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Add timeout only if no deadline is set
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		timeout := t.Timeout
		if timeout == 0 {
			timeout = DefaultHTTPTimeout
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		req = req.WithContext(ctx)
	}

	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	// If we created a cancel function, wrap the body to ensure it's called on Close
	if cancel != nil {
		resp.Body = &cancelBody{
			ReadCloser: resp.Body,
			cancel:     cancel,
		}
	}

	return resp, nil
}

// cancelBody wraps io.ReadCloser to call a specific cancel function on Close.
type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelBody) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}
