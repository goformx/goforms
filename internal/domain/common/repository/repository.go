// Package repository provides common repository interfaces and utilities
// for data access patterns used throughout the domain layer.
package repository

import (
	"context"
)

// Entity represents a domain entity
type Entity interface {
	// GetID returns the entity's ID
	GetID() string
	// SetID sets the entity's ID
	SetID(id string)
	// Validate validates the entity
	Validate() error
}

// Repository defines the base interface for all repositories
type Repository[T Entity] interface {
	// Create creates a new entity
	Create(ctx context.Context, entity T) error
	// GetByID gets an entity by ID
	GetByID(ctx context.Context, id string) (T, error)
	// Update updates an entity
	Update(ctx context.Context, entity T) error
	// Delete deletes an entity
	Delete(ctx context.Context, id string) error
	// List lists entities with pagination
	List(ctx context.Context, offset, limit int) ([]T, error)
	// Count returns the total number of entities
	Count(ctx context.Context) (int, error)
	// Search searches entities
	Search(ctx context.Context, query string, offset, limit int) ([]T, error)
}

// QueryBuilder defines the interface for building queries
type QueryBuilder interface {
	// Where adds a where clause
	Where(field string, operator string, value any) QueryBuilder
	// OrderBy adds an order by clause
	OrderBy(field string, direction string) QueryBuilder
	// Limit sets the limit
	Limit(limit int) QueryBuilder
	// Offset sets the offset
	Offset(offset int) QueryBuilder
	// Build builds the query
	Build() (string, []any, error)
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int
	PageSize int
	Total    int
}

// NewPagination creates a new pagination
func NewPagination(page, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	return Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    0,
	}
}

// GetOffset returns the offset for the pagination
func (p Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the limit for the pagination
func (p Pagination) GetLimit() int {
	return p.PageSize
}

// GetTotalPages returns the total number of pages
func (p Pagination) GetTotalPages() int {
	if p.Total == 0 {
		return 0
	}

	return (p.Total + p.PageSize - 1) / p.PageSize
}

// HasNextPage returns whether there is a next page
func (p Pagination) HasNextPage() bool {
	return p.Page < p.GetTotalPages()
}

// HasPreviousPage returns whether there is a previous page
func (p Pagination) HasPreviousPage() bool {
	return p.Page > 1
}
