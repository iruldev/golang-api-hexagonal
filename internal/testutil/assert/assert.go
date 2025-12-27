// Package assert provides test assertion helpers using go-cmp.
//
// This package will contain assertion helpers that wrap go-cmp for
// clean test assertions with readable diffs.
//
// Planned helpers (to be implemented in future stories):
//   - Diff: compare two values and return readable diff
//   - Equal: assert two values are equal
//   - ErrorIs: assert error matches expected
//   - ErrorAs: assert error can be cast to type
package assert
