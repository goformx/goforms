package compose

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
	containertypes "github.com/docker/docker/api/types/container"
)

// service implements the Service interface using Docker Compose SDK.
type service struct {
	composeService api.Compose
	logger         Logger
	dockerCLI      command.Cli
}

// NewService creates a new Compose service instance.
func NewService(logger Logger) (Service, error) {
	if logger == nil {
		logger = &NullLogger{}
	}

	// Initialize Docker CLI
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker CLI: %w", err)
	}

	err = dockerCLI.Initialize(&flags.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker CLI: %w", err)
	}

	// Create Compose service with progress writer
	outputWriter := os.Stdout
	errorWriter := os.Stderr

	composeService, err := compose.NewComposeService(
		dockerCLI,
		compose.WithOutputStream(outputWriter),
		compose.WithErrorStream(errorWriter),
		compose.WithMaxConcurrency(4),
		compose.WithPrompt(compose.AlwaysOkPrompt()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose service: %w", err)
	}

	return &service{
		composeService: composeService,
		logger:         logger,
		dockerCLI:      dockerCLI,
	}, nil
}

// NewServiceWithOptions creates a new Compose service with custom options.
func NewServiceWithOptions(logger Logger, opts ...compose.Option) (Service, error) {
	if logger == nil {
		logger = &NullLogger{}
	}

	// Initialize Docker CLI
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker CLI: %w", err)
	}

	err = dockerCLI.Initialize(&flags.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker CLI: %w", err)
	}

	// Create Compose service with provided options
	composeService, err := compose.NewComposeService(dockerCLI, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose service: %w", err)
	}

	return &service{
		composeService: composeService,
		logger:         logger,
		dockerCLI:      dockerCLI,
	}, nil
}

