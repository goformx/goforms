package handler

import (
	"net/http"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/labstack/echo/v4"
)

// VersionInfo contains build and version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	GitCommit string `json:"gitCommit"`
	GoVersion string `json:"goVersion"`
}

// VersionHandler handles version-related endpoints
type VersionHandler struct {
	*handlers.BaseHandler
	info VersionInfo
}

// NewVersionHandler creates a new version handler
func NewVersionHandler(info VersionInfo, logger logging.Logger) *VersionHandler {
	return &VersionHandler{
		BaseHandler: handlers.NewBaseHandler(nil, logger),
		info:        info,
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
