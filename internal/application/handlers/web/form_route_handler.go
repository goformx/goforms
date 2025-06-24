// internal/application/handlers/web/form_route_handler.go
package web

import (
	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/labstack/echo/v4"
)

// FormRouteHandler handles route registration and routing logic
type FormRouteHandler struct {
	handler *FormWebHandler
}

// NewFormRouteHandler creates a new FormRouteHandler
func NewFormRouteHandler(handler *FormWebHandler) *FormRouteHandler {
	return &FormRouteHandler{
		handler: handler,
	}
}

// RegisterRoutes registers all form-related routes with middleware
func (r *FormRouteHandler) RegisterRoutes(e *echo.Echo, accessManager *access.AccessManager) {
	forms := e.Group(constants.PathForms)
	forms.Use(access.Middleware(accessManager, r.handler.Logger))

	// Form management routes
	forms.GET("/new", r.handleNew)
	forms.POST("", r.handleCreate)
	forms.GET("/:id/edit", r.handleEdit)
	forms.POST("/:id/edit", r.handleUpdate)
	forms.DELETE("/:id", r.handleDelete)
	forms.GET("/:id/submissions", r.handleSubmissions)
}

// Route handler methods - these delegate to the appropriate handlers
func (r *FormRouteHandler) handleNew(c echo.Context) error {
	return r.handler.handleNew(c)
}

func (r *FormRouteHandler) handleCreate(c echo.Context) error {
	return r.handler.handleCreate(c)
}

func (r *FormRouteHandler) handleEdit(c echo.Context) error {
	return r.handler.handleEdit(c)
}

func (r *FormRouteHandler) handleUpdate(c echo.Context) error {
	return r.handler.handleUpdate(c)
}

func (r *FormRouteHandler) handleDelete(c echo.Context) error {
	return r.handler.handleDelete(c)
}

func (r *FormRouteHandler) handleSubmissions(c echo.Context) error {
	return r.handler.handleSubmissions(c)
}
