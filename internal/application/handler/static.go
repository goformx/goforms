package handler

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// StaticHandler handles serving static files
type StaticHandler struct {
	logger logging.Logger
	config *config.Config
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(logger logging.Logger, cfg *config.Config) *StaticHandler {
	return &StaticHandler{
		logger: logger,
		config: cfg,
	}
}

// Register sets up routes for static file serving
func (h *StaticHandler) Register(e *echo.Echo) {
	// Handle Chrome DevTools well-known route only in development
	if h.config.App.IsDevelopment() {
		e.GET("/.well-known/appspecific/com.chrome.devtools.json", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"devtoolsFrontendUrl":  "",
				"faviconUrl":           "/favicon.ico",
				"id":                   "goforms",
				"title":                "GoFormX",
				"type":                 "node",
				"url":                  "/",
				"webSocketDebuggerUrl": "",
			})
		})
	}

	// Serve static files using Echo's built-in middleware
	distDir := h.config.Static.DistDir
	if stat, err := os.Stat(distDir); err == nil && stat.IsDir() {
		h.logger.Info("serving static files from dist directory",
			logging.StringField("dir", distDir),
		)
		e.Static("/", distDir)
	} else {
		h.logger.Warn("static directory not found or inaccessible",
			logging.StringField("dir", distDir),
			logging.ErrorField("error", err),
		)
	}
}
