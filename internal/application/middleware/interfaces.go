// Package middleware provides framework-agnostic middleware interfaces and abstractions
// for the GoForms application. This package defines the core contracts that all
// middleware implementations must follow, enabling clean architecture principles
// and framework independence.
package middleware

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/core"
)

// Middleware defines the core interface for all middleware components.
// This interface is framework-agnostic and follows clean architecture principles.
type Middleware interface {
	// Process handles the middleware logic for a given request.
	// It receives the request context, request data, and the next handler in the chain.
	// Returns a response that can be processed by upstream middleware or handlers.
	Process(ctx context.Context, req Request, next Handler) Response

	// Name returns the unique identifier for this middleware.
	// Used for logging, debugging, and middleware registry management.
	Name() string

	// Priority returns the execution priority of this middleware.
	// Lower numbers indicate higher priority (executed first).
	// Middleware with the same priority are executed in registration order.
	Priority() int
}

// Handler represents the next handler in the middleware chain.
// This is a function type that processes a request and returns a response.
type Handler func(ctx context.Context, req Request) Response

// Registry manages middleware registration and discovery.
// Provides a centralized way to register, retrieve, and manage middleware components.
type Registry interface {
	// Register adds middleware to the registry with a unique name.
	// Returns an error if middleware with the same name already exists.
	Register(name string, middleware Middleware) error

	// Get retrieves middleware by name from the registry.
	// Returns nil and false if middleware is not found.
	Get(name string) (Middleware, bool)

	// List returns all registered middleware names.
	List() []string

	// Remove removes middleware by name from the registry.
	// Returns true if middleware was found and removed.
	Remove(name string) bool

	// Clear removes all middleware from the registry.
	Clear()

	// Count returns the number of registered middleware.
	Count() int
}

// Orchestrator manages the composition and execution of middleware chains.
// Responsible for creating, configuring, and managing middleware chains
// based on application requirements and configuration.
type Orchestrator interface {
	// CreateChain creates a new middleware chain with the specified type.
	// ChainType determines which middleware are included and their order.
	CreateChain(chainType core.ChainType) (core.Chain, error)

	// GetChain retrieves a pre-configured chain by name.
	// Returns nil if chain is not found.
	GetChain(name string) (core.Chain, bool)

	// RegisterChain registers a named chain for later retrieval.
	// Returns an error if chain with the same name already exists.
	RegisterChain(name string, chain core.Chain) error

	// ListChains returns all registered chain names.
	ListChains() []string

	// RemoveChain removes a chain by name.
	// Returns true if chain was found and removed.
	RemoveChain(name string) bool
}

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
		return UnknownChainType
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

// ErrorCode represents common middleware error codes.
const (
	// ErrorCodeValidation indicates a validation error.
	ErrorCodeValidation = "VALIDATION_ERROR"

	// ErrorCodeAuthentication indicates an authentication error.
	ErrorCodeAuthentication = "AUTHENTICATION_ERROR"

	// ErrorCodeAuthorization indicates an authorization error.
	ErrorCodeAuthorization = "AUTHORIZATION_ERROR"

	// ErrorCodeRateLimit indicates a rate limiting error.
	ErrorCodeRateLimit = "RATE_LIMIT_ERROR"

	// ErrorCodeTimeout indicates a timeout error.
	ErrorCodeTimeout = "TIMEOUT_ERROR"

	// ErrorCodeInternal indicates an internal middleware error.
	ErrorCodeInternal = "INTERNAL_ERROR"

	// UnknownChainType represents an unknown chain type.
	UnknownChainType = "unknown"
)
