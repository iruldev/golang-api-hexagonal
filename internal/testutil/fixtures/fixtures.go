// Package fixtures provides test data builders and factories.
//
// This package will contain builder pattern helpers for creating
// test data with sensible defaults that can be overridden:
//   - User builders
//   - Entity factories
//   - Random data generators
//
// Example usage (planned):
//
//	user := fixtures.NewUserBuilder().
//	    WithEmail("test@example.com").
//	    Build()
package fixtures
