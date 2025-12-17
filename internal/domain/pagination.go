package domain

// DefaultPageSize is the default number of items per page when PageSize is not specified.
const DefaultPageSize = 20

// MaxPageSize is an upper bound to prevent unbounded list requests.
const MaxPageSize = 100

// ListParams contains pagination parameters for list operations.
// Page is 1-based (first page is 1, not 0).
type ListParams struct {
	Page     int
	PageSize int
}

// Offset calculates the database offset for the current page.
// Returns 0 for page <= 0.
func (p ListParams) Offset() int {
	page := p.Page
	if page <= 0 {
		page = 1
	}
	return (page - 1) * p.Limit()
}

// Limit returns the page size, defaulting to DefaultPageSize if not specified.
func (p ListParams) Limit() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return p.PageSize
}
