package http

import (
	"context"
	"net/url"
)

// Minimal Request interface stub for context reference
type Request interface{}

// Minimal Response interface stub for context reference
type Response interface{}

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
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Destroy() error
}

// Context is a framework-agnostic HTTP context abstraction
// for use in handlers and middleware.
type Context interface {
	// Request/Response access
	Request() Request
	Response() Response

	// Context propagation
	Context() context.Context
	WithContext(ctx context.Context) Context

	// Value storage (for middleware, etc.)
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)

	// User/session access
	User() *UserContext
	SetUser(user *UserContext)
	Session() Session
	SetSession(session Session)

	// Route/query/path params
	Param(name string) string
	QueryParam(name string) string
	FormValue(name string) string
	Form() (url.Values, error)
	PathParam(name string) string

	// Response methods
	JSON(code int, i interface{}) error
	JSONBlob(code int, b []byte) error
	String(code int, s string) error
	HTML(code int, html string) error
	NoContent(code int) error
	Redirect(code int, url string) error
	Error(err error) error
}
