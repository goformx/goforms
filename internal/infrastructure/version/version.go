// Package version provides version information for the application.
// It is designed to be populated at build time using ldflags.
//
// Example usage:
//
//	go build -ldflags "-X github.com/goformx/goforms/internal/infrastructure/version.Version=1.0.0"
package version

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	GitCommit string `json:"gitCommit"`
	GoVersion string `json:"goVersion"`
}

const (
	// UnknownVersion represents an unknown version value
	UnknownVersion = "unknown"
)

//nolint:gochecknoglobals // Populated via -ldflags
var (
	// Version is the semantic version of the application
	Version = UnknownVersion
	// BuildTime is the time when the application was built
	BuildTime = UnknownVersion
	// GitCommit is the git commit hash of the build
	GitCommit = UnknownVersion
	// GoVersion is the version of Go used to build the application
	GoVersion = UnknownVersion
)

// GetInfo returns the version information as an Info struct
func GetInfo() Info {
	return Info{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
	}
}

// String returns a string representation of the version information
func (i Info) String() string {
	return fmt.Sprintf("Version: %s\nBuild Time: %s\nGit Commit: %s\nGo Version: %s",
		i.Version, i.BuildTime, i.GitCommit, i.GoVersion)
}

// IsDev returns true if the version is a development version
func (i Info) IsDev() bool {
	return i.Version == "dev"
}

// IsRelease returns true if the version is a release version
func (i Info) IsRelease() bool {
	return !i.IsDev() && !strings.HasPrefix(i.Version, "v")
}

// IsPreRelease returns true if the version is a pre-release version
func (i Info) IsPreRelease() bool {
	return !i.IsDev() && strings.HasPrefix(i.Version, "v")
}

// GetBuildTime returns the build time as a time.Time
func (i Info) GetBuildTime() (time.Time, error) {
	return time.Parse(time.RFC3339, i.BuildTime)
}

// Validate checks if the version information is valid
func (i Info) Validate() error {
	if i.Version == "" {
		return errors.New("version is required")
	}

	if i.BuildTime != UnknownVersion {
		if _, err := time.Parse(time.RFC3339, i.BuildTime); err != nil {
			return fmt.Errorf("invalid build time format: %w", err)
		}
	}

	if i.GitCommit != "" && len(i.GitCommit) < 7 {
		return errors.New("git commit hash must be at least 7 characters")
	}

	if i.GoVersion != "" && !strings.HasPrefix(i.GoVersion, "go") {
		return errors.New("go version must start with 'go'")
	}

	return nil
}

// Compare compares this version with another version
// Returns:
//
//	-1 if this version is less than the other version
//	 0 if this version is equal to the other version
//	 1 if this version is greater than the other version
func (i Info) Compare(other Info) int {
	return strings.Compare(i.Version, other.Version)
}
