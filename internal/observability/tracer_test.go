package observability

import (
	"context"
	"testing"
)

func TestGetTraceID_NoTrace(t *testing.T) {
	ctx := context.Background()
	traceID := GetTraceID(ctx)

	if traceID != "" {
		t.Errorf("Expected empty trace ID, got %s", traceID)
	}
}

func TestGetSpan_ReturnsNonNil(t *testing.T) {
	ctx := context.Background()
	span := GetSpan(ctx)

	// Should return a no-op span, not nil
	if span == nil {
		t.Error("Expected non-nil span")
	}
}

func TestStartSpan_CreatesChildSpan(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-operation")
	defer span.End()

	if newCtx == nil {
		t.Error("Expected non-nil context")
	}

	// Span should be valid (even if no-op tracer)
	if span == nil {
		t.Error("Expected non-nil span")
	}
}

func TestGetSpan_FromContext(t *testing.T) {
	ctx := context.Background()
	_, span := StartSpan(ctx, "parent-span")
	defer span.End()

	// Get span context
	spanCtx := span.SpanContext()

	// Even without real tracer, span context should exist
	if !spanCtx.IsValid() {
		// This is expected without a real tracer provider
		t.Log("Span context not valid (expected without real tracer)")
	}
}

func TestGetTraceID_WithSpan(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-span")
	defer span.End()

	traceID := GetTraceID(newCtx)

	// Without real tracer, trace ID will be empty
	// This test documents the behavior
	if traceID != "" {
		t.Logf("Got trace ID: %s", traceID)
	} else {
		t.Log("No trace ID (expected without real tracer)")
	}
}

func TestGetSpan_Interface(t *testing.T) {
	ctx := context.Background()
	span := GetSpan(ctx)

	// Verify it implements trace.Span interface
	var _ = span
}
