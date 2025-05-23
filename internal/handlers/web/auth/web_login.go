package auth

import (
	"net/http"

	"github.com/goformx/goforms/internal/handlers"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// WebLoginHandler handles the login page routes
type WebLoginHandler struct {
	base     handlers.Base
	logger   logging.Logger
	renderer *view.Renderer
}

// NewWebLoginHandler creates a new WebLoginHandler
func NewWebLoginHandler(logger logging.Logger, renderer *view.Renderer) *WebLoginHandler {
	return &WebLoginHandler{
		base: handlers.Base{
			Logger: logger,
		},
		logger:   logger,
		renderer: renderer,
	}
}

// Register sets up the routes for the web login handler
func (h *WebLoginHandler) Register(e *echo.Echo) {
	h.base.RegisterRoute(e, "GET", "/login", h.handleLogin)
}

// handleLogin renders the login page
func (h *WebLoginHandler) handleLogin(c echo.Context) error {
	h.base.Logger.Debug("handling login page request")

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok || csrfToken == "" {
		h.base.Logger.Error("CSRF token not found in context")
		return echo.NewHTTPError(http.StatusInternalServerError, "CSRF token not found")
	}

	data := shared.PageData{
		Title:     "Login - GoFormX",
		CSRFToken: csrfToken,
	}

	return h.renderer.Render(c, pages.Login(data))
}
