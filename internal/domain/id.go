// Package domain contains the core business entities and interfaces.
// This package follows hexagonal architecture principles and must remain
// independent of infrastructure concerns (database, HTTP, logging, etc.).
//
// All types in this package should only depend on Go standard library.
package domain

// ID is a minimal identifier type kept stdlib-only for domain boundary compliance.
// It represents a domain entity identifier without coupling to specific UUID
// implementations or external libraries.
type ID string

// String returns the string representation of the ID.
func (id ID) String() string {
	return string(id)
}

// IsEmpty returns true if the ID is empty.
func (id ID) IsEmpty() bool {
	return id == ""
}
