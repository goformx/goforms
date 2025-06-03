package auth

import (
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// WebLoginHandler handles web login requests
type WebLoginHandler struct {
	*handlers.BaseHandler
}

// NewWebLoginHandler creates a new web login handler
func NewWebLoginHandler(logger logging.Logger) *WebLoginHandler {
	return &WebLoginHandler{
		BaseHandler: handlers.NewBaseHandler(nil, nil, logger),
	}
}

// Register registers the web login routes
func (h *WebLoginHandler) Register(e *echo.Echo) {
	e.GET("/login", h.Login)
}

// Login handles the login page request
func (h *WebLoginHandler) Login(c echo.Context) error {
	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:     "Login - GoFormX",
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}

	// Render login page
	return pages.Login(data).Render(c.Request().Context(), c.Response().Writer)
}
