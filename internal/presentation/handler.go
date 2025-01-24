package web

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/ui/pages"
	"github.com/jonesrussell/goforms/internal/view"
)

type Handler struct {
	renderer *view.Renderer
	log      logger.Logger
}

func NewHandler(renderer *view.Renderer, log logger.Logger) *Handler {
	return &Handler{
		renderer: renderer,
		log:      log,
	}
}

func (h *Handler) wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func (h *Handler) Home(c echo.Context) error {
	if err := h.renderer.Render(c, pages.Home()); err != nil {
		h.log.Error("failed to render home page", logger.Error(err))
		return h.wrapError(err, "failed to render home page")
	}
	return nil
}

func (h *Handler) Contact(c echo.Context) error {
	if err := h.renderer.Render(c, pages.Contact()); err != nil {
		h.log.Error("failed to render contact page", logger.Error(err))
		return h.wrapError(err, "failed to render contact page")
	}
	return nil
}

func (h *Handler) Register(e *echo.Echo) {
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

	// Custom logger middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			h.log.Info("request",
				logger.String("uri", v.URI),
				logger.Int("status", v.Status),
				logger.String("method", c.Request().Method),
				logger.String("ip", c.RealIP()),
				logger.Duration("latency", v.Latency),
			)
			return nil
		},
	}))
}
