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

// WebHandler handles web page requests and template rendering
type WebHandler struct {
	renderer *view.Renderer
	logger   logging.Logger
}

// NewWebHandler creates a new web handler
//
// Dependencies:
//   - renderer: view.Renderer for template rendering
//   - logger: logging.Logger for structured logging
//
// The handler implements web page endpoints:
//   - GET / - Home page
//   - GET /contact - Contact page
//   - GET /subscribe - Subscription page
func NewWebHandler(renderer *view.Renderer, logger logging.Logger) *WebHandler {
	return &WebHandler{
		renderer: renderer,
		logger:   logger,
	}
}

// Register registers the web routes with the given Echo instance
func (h *WebHandler) Register(e *echo.Echo) {
	e.GET("/", h.Home)
	e.GET("/contact", h.Contact)
	e.GET("/subscribe", h.Subscribe)

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

// wrapError wraps an error with additional context
func (h *WebHandler) wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Home handles the home page request
// @Summary Render home page
// @Description Renders the home page template
// @Tags web
// @Produce html
// @Success 200 {string} string "HTML content"
// @Failure 500 {object} response.ErrorResponse
// @Router / [get]
func (h *WebHandler) Home(c echo.Context) error {
	if err := h.renderer.Render(c, pages.Home()); err != nil {
		h.logger.Error("failed to render home page", logging.Error(err))
		return h.wrapError(err, "failed to render home page")
	}
	return nil
}

// Contact handles the contact page request
// @Summary Render contact page
// @Description Renders the contact page template
// @Tags web
// @Produce html
// @Success 200 {string} string "HTML content"
// @Failure 500 {object} response.ErrorResponse
// @Router /contact [get]
func (h *WebHandler) Contact(c echo.Context) error {
	if err := h.renderer.Render(c, pages.Contact()); err != nil {
		h.logger.Error("failed to render contact page", logging.Error(err))
		return h.wrapError(err, "failed to render contact page")
	}
	return nil
}
