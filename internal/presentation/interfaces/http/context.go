package http

import (
	"context"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/core"
)

// Request interface from application layer
type Request = middleware.Request

// Response interface from application layer
type Response = core.Response

// UserContext represents authenticated user information
// (can be extended as needed)
type UserContext struct {
	ID    string
	Email string
	Role  string
}

// Session represents a user session abstraction
type Session interface {
	ID() string
	UserID() string
	Get(key string) (any, bool)
	Set(key string, value any)
	Delete(key string)
	Destroy() error
}

// Context is a framework-agnostic HTTP context abstraction
// for use in handlers and middleware.
// This interface matches the infrastructure Context interface to avoid conflicts.
type Context interface {
	// Request methods
	Method() string
	Path() string
	Param(name string) string
	QueryParam(name string) string
	FormValue(name string) string
	Body() []byte
	Headers() map[string]string

	// Response methods
	JSON(statusCode int, data interface{}) error
	JSONBlob(statusCode int, data []byte) error
	String(statusCode int, data string) error
	Redirect(statusCode int, url string) error
	NoContent(statusCode int) error

	// Context methods
	Get(key string) interface{}
	Set(key string, value interface{})

	// Context propagation (needed for application services)
	RequestContext() context.Context

	// Presentation methods
	RenderComponent(component interface{}) error

	// Request access (for bridge methods)
	GetUnderlyingContext() interface{}
}
