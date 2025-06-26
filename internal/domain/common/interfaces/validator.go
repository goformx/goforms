// Package interfaces provides common interfaces used throughout the domain layer
// for validation, repositories, and other cross-cutting concerns.
package interfaces

import "github.com/go-playground/validator/v10"

// Validator defines the interface for validation operations
type Validator interface {
	// Struct validates a struct based on validation tags
	Struct(any) error
	// Var validates a single variable using a tag
	Var(any, string) error
	// RegisterValidation adds a custom validation with the given tag
	RegisterValidation(string, func(fl validator.FieldLevel) bool) error
	// RegisterStructValidation adds a struct-level validation
	RegisterStructValidation(func(sl validator.StructLevel), any) error
	// RegisterCrossFieldValidation adds a cross-field validation
	RegisterCrossFieldValidation(tag string, fn func(fl validator.FieldLevel) bool) error
	// GetErrors returns detailed validation errors
	GetErrors(err error) map[string]string
	// ValidateStruct is an alias for Struct for backward compatibility
	ValidateStruct(any) error
}
