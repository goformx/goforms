package handler

import (
	"net/http"

	"errors"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// BaseOption defines a function that configures a Base handler
type BaseOption func(*Base)

// WithLogger sets the logger for the base handler
func WithLogger(logger logging.Logger) BaseOption {
	return func(b *Base) {
		b.logger = logger
	}
}

// Base provides common functionality for handlers
type Base struct {
	logger logging.Logger
}

// Logger returns the logger instance
func (b *Base) Logger() logging.Logger {
	return b.logger
}

// Validate ensures all required dependencies are properly set
func (b *Base) Validate() error {
	if b.logger == nil {
		return errors.New("logger is required")
	}
	return nil
}

// NewBase creates a new base handler
func NewBase(opts ...BaseOption) Base {
	b := Base{}
	for _, opt := range opts {
		opt(&b)
	}
	return b
}

// VersionInfo contains build and version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	GitCommit string `json:"gitCommit"`
	GoVersion string `json:"goVersion"`
}

// VersionHandler handles version-related endpoints
type VersionHandler struct {
	Base
	info VersionInfo
}

// NewVersionHandler creates a new version handler
func NewVersionHandler(info VersionInfo, logger logging.Logger) *VersionHandler {
	return &VersionHandler{
		Base: NewBase(WithLogger(logger)),
		info: info,
	}
}

// Register registers the version routes
func (h *VersionHandler) Register(e *echo.Echo) {
	e.GET("/v1/version", h.GetVersion)
}

// GetVersion returns the application version information
func (h *VersionHandler) GetVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, h.info)
}
