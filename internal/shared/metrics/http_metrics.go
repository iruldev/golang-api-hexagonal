package metrics

// HTTPMetrics defines the minimal contract needed by HTTP middleware to record metrics.
// Keeping this in a shared package avoids transport → infra imports.
type HTTPMetrics interface {
	IncRequest(method, route, status string)
	ObserveRequestDuration(method, route string, seconds float64)
}
