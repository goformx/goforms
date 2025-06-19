package web

import (
	"context"
	"net/http"

	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/labstack/echo/v4"
)

const (
	// StatusFound is the HTTP status code for redirects
	StatusFound = http.StatusFound // 302
)

// WebHandler handles web page requests
type WebHandler struct {
	*BaseHandler
}

// NewWebHandler creates a new web handler using BaseHandler
func NewWebHandler(base *BaseHandler) (*WebHandler, error) {
	return &WebHandler{BaseHandler: base}, nil
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	e.GET("/", h.handleHome)
	e.GET("/demo", h.handleDemo)
}

// handleHome handles the home page request
func (h *WebHandler) handleHome(c echo.Context) error {
	data := h.BuildPageData(c, "Home")
	if h.Logger != nil {
		h.Logger.Debug("handleHome: data.User", "user", data.User)
	}

	// Check if user is authenticated and redirect to dashboard
	user, err := h.RequireAuthenticatedUser(c)
	if err == nil && user != nil {
		return c.Redirect(StatusFound, "/dashboard")
	}

	// User is not authenticated, render home page
	if err := h.Renderer.Render(c, pages.Home(data)); err != nil {
		return h.HandleError(c, err, "Failed to render home page")
	}
	return nil
}

// handleDemo handles the demo page request
func (h *WebHandler) handleDemo(c echo.Context) error {
	data := h.BuildPageData(c, "Demo")
	if h.Logger != nil {
		h.Logger.Debug("handleDemo: data.User", "user", data.User)
	}

	// Check if user is authenticated and redirect to dashboard
	user, err := h.RequireAuthenticatedUser(c)
	if err == nil && user != nil {
		return c.Redirect(StatusFound, "/dashboard")
	}

	// User is not authenticated, render demo page
	return h.Renderer.Render(c, pages.Demo(data))
}

// Start initializes the web handler.
// This is called during application startup.
func (h *WebHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the web handler.
// This is called during application shutdown.
func (h *WebHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}
