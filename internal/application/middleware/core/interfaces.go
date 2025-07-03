// Package core provides the core interfaces and types for middleware functionality.
// This package contains only interfaces and types with no external dependencies,
// enabling clean architecture and avoiding import cycles.
package core

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
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

// Request represents an HTTP request abstraction that is framework-agnostic.
// This interface provides access to request data without depending on specific
// HTTP framework implementations like Echo.
type Request interface {
	// Method returns the HTTP method
	Method() string

	// URL returns the request URL
	URL() *http.Request

	// Path returns the request path
	Path() string

	// Query returns the query parameters
	Query() map[string][]string

	// Headers returns the request headers
	Headers() http.Header

	// Body returns the request body as an io.Reader
	Body() io.Reader

	// ContentLength returns the length of the request body
	ContentLength() int64

	// ContentType returns the Content-Type header value
	ContentType() string

	// RemoteAddr returns the client's network address
	RemoteAddr() string

	// UserAgent returns the User-Agent header value
	UserAgent() string

	// Referer returns the Referer header value
	Referer() string

	// Host returns the Host header value
	Host() string

	// IsSecure returns true if the request was made over HTTPS
	IsSecure() bool

	// Context returns the request context
	Context() context.Context

	// WithContext returns a new request with the given context
	WithContext(ctx context.Context) Request

	// Get retrieves a value from the request context
	Get(key string) any

	// Set stores a value in the request context
	Set(key string, value any)

	// Param returns a path parameter by name
	Param(name string) string

	// Params returns all path parameters
	Params() map[string]string

	// Cookie returns a cookie by name
	Cookie(name string) (*http.Cookie, error)

	// Cookies returns all cookies
	Cookies() []*http.Cookie

	// FormValue returns a form value by name
	FormValue(name string) string

	// Form returns the parsed form data
	Form() (url.Values, error)

	// MultipartForm returns the parsed multipart form data
	MultipartForm() (*multipart.Form, error)

	// IsAJAX returns true if the request is an AJAX request
	IsAJAX() bool

	// IsWebSocket returns true if the request is a WebSocket upgrade request
	IsWebSocket() bool

	// IsJSON returns true if the request expects JSON response
	IsJSON() bool

	// IsXML returns true if the request expects XML response
	IsXML() bool

	// Accepts returns true if the request accepts the given content type
	Accepts(contentType string) bool

	// AcceptsEncoding returns true if the request accepts the given encoding
	AcceptsEncoding(encoding string) bool

	// AcceptsLanguage returns true if the request accepts the given language
	AcceptsLanguage(language string) bool

	// RealIP returns the real IP address of the client
	RealIP() string

	// ForwardedFor returns the X-Forwarded-For header value
	ForwardedFor() string

	// RequestID returns the request ID if present
	RequestID() string

	// Timestamp returns when the request was received
	Timestamp() time.Time
}

// Response represents an HTTP response abstraction that is framework-agnostic.
// This interface provides access to response data without depending on specific
// HTTP framework implementations like Echo.
type Response interface {
	// StatusCode returns the HTTP status code
	StatusCode() int

	// SetStatusCode sets the HTTP status code
	SetStatusCode(code int) Response

	// Headers returns the response headers
	Headers() http.Header

	// SetHeader sets a response header
	SetHeader(key, value string) Response

	// AddHeader adds a response header (doesn't overwrite existing)
	AddHeader(key, value string) Response

	// Body returns the response body as an io.Reader
	Body() io.Reader

	// SetBody sets the response body
	SetBody(body io.Reader) Response

	// BodyBytes returns the response body as bytes
	BodyBytes() []byte

	// SetBodyBytes sets the response body from bytes
	SetBodyBytes(body []byte) Response

	// ContentType returns the Content-Type header value
	ContentType() string

	// SetContentType sets the Content-Type header
	SetContentType(contentType string) Response

	// ContentLength returns the length of the response body
	ContentLength() int64

	// SetContentLength sets the Content-Length header
	SetContentLength(length int64) Response

	// Location returns the Location header value (for redirects)
	Location() string

	// SetLocation sets the Location header (for redirects)
	SetLocation(location string) Response

	// SetCookie adds a cookie to the response
	SetCookie(cookie *http.Cookie) Response

	// Cookies returns all cookies that will be set
	Cookies() []*http.Cookie

	// Error returns the error associated with this response
	Error() error

	// SetError sets the error for this response
	SetError(err error) Response

	// IsError returns true if this response represents an error
	IsError() bool

	// IsRedirect returns true if this response is a redirect
	IsRedirect() bool

	// IsJSON returns true if the response content type is JSON
	IsJSON() bool

	// IsXML returns true if the response content type is XML
	IsXML() bool

	// IsHTML returns true if the response content type is HTML
	IsHTML() bool

	// IsText returns true if the response content type is plain text
	IsText() bool

	// IsBinary returns true if the response content type is binary
	IsBinary() bool

	// Context returns the response context
	Context() context.Context

	// WithContext returns a new response with the given context
	WithContext(ctx context.Context) Response

	// Get retrieves a value from the response context
	Get(key string) any

	// Set stores a value in the response context
	Set(key string, value any)

	// Timestamp returns when the response was created
	Timestamp() time.Time

	// SetTimestamp sets the response timestamp
	SetTimestamp(timestamp time.Time) Response

	// RequestID returns the request ID associated with this response
	RequestID() string

	// SetRequestID sets the request ID for this response
	SetRequestID(id string) Response

	// WriteTo writes the response to the given io.Writer
	WriteTo(io.Writer) (int64, error)

	// Clone creates a copy of this response
	Clone() Response
}

