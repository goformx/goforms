package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/compose"
)

const stateFileName = ".compose-state.json"

// DeploymentState tracks deployment metadata for rollback.
type DeploymentState struct {
	LastTag      string    `json:"lastTag"`
	DeployedAt   time.Time `json:"deployedAt"`
	Services     []string  `json:"services"`
	ComposeFiles []string  `json:"composeFiles"`
	ProjectName  string    `json:"projectName"`
}

func handleProdDeploy(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext, tag string, pull bool, dryRun bool) {
	if tag == "" {
		fmt.Fprintf(os.Stderr, "Error: --tag is required for production deployment\n")
		os.Exit(1)
	}

	// Set IMAGE_TAG in environment if not already set
	if os.Getenv("IMAGE_TAG") == "" {
		os.Setenv("IMAGE_TAG", tag)
	}

	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	// Pull images if requested
	if pull && !dryRun {
		pullOptions := compose.PullOptions{
			Quiet:           false,
			IgnoreBuildable: false,
		}
		if err := svc.Pull(ctx, project, pullOptions); err != nil {
			fmt.Fprintf(os.Stderr, "Error pulling images: %v\n", err)
			os.Exit(1)
		}
	}

	// Start services
	upOptions := compose.UpOptions{
		Create: compose.CreateOptions{
			Recreate:      "diverged",
			RemoveOrphans: true,
			Quiet:         false,
		},
		Start: compose.StartOptions{
			Wait:        true,
			WaitTimeout: 120,
		},
		DryRun: dryRun,
	}

	if err := svc.Up(ctx, project, upOptions); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting services: %v\n", err)
		os.Exit(1)
	}

	// Save deployment state
	if !dryRun {
		if err := saveDeploymentState(projectCtx, project, tag); err != nil {
			logger.Warn(fmt.Sprintf("Failed to save deployment state: %v", err))
		}
	}

	fmt.Printf("Successfully deployed project '%s' with tag '%s'\n", project.Name, tag)
}

func handleProdRollback(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext, dryRun bool) {
	// Load previous deployment state
	state, err := loadDeploymentState(projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading deployment state: %v\n", err)
		fmt.Fprintf(os.Stderr, "Cannot rollback without previous deployment state\n")
		os.Exit(1)
	}

	fmt.Printf("Rolling back to tag: %s (deployed at: %s)\n", state.LastTag, state.DeployedAt.Format(time.RFC3339))

	// Validate that current project matches previous deployment
	if state.ProjectName != projectCtx.Name {
		logger.Warn(fmt.Sprintf("Project name mismatch: state has '%s', current is '%s'", state.ProjectName, projectCtx.Name))
	}

	// Set the previous tag
	os.Setenv("IMAGE_TAG", state.LastTag)

	// Load and deploy with previous tag
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	// Validate services match
	currentServices := make(map[string]bool)
	for name := range project.Services {
		currentServices[name] = true
	}
	for _, prevService := range state.Services {
		if !currentServices[prevService] {
			logger.Warn(fmt.Sprintf("Service '%s' from previous deployment not found in current compose file", prevService))
		}
	}

	upOptions := compose.UpOptions{
		Create: compose.CreateOptions{
			Recreate:      "diverged",
			RemoveOrphans: true,
			Quiet:         false,
		},
		Start: compose.StartOptions{
			Wait:        true,
			WaitTimeout: 120,
		},
		DryRun: dryRun,
	}

	if err := svc.Up(ctx, project, upOptions); err != nil {
		fmt.Fprintf(os.Stderr, "Error rolling back services: %v\n", err)
		os.Exit(1)
	}

	if !dryRun {
		// Update state to reflect rollback
		if err := saveDeploymentState(projectCtx, project, state.LastTag); err != nil {
			logger.Warn(fmt.Sprintf("Failed to update deployment state: %v", err))
		}
	}

	fmt.Printf("Successfully rolled back project '%s' to tag '%s'\n", project.Name, state.LastTag)
}

func handleProdStatus(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext) {
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	statuses, err := svc.Ps(ctx, project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		os.Exit(1)
	}

	// Load deployment state if available
	state, err := loadDeploymentState(projectCtx)
	if err == nil {
		fmt.Printf("Current deployment: tag=%s, deployed=%s\n\n", state.LastTag, state.DeployedAt.Format(time.RFC3339))
	}

	if len(statuses) == 0 {
		fmt.Println("No containers running")
		return
	}

	fmt.Printf("%-20s %-15s %-30s %-20s %-15s\n", "NAME", "STATE", "STATUS", "PORTS", "HEALTH")
	fmt.Println(strings.Repeat("-", 100))
	for _, status := range statuses {
		fmt.Printf("%-20s %-15s %-30s %-20s %-15s\n", status.Name, status.State, status.Status, status.Ports, status.Health)
	}
}

func handleProdLogs(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext, services []string) {
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	if err := svc.Logs(ctx, project, services, true, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting logs: %v\n", err)
		os.Exit(1)
	}
}

func handleProdHealth(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext) {
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	config := compose.HealthWaitConfig{
		Timeout:       120,
		RetryInterval: 3,
		Jitter:        true,
	}

	if err := svc.WaitForHealthy(ctx, project, nil, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error waiting for health: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All services are healthy")
}

// saveDeploymentState saves deployment metadata to a state file.
func saveDeploymentState(projectCtx compose.ProjectContext, project *compose.Project, tag string) error {
	projectDir := projectCtx.ProjectDir
	if projectDir == "" {
		if len(projectCtx.ComposeFiles) > 0 {
			projectDir = filepath.Dir(projectCtx.ComposeFiles[0])
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			projectDir = wd
		}
	}

	statePath := filepath.Join(projectDir, stateFileName)

	services := make([]string, 0, len(project.Services))
	for name := range project.Services {
		services = append(services, name)
	}

	state := DeploymentState{
		LastTag:      tag,
		DeployedAt:   time.Now(),
		Services:     services,
		ComposeFiles: projectCtx.ComposeFiles,
		ProjectName:  project.Name,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// loadDeploymentState loads deployment metadata from a state file.
func loadDeploymentState(projectCtx compose.ProjectContext) (*DeploymentState, error) {
	projectDir := projectCtx.ProjectDir
	if projectDir == "" {
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

	statePath := filepath.Join(projectDir, stateFileName)

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state DeploymentState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}
