package errors

import "errors"

// ErrorCategory represents a high-level category of errors
type ErrorCategory string

const (
	CategoryNotFound       ErrorCategory = "not_found"
	CategoryValidation     ErrorCategory = "validation"
	CategoryForm           ErrorCategory = "form"
	CategoryUser           ErrorCategory = "user"
	CategoryAuthentication ErrorCategory = "authentication"
	CategorySystem         ErrorCategory = "system"
	CategoryConflict       ErrorCategory = "conflict"
	CategoryForbidden      ErrorCategory = "forbidden"
)

// errorCategories maps error codes to their categories
var errorCategories = map[ErrorCode][]ErrorCategory{
	// Not found errors
	ErrCodeNotFound:     {CategoryNotFound},
	ErrCodeFormNotFound: {CategoryNotFound, CategoryForm},
	ErrCodeUserNotFound: {CategoryNotFound, CategoryUser},

	// Validation errors
	ErrCodeValidation:     {CategoryValidation},
	ErrCodeRequired:       {CategoryValidation},
	ErrCodeInvalid:        {CategoryValidation},
	ErrCodeInvalidFormat:  {CategoryValidation},
	ErrCodeInvalidInput:   {CategoryValidation},
	ErrCodeBadRequest:     {CategoryValidation},
	ErrCodeFormValidation: {CategoryValidation, CategoryForm},
	ErrCodeFormInvalid:    {CategoryValidation, CategoryForm},
	ErrCodeUserInvalid:    {CategoryValidation, CategoryUser},
	ErrCodeFormSubmission: {CategoryValidation, CategoryForm},
	ErrCodeFormExpired:    {CategoryValidation, CategoryForm},
	ErrCodeUserDisabled:   {CategoryValidation, CategoryUser},

	// Form errors
	ErrCodeFormAccessDenied: {CategoryForm, CategoryForbidden},

	// User errors
	ErrCodeUserExists:       {CategoryUser, CategoryConflict},
	ErrCodeUserUnauthorized: {CategoryUser, CategoryAuthentication},

	// Authentication errors
	ErrCodeUnauthorized:     {CategoryAuthentication},
	ErrCodeAuthentication:   {CategoryAuthentication},
	ErrCodeInsufficientRole: {CategoryAuthentication, CategoryForbidden},

	// Forbidden errors
	ErrCodeForbidden: {CategoryForbidden},

	// Conflict errors
	ErrCodeConflict:      {CategoryConflict},
	ErrCodeAlreadyExists: {CategoryConflict},

	// System errors
	ErrCodeServerError: {CategorySystem},
	ErrCodeDatabase:    {CategorySystem},
	ErrCodeConfig:      {CategorySystem},
	ErrCodeStartup:     {CategorySystem},
	ErrCodeShutdown:    {CategorySystem},
	ErrCodeTimeout:     {CategorySystem},
}

// HasCategory checks if an error belongs to a specific category
func HasCategory(err error, category ErrorCategory) bool {
	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		return false
	}

	categories, ok := errorCategories[domainErr.Code]
	if !ok {
		return false
	}

	for _, c := range categories {
		if c == category {
			return true
		}
	}

	return false
}

// GetCategories returns all categories for an error
func GetCategories(err error) []ErrorCategory {
	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		return nil
	}

	return errorCategories[domainErr.Code]
}
