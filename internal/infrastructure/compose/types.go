// Package compose provides a wrapper around Docker Compose SDK for managing
// containerized applications defined in Compose files.
package compose

import (
	"context"
	"io"
)

// Logger provides a simple logging interface for compose operations.
// This can be adapted to use the existing infrastructure logger.
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Debug(msg string)
}

// ProjectContext captures all information needed to load and manage a Compose project.
type ProjectContext struct {
	// Name is the Compose project name (used for resource naming)
	Name string
	// ComposeFiles is a slice of Compose file paths (supports multi-file stacks)
	ComposeFiles []string
	// EnvFile is the path to the environment file (optional)
	EnvFile string
	// ProjectDir is the working directory for the project
	ProjectDir string
}

// HealthWaitConfig configures health check waiting behavior.
type HealthWaitConfig struct {
	// Timeout is the maximum time to wait for services to become healthy
	Timeout int // seconds
	// RetryInterval is the time between health check attempts
	RetryInterval int // seconds
	// Jitter enables random jitter to avoid thundering herd
	Jitter bool
}

// ServiceStatus represents the status of a service in a Compose project.
type ServiceStatus struct {
	Name    string
	State   string
	Status  string
	Ports   string
	Image   string
	Health  string
}

// Service provides methods for managing Compose projects.
type Service interface {
	// LoadProject loads a Compose project from the given context
	LoadProject(ctx context.Context, projectCtx ProjectContext) (*Project, error)

	// Up creates and starts services
	Up(ctx context.Context, project *Project, options UpOptions) error

	// Down stops and removes services
	Down(ctx context.Context, project *Project, options DownOptions) error

	// Pull pulls images for services
	Pull(ctx context.Context, project *Project, options PullOptions) error

	// Ps lists running containers
	Ps(ctx context.Context, project *Project) ([]ServiceStatus, error)

	// Logs writes logs for services to the given writer
	Logs(ctx context.Context, project *Project, services []string, follow bool, writer io.Writer) error

	// WaitForHealthy waits for services to become healthy
	WaitForHealthy(ctx context.Context, project *Project, services []string, config HealthWaitConfig) error
}

// Project represents a loaded Compose project.
// This wraps the SDK's project type to provide a stable interface.
type Project struct {
	Name     string
	Services map[string]ServiceConfig
	// Internal project from SDK - we'll keep this private
	internal any
}

// ServiceConfig represents a service configuration in a project.
type ServiceConfig struct {
	Name      string
	Image     string
	Build     *BuildConfig
	Ports     []string
	Environment map[string]string
	DependsOn []string
}

// BuildConfig represents build configuration for a service.
type BuildConfig struct {
	Context    string
	Dockerfile string
	Args       map[string]string
}

// UpOptions configures the Up operation.
type UpOptions struct {
	// Create options for container creation
	Create CreateOptions
	// Start options for starting containers
	Start StartOptions
	// DryRun performs validation without making changes
	DryRun bool
}

// CreateOptions configures container creation.
type CreateOptions struct {
	// Recreate forces recreation of containers
	Recreate string // "always", "never", "missing", "diverged"
	// RemoveOrphans removes containers for services not in compose file
	RemoveOrphans bool
	// Quiet suppresses progress output
	Quiet bool
}

// StartOptions configures container starting.
type StartOptions struct {
	// Wait waits for services to become healthy
	Wait bool
	// WaitTimeout is the timeout for waiting (seconds)
	WaitTimeout int
}

// DownOptions configures the Down operation.
type DownOptions struct {
	// RemoveVolumes removes volumes defined in the compose file
	RemoveVolumes bool
	// RemoveOrphans removes containers for services not in compose file
	RemoveOrphans bool
	// Timeout is the timeout for stopping containers (seconds)
	Timeout int
}

// PullOptions configures the Pull operation.
type PullOptions struct {
	// Quiet suppresses progress output
	Quiet bool
	// IgnoreBuildable skips services with build configuration
	IgnoreBuildable bool
}
