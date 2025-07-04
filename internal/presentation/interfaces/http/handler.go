package http

import "context"

// Handler defines the base interface for all HTTP handlers
// in the presentation layer.
type Handler interface {
	// Name returns a unique identifier for this handler
	Name() string

	// Routes returns the routes this handler manages
	Routes() []Route

	// Middleware returns any handler-level middleware (optional)
	Middleware() []Middleware

	// Lifecycle methods
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Route defines a single route with its handler method
type Route struct {
	Method      string
	Path        string
	Handler     HandlerMethod
	Middleware  []Middleware
	Name        string
	Description string
}

// HandlerMethod represents a single handler method
// (receives a framework-agnostic Context)
type HandlerMethod func(Context) error

// Middleware interface stub for handler-level and route-level middleware
type Middleware interface {
	Process(ctx context.Context, req Request, next HandlerMethod) Response
	Name() string
	Priority() int
}