// Chain represents a sequence of middleware that can be executed in order.
// The chain manages the execution flow and ensures proper middleware ordering.
type Chain interface {
	// Process executes the middleware chain for a given request.
	// Returns the final response after all middleware have been processed.
	Process(ctx context.Context, req Request) Response

	// Add appends middleware to the end of the chain.
	// Returns the chain for method chaining.
	Add(middleware ...Middleware) Chain

	// Insert adds middleware at a specific position in the chain.
	// Position 0 inserts at the beginning, position -1 inserts at the end.
	Insert(position int, middleware ...Middleware) Chain

	// Remove removes middleware by name from the chain.
	// Returns true if middleware was found and removed.
	Remove(name string) bool

	// Get returns middleware by name from the chain.
	// Returns nil if middleware is not found.
	Get(name string) Middleware

	// List returns all middleware in the chain in execution order.
	List() []Middleware

	// Clear removes all middleware from the chain.
	Clear() Chain

	// Length returns the number of middleware in the chain.
	Length() int
}

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

// Logger is a minimal logging interface for orchestration events.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Orchestrator manages the composition and execution of middleware chains.
// Responsible for creating, configuring, and managing middleware chains
// based on application requirements and configuration.
type Orchestrator interface {
	// CreateChain creates a new middleware chain with the specified type.
	// ChainType determines which middleware are included and their order.
	CreateChain(chainType ChainType) (Chain, error)

	// BuildChain is an alias for CreateChain for backward compatibility.
	BuildChain(chainType ChainType) (Chain, error)

	// BuildChainForPath creates a middleware chain for a specific path and chain type.
	BuildChainForPath(chainType ChainType, requestPath string) (Chain, error)

	// GetChainForPath returns a cached chain for a path or builds a new one.
	GetChainForPath(chainType ChainType, requestPath string) (Chain, error)

	// GetChain retrieves a pre-configured chain by name.
	// Returns nil if chain is not found.
	GetChain(name string) (Chain, bool)

	// RegisterChain registers a named chain for later retrieval.
	// Returns an error if chain with the same name already exists.
	RegisterChain(name string, chain Chain) error

	// ListChains returns all registered chain names.
	ListChains() []string

	// RemoveChain removes a chain by name.
	// Returns true if chain was found and removed.
	RemoveChain(name string) bool

	// ClearCache clears the chain cache.
	ClearCache()

	// GetCacheStats returns cache statistics.
	GetCacheStats() map[string]any

	// GetChainPerformance returns performance metrics for chain building.
	GetChainPerformance() map[string]time.Duration

	// GetChainInfo returns information about a chain type.
	GetChainInfo(chainType ChainType) ChainInfo

	// ValidateConfiguration validates the current middleware configuration.
	ValidateConfiguration() error
}

// Response constructors for common use cases

// NewResponse creates a new response with the given status code
func NewResponse(statusCode int) Response {
	return &httpResponse{
		statusCode:    statusCode,
		headers:       make(http.Header),
		body:          nil,
		bodyBytes:     nil,
		contentType:   "",
		contentLength: 0,
		location:      "",
		cookies:       make([]*http.Cookie, 0),
		err:           nil,
		context:       context.Background(),
		timestamp:     time.Now(),
		requestID:     "",
		values:        make(map[string]any),
	}
}

// NewErrorResponse creates a new error response with the given status code and error
func NewErrorResponse(statusCode int, err error) Response {
	resp := NewResponse(statusCode)
	resp.SetError(err)

	return resp
}
