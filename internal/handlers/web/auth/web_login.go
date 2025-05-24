package auth

import (
	"net/http"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
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
	return c.Redirect(http.StatusFound, "/dashboard")
}
