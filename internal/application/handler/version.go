package handler

import (
	"github.com/goformx/goforms/internal/presentation/handlers"
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
}
