package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
)

const (
	// StatusFound is the HTTP status code for redirects
	StatusFound = http.StatusFound // 302
)

// PageHandler handles web page requests
type PageHandler struct {
	*BaseHandler
	AuthMiddleware *auth.Middleware
}

// NewPageHandler creates a new web handler using BaseHandler
func NewPageHandler(base *BaseHandler, authMiddleware *auth.Middleware) (*PageHandler, error) {
	if base == nil {
		return nil, errors.New("base handler cannot be nil")
	}

	if authMiddleware == nil {
		return nil, errors.New("auth middleware cannot be nil")
	}

	return &PageHandler{
		BaseHandler:    base,
		AuthMiddleware: authMiddleware,
	}, nil
}

// Register registers the web routes
func (h *PageHandler) Register(e *echo.Echo) {
	e.GET("/", h.handleHome)
	e.GET("/demo", h.handleDemo)
}

// handleHome handles the home page request
func (h *PageHandler) handleHome(c echo.Context) error {
	data := h.NewPageData(c, "Home")
	if h.Logger != nil {
		h.Logger.Debug("handleHome: data.User", "user", data.User)
	}

	// Check if user is authenticated and redirect to dashboard
	if mwcontext.IsAuthenticated(c) {
		return fmt.Errorf("redirect to dashboard: %w", c.Redirect(constants.StatusSeeOther, constants.PathDashboard))
	}

	// User is not authenticated, render home page
	if renderErr := h.Renderer.Render(c, pages.Home(*data)); renderErr != nil {
		return h.HandleError(c, renderErr, "Failed to render home page")
	}

	return nil
}

// handleDemo handles the demo page request
func (h *PageHandler) handleDemo(c echo.Context) error {
	data := h.NewPageData(c, "Demo")
	if h.Logger != nil {
		h.Logger.Debug("handleDemo: data.User", "user", data.User)
	}

	// Check if user is authenticated and redirect to dashboard
	if mwcontext.IsAuthenticated(c) {
		return fmt.Errorf("redirect to dashboard: %w", c.Redirect(constants.StatusSeeOther, constants.PathDashboard))
	}

	// User is not authenticated, render demo page
	return fmt.Errorf("render demo page: %w", h.Renderer.Render(c, pages.Demo(*data)))
}

// Start initializes the page handler.
// This is called during application startup.
func (h *PageHandler) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the page handler.
// This is called during application shutdown.
func (h *PageHandler) Stop(_ context.Context) error {
	return nil // No cleanup needed
}
