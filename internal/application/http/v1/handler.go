package v1

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
)

// Handler handles web page requests
type Handler struct {
	logger  logging.Logger
	service contact.Service
}

// NewHandler creates a new web handler
func NewHandler(logger logging.Logger, service contact.Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// Register registers the web routes
func (h *Handler) Register(e *echo.Echo) {
	e.GET("/", h.handleHome)
	e.GET("/contact", h.handleContact)
}

// handleHome renders the home page
func (h *Handler) handleHome(c echo.Context) error {
	return pages.Home().Render(c.Request().Context(), c.Response().Writer)
}

// handleContact renders the contact page
func (h *Handler) handleContact(c echo.Context) error {
	return pages.Contact().Render(c.Request().Context(), c.Response().Writer)
}
