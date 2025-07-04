package handlers

import (
	"context"

	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// BaseHandler provides common functionality for all HTTP handlers
// in the presentation layer. Embed this in your handler structs.
type BaseHandler struct {
	name       string
	routes     []httpiface.Route
	middleware []httpiface.Middleware
}

// NewBaseHandler creates a new BaseHandler with the given name.
func NewBaseHandler(name string) *BaseHandler {
	return &BaseHandler{
		name: name,
	}
}

// Name returns the handler's unique name.
func (h *BaseHandler) Name() string {
	return h.name
}

// Routes returns all routes registered for this handler.
func (h *BaseHandler) Routes() []httpiface.Route {
	return h.routes
}

// Middleware returns all handler-level middleware.
func (h *BaseHandler) Middleware() []httpiface.Middleware {
	return h.middleware
}

// Start is a no-op by default. Override if needed.
func (h *BaseHandler) Start(_ context.Context) error {
	return nil
}

// Stop is a no-op by default. Override if needed.
func (h *BaseHandler) Stop(_ context.Context) error {
	return nil
}

// AddRoute adds a new route to the handler.
func (h *BaseHandler) AddRoute(route httpiface.Route) {
	h.routes = append(h.routes, route)
}

// AddMiddleware adds handler-level middleware.
func (h *BaseHandler) AddMiddleware(mw httpiface.Middleware) {
	h.middleware = append(h.middleware, mw)
}
