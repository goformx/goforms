// Package common provides shared utilities and types for repository implementations
// including error handling, pagination, and common data structures.
package common

// PaginationParams represents the parameters for pagination
type PaginationParams struct {
	Page     int
	PageSize int
}

// PaginationResult represents the result of a paginated query
type PaginationResult struct {
	Items      any // The items in the current page
	TotalItems int // Total number of items
	Page       int // Current page number
	PageSize   int // Number of items per page
	TotalPages int // Total number of pages
}

// NewPaginationParams creates a new PaginationParams with default values
func NewPaginationParams(page, pageSize int) PaginationParams {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// GetOffset calculates the offset for the current page
func (p PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size as the limit
func (p PaginationParams) GetLimit() int {
	return p.PageSize
}

// NewPaginationResult creates a new PaginationResult
func NewPaginationResult(items any, totalItems, page, pageSize int) PaginationResult {
	totalPages := (totalItems + pageSize - 1) / pageSize

	return PaginationResult{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
