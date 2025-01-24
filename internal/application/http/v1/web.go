package v1

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// WebHandler handles web page requests
type WebHandler struct {
	renderer *view.Renderer
	log      logging.Logger
}

// NewWebHandler creates a new web handler
func NewWebHandler(renderer *view.Renderer, log logging.Logger) *WebHandler {
	return &WebHandler{
		renderer: renderer,
		log:      log,
	}
}

func (h *WebHandler) wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Home handles the home page request
func (h *WebHandler) Home(c echo.Context) error {
	if err := h.renderer.Render(c, pages.Home()); err != nil {
		h.log.Error("failed to render home page", logging.Error(err))
		return h.wrapError(err, "failed to render home page")
	}
	return nil
}

// Contact handles the contact page request
func (h *WebHandler) Contact(c echo.Context) error {
	if err := h.renderer.Render(c, pages.Contact()); err != nil {
		h.log.Error("failed to render contact page", logging.Error(err))
		return h.wrapError(err, "failed to render contact page")
	}
	return nil
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	// Register routes
	e.GET("/", h.Home)
	e.GET("/contact", h.Contact)

	// Configure static file serving with proper caching and security
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")

	// Add cache control headers for static files
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "static",
		Browse: false,
		HTML5:  true,
		Index:  "index.html",
		Skipper: func(c echo.Context) bool {
			return !strings.HasPrefix(c.Request().URL.Path, "/static")
		},
	}))
}
