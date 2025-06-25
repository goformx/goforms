// Package model contains domain models and error definitions for forms.
package model

import "errors"

var (
	// ErrFormTitleRequired is returned when a form is created without a title
	ErrFormTitleRequired = errors.New("form title is required")

	// ErrFormSchemaRequired is returned when a form is created without a schema
	ErrFormSchemaRequired = errors.New("form schema is required")

	// ErrFormNotFound is returned when a form cannot be found
	ErrFormNotFound = errors.New("form not found")

	// ErrFormInvalid is returned when a form is invalid
	ErrFormInvalid = errors.New("form is invalid")

	// ErrSubmissionNotFound is returned when a form submission cannot be found
	ErrSubmissionNotFound = errors.New("form submission not found")
)