// LoadProject loads a Compose project from the given context.
func (s *service) LoadProject(ctx context.Context, projectCtx ProjectContext) (*Project, error) {
	// Resolve project directory
	projectDir := projectCtx.ProjectDir
	if projectDir == "" {
		// Default to directory of first compose file
		if len(projectCtx.ComposeFiles) > 0 {
			projectDir = filepath.Dir(projectCtx.ComposeFiles[0])
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			projectDir = wd
		}
	}

	// Resolve compose file paths relative to project dir
	configPaths := make([]string, 0, len(projectCtx.ComposeFiles))
	for _, file := range projectCtx.ComposeFiles {
		if !filepath.IsAbs(file) {
			file = filepath.Join(projectDir, file)
		}
		configPaths = append(configPaths, file)
	}

	// Build load options
	loadOptions := api.ProjectLoadOptions{
		ConfigPaths: configPaths,
		ProjectName: projectCtx.Name,
		WorkingDir:  projectDir,
	}

	// Load environment file if provided
	if projectCtx.EnvFile != "" {
		envFile := projectCtx.EnvFile
		if !filepath.IsAbs(envFile) {
			envFile = filepath.Join(projectDir, envFile)
		}
		loadOptions.EnvFiles = []string{envFile}
	}

	// Load the project
	project, err := s.composeService.LoadProject(ctx, loadOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	// Convert to our Project type
	result := &Project{
		Name:     project.Name,
		Services: make(map[string]ServiceConfig),
		internal: project,
	}

	// Extract service configurations for dry-run and status display
	for name, svc := range project.Services {
		serviceConfig := ServiceConfig{
			Name:        name,
			Image:       svc.Image,
			Ports:       []string{},
			Environment: make(map[string]string),
			DependsOn:   make([]string, 0),
		}

		// Extract depends_on
		if svc.DependsOn != nil {
			for depName := range svc.DependsOn {
				serviceConfig.DependsOn = append(serviceConfig.DependsOn, depName)
			}
		}

		// Extract ports
		if svc.Ports != nil {
			for _, port := range svc.Ports {
				if port.Published != "" && port.Target != 0 {
					serviceConfig.Ports = append(serviceConfig.Ports,
						fmt.Sprintf("%s:%d/%s", port.Published, port.Target, port.Protocol))
				}
			}
		}

		// Extract environment variables
		if svc.Environment != nil {
			for envName, envValue := range svc.Environment {
				if envValue != nil {
					serviceConfig.Environment[envName] = *envValue
				}
			}
		}

		// Extract build config if present
		if svc.Build != nil {
			serviceConfig.Build = &BuildConfig{
				Context:    svc.Build.Context,
				Dockerfile: svc.Build.Dockerfile,
				Args:       make(map[string]string),
			}
			if svc.Build.Args != nil {
				for k, v := range svc.Build.Args {
					if v != nil {
						serviceConfig.Build.Args[k] = *v
					}
				}
			}
		}

		result.Services[name] = serviceConfig
	}

	s.logger.Info(fmt.Sprintf("Loaded project '%s' with %d services", result.Name, len(result.Services)))

	return result, nil
}

// Up creates and starts services.
func (s *service) Up(ctx context.Context, project *Project, options UpOptions) error {
	if options.DryRun {
		return s.dryRunUp(ctx, project, options)
	}

	internalProject := project.internal.(*types.Project)

	upOptions := api.UpOptions{
		Create: api.CreateOptions{
			Recreate:      options.Create.Recreate,
			RemoveOrphans: options.Create.RemoveOrphans,
			QuietPull:     options.Create.Quiet,
		},
		Start: api.StartOptions{
			Wait:        options.Start.Wait,
			WaitTimeout: time.Duration(options.Start.WaitTimeout) * time.Second,
		},
	}

	err := s.composeService.Up(ctx, internalProject, upOptions)
	if err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Successfully started project '%s'", project.Name))
	return nil
}

// dryRunUp performs a dry-run of the Up operation.
func (s *service) dryRunUp(ctx context.Context, project *Project, options UpOptions) error {
	s.logger.Info("DRY RUN: Would start the following services:")

	for name, svc := range project.Services {
		s.logger.Info(fmt.Sprintf("  Service: %s", name))
		if svc.Image != "" {
			s.logger.Info(fmt.Sprintf("    Image: %s", svc.Image))
		}
		if svc.Build != nil {
			s.logger.Info(fmt.Sprintf("    Build: %s (Dockerfile: %s)", svc.Build.Context, svc.Build.Dockerfile))
		}
		if len(svc.Ports) > 0 {
			s.logger.Info(fmt.Sprintf("    Ports: %v", svc.Ports))
		}
		if len(svc.DependsOn) > 0 {
			s.logger.Info(fmt.Sprintf("    Depends on: %v", svc.DependsOn))
		}
	}

	s.logger.Info("DRY RUN: No changes were made")
	return nil
}

// Down stops and removes services.
func (s *service) Down(ctx context.Context, project *Project, options DownOptions) error {
	internalProject := project.internal.(*types.Project)

	downOptions := api.DownOptions{
		Volumes:       options.RemoveVolumes,
		RemoveOrphans: options.RemoveOrphans,
		Project:       internalProject,
	}

	if options.Timeout > 0 {
		timeout := time.Duration(options.Timeout) * time.Second
		downOptions.Timeout = &timeout
	}

	err := s.composeService.Down(ctx, internalProject.Name, downOptions)
	if err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Successfully stopped project '%s'", project.Name))
	return nil
}

// Pull pulls images for services.
func (s *service) Pull(ctx context.Context, project *Project, options PullOptions) error {
	internalProject := project.internal.(*types.Project)

	pullOptions := api.PullOptions{
		Quiet:           options.Quiet,
		IgnoreBuildable: options.IgnoreBuildable,
	}

	err := s.composeService.Pull(ctx, internalProject, pullOptions)
	if err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Successfully pulled images for project '%s'", project.Name))
	return nil
}

// Build builds images for services.
func (s *service) Build(ctx context.Context, project *Project, options BuildOptions) error {
	internalProject := project.internal.(*types.Project)

	buildOptions := api.BuildOptions{
		Pull:     options.Pull,
		NoCache:  options.NoCache,
		Quiet:    options.Quiet,
		Services: options.Services,
		Deps:     options.Deps,
		Progress: "auto",
	}

	err := s.composeService.Build(ctx, internalProject, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build images: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Successfully built images for project '%s'", project.Name))
	return nil
}

// Ps lists running containers.
func (s *service) Ps(ctx context.Context, project *Project) ([]ServiceStatus, error) {
	internalProject := project.internal.(*types.Project)

	containers, err := s.composeService.Ps(ctx, internalProject.Name, api.PsOptions{
		Project: internalProject,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	statuses := make([]ServiceStatus, 0, len(containers))
	for _, container := range containers {
		portsStr := ""
		if len(container.Publishers) > 0 {
			portParts := make([]string, 0, len(container.Publishers))
			for _, pub := range container.Publishers {
				if pub.URL != "" {
					portParts = append(portParts, pub.URL)
				} else if pub.PublishedPort > 0 {
					portParts = append(portParts, fmt.Sprintf("%d:%d/%s", pub.PublishedPort, pub.TargetPort, pub.Protocol))
				}
			}
			if len(portParts) > 0 {
				portsStr = fmt.Sprintf("%v", portParts)
			}
		}
		statuses = append(statuses, ServiceStatus{
			Name:   container.Name,
			State:  string(container.State),
			Status: container.Status,
			Ports:  portsStr,
			Image:  container.Image,
			Health: container.Health,
		})
	}

	return statuses, nil
}

// Logs writes logs for services to the given writer.
// Uses Docker CLI directly since Compose SDK may not expose Logs method.
func (s *service) Logs(ctx context.Context, project *Project, services []string, follow bool, outputWriter io.Writer) error {
	internalProject := project.internal.(*types.Project)

	// Use Docker client to get logs for containers in the project
	containers, err := s.composeService.Ps(ctx, internalProject.Name, api.PsOptions{
		Project:  internalProject,
		Services: services,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// For each container, stream logs using Docker client
	for _, container := range containers {
		opts := containertypes.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     follow,
			Timestamps: false,
			Tail:       "all",
		}

		containerLogs, err := s.dockerCLI.Client().ContainerLogs(ctx, container.ID, opts)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to get logs for container %s: %v", container.Name, err))
			continue
		}
		defer containerLogs.Close()

		// Write container name header
		fmt.Fprintf(outputWriter, "\n=== %s ===\n", container.Name)
		_, err = io.Copy(outputWriter, containerLogs)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Error copying logs for container %s: %v", container.Name, err))
		}
	}

	return nil
}

// WaitForHealthy waits for services to become healthy.
func (s *service) WaitForHealthy(ctx context.Context, project *Project, services []string, config HealthWaitConfig) error {
	internalProject := project.internal.(*types.Project)

	timeout := time.Duration(config.Timeout) * time.Second
	retryInterval := time.Duration(config.RetryInterval) * time.Second

	// Create context with timeout
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Determine which services to wait for
	servicesToWait := services
	if len(servicesToWait) == 0 {
		// Wait for all services
		for name := range project.Services {
			servicesToWait = append(servicesToWait, name)
		}
	}

	// Wait for each service
	for _, serviceName := range servicesToWait {
		if err := s.waitForServiceHealthy(waitCtx, internalProject, serviceName, retryInterval, config.Jitter); err != nil {
			return fmt.Errorf("service '%s' failed health check: %w", serviceName, err)
		}
		s.logger.Info(fmt.Sprintf("Service '%s' is healthy", serviceName))
	}

	return nil
}

// waitForServiceHealthy waits for a single service to become healthy.
func (s *service) waitForServiceHealthy(ctx context.Context, project *types.Project, serviceName string, interval time.Duration, jitter bool) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Add jitter if enabled
			if jitter {
				jitterMs := rand.Intn(500) // 0-500ms jitter
				time.Sleep(time.Duration(jitterMs) * time.Millisecond)
			}

			// Check service health
			containers, err := s.composeService.Ps(ctx, project.Name, api.PsOptions{
				Project:  project,
				Services: []string{serviceName},
			})
			if err != nil {
				s.logger.Debug(fmt.Sprintf("Failed to check service status: %v", err))
				continue
			}

			// Find the container for this service
			for _, container := range containers {
				if container.Service == serviceName {
					if container.Health == "healthy" || (container.Health == "" && container.State == "running") {
						return nil
					}
					s.logger.Debug(fmt.Sprintf("Service '%s' not healthy yet: state=%s, health=%s", serviceName, container.State, container.Health))
				}
			}
		}
	}
}
