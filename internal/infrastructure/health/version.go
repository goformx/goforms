package health

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/version"
)

// VersionChecker implements the Checker interface for version health checks
type VersionChecker struct {
	logger logging.Logger
}

// NewVersionChecker creates a new version health checker
func NewVersionChecker(logger logging.Logger) *VersionChecker {
	return &VersionChecker{
		logger: logger,
	}
}

// Check performs a version health check
func (c *VersionChecker) Check(_ context.Context) error {
	info := version.GetInfo()

	// Check if version is set
	if info.Version == version.UnknownVersion {
		return errors.New("version not set")
	}

	// Check if build time is valid
	if info.BuildTime != version.UnknownVersion {
		if _, err := time.Parse(time.RFC3339, info.BuildTime); err != nil {
			return fmt.Errorf("invalid build time format: %w", err)
		}
	}

	// Check if git commit is set
	if info.GitCommit == version.UnknownVersion {
		return errors.New("git commit not set")
	}

	// Log version info
	c.logger.Info("version health check passed",
		"version", info.Version,
		"build_time", info.BuildTime,
		"git_commit", info.GitCommit,
		"go_version", info.GoVersion)

	return nil
}
