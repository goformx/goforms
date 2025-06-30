// Package core provides the core interfaces and types for middleware functionality.
package core

import "time"

// ChainType represents different types of middleware chains.
// Each type corresponds to a specific use case or request pattern.
type ChainType int

const (
	// ChainTypeDefault represents the default middleware chain for most requests.
	ChainTypeDefault ChainType = iota

	// ChainTypeAPI represents middleware chain for API requests.
	ChainTypeAPI

	// ChainTypeWeb represents middleware chain for web page requests.
	ChainTypeWeb

	// ChainTypeAuth represents middleware chain for authentication endpoints.
	ChainTypeAuth

	// ChainTypeAdmin represents middleware chain for admin-only endpoints.
	ChainTypeAdmin

	// ChainTypePublic represents middleware chain for public endpoints.
	ChainTypePublic

	// ChainTypeStatic represents middleware chain for static asset requests.
	ChainTypeStatic
)

// String returns the string representation of the chain type.
func (ct ChainType) String() string {
	switch ct {
	case ChainTypeDefault:
		return "default"
	case ChainTypeAPI:
		return "api"
	case ChainTypeWeb:
		return "web"
	case ChainTypeAuth:
		return "auth"
	case ChainTypeAdmin:
		return "admin"
	case ChainTypePublic:
		return "public"
	case ChainTypeStatic:
		return "static"
	default:
		return "unknown"
	}
}

// Error represents a middleware-specific error.
// Provides additional context about middleware failures.
type Error struct {
	// Code is the error code for programmatic handling.
	Code string

	// Message is the human-readable error message.
	Message string

	// Middleware is the name of the middleware that generated the error.
	Middleware string

	// Cause is the underlying error that caused this middleware error.
	Cause error

	// Timestamp is when the error occurred.
	Timestamp time.Time
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new middleware error.
func NewError(code, message, middleware string, cause error) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		Middleware: middleware,
		Cause:      cause,
		Timestamp:  time.Now(),
	}
}
