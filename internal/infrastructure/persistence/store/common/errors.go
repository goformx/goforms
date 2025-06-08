package common

import (
	"errors"
	"fmt"
)

// Common store errors
var (
	ErrNotFound      = errors.New("record not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrDatabaseError = errors.New("database error")
)

// StoreError represents a store operation error
type StoreError struct {
	Op      string // Operation that failed
	Entity  string // Entity type (e.g., "user", "form")
	ID      string // Entity ID
	Err     error  // The underlying error
	Details string // Additional error details
}

// Error implements the error interface
func (e *StoreError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s %s", e.Op, e.Entity, e.ID)
	}
	return fmt.Sprintf("%s: %s %s: %v", e.Op, e.Entity, e.ID, e.Err)
}

// Unwrap returns the underlying error
func (e *StoreError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(op, entity, id string) error {
	return &StoreError{
		Op:     op,
		Entity: entity,
		ID:     id,
		Err:    ErrNotFound,
	}
}

// NewInvalidInputError creates a new invalid input error
func NewInvalidInputError(op, entity, id string, err error) error {
	return &StoreError{
		Op:     op,
		Entity: entity,
		ID:     id,
		Err:    err,
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(op, entity, id string, err error) error {
	return &StoreError{
		Op:     op,
		Entity: entity,
		ID:     id,
		Err:    err,
	}
}
