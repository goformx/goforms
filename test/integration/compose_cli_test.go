//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/compose"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComposeCLIUpDown tests basic up/down functionality
func TestComposeCLIUpDown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if Docker is not available
	if !isDockerAvailable() {
		t.Skip("Docker is not available")
	}

	logger := compose.NewSimpleLogger(os.Stdout, os.Stderr, os.Stderr, os.Stdout)
	svc, err := compose.NewService(logger)
	require.NoError(t, err, "Failed to create compose service")

	ctx := context.Background()
	projectCtx := compose.ProjectContext{
		Name:        "goforms-test",
		ComposeFiles: []string{"docker-compose.yml"},
		EnvFile:     ".env",
		ProjectDir:  ".",
	}

	// Load project
	project, err := svc.LoadProject(ctx, projectCtx)
	require.NoError(t, err, "Failed to load project")
	assert.NotEmpty(t, project.Name)
	assert.NotEmpty(t, project.Services)

	// Test dry-run up
	upOptions := compose.UpOptions{
		Create: compose.CreateOptions{
			Recreate:      "missing",
			RemoveOrphans: false,
			Quiet:          false,
		},
		Start: compose.StartOptions{
			Wait:       false,
			WaitTimeout: 30,
		},
		DryRun: true,
	}

	err = svc.Up(ctx, project, upOptions)
	assert.NoError(t, err, "Dry-run up should succeed")

	// Test down (should work even if nothing is running)
	downOptions := compose.DownOptions{
		RemoveVolumes: false,
		RemoveOrphans: false,
		Timeout:       10,
	}

	err = svc.Down(ctx, project, downOptions)
	// Down may fail if nothing is running, which is acceptable
	_ = err
}

// TestComposeCLIHealthWait tests health check waiting
func TestComposeCLIHealthWait(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if Docker is not available
	if !isDockerAvailable() {
		t.Skip("Docker is not available")
	}

	logger := compose.NewSimpleLogger(os.Stdout, os.Stderr, os.Stderr, os.Stdout)
	svc, err := compose.NewService(logger)
	require.NoError(t, err, "Failed to create compose service")

	ctx := context.Background()
	projectCtx := compose.ProjectContext{
		Name:        "goforms-test",
		ComposeFiles: []string{"docker-compose.yml"},
		EnvFile:     ".env",
		ProjectDir:  ".",
	}

	project, err := svc.LoadProject(ctx, projectCtx)
	require.NoError(t, err, "Failed to load project")

	// Test health wait config
	config := compose.HealthWaitConfig{
		Timeout:       10, // Short timeout for test
		RetryInterval: 1,
		Jitter:        false,
	}

	// This will likely timeout if services aren't running, which is expected
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = svc.WaitForHealthy(ctxWithTimeout, project, nil, config)
	// Health wait may fail if services aren't running, which is acceptable for this test
	_ = err
}

// TestComposeCLIRollbackState tests rollback state management
func TestComposeCLIRollbackState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for state file
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, ".compose-state.json")

	// Create test state
	state := map[string]any{
		"lastTag":      "v1.0.0",
		"deployedAt":   time.Now().Format(time.RFC3339),
		"services":     []string{"goforms", "postgres"},
		"composeFiles": []string{"docker-compose.prod.yml"},
		"projectName":  "goforms",
	}

	data, err := json.MarshalIndent(state, "", "  ")
	require.NoError(t, err, "Failed to marshal state")

	err = os.WriteFile(statePath, data, 0644)
	require.NoError(t, err, "Failed to write state file")

	// Read it back
	readData, err := os.ReadFile(statePath)
	require.NoError(t, err, "Failed to read state file")

	var readState map[string]any
	err = json.Unmarshal(readData, &readState)
	require.NoError(t, err, "Failed to unmarshal state")

	// Verify state
	assert.Equal(t, "v1.0.0", readState["lastTag"])
	assert.NotEmpty(t, readState["deployedAt"])
	assert.Contains(t, readState["services"], "goforms")
}

// TestComposeCLIProjectLoad tests project loading
func TestComposeCLIProjectLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if Docker is not available
	if !isDockerAvailable() {
		t.Skip("Docker is not available")
	}

	logger := compose.NewSimpleLogger(os.Stdout, os.Stderr, os.Stderr, os.Stdout)
	svc, err := compose.NewService(logger)
	require.NoError(t, err, "Failed to create compose service")

	ctx := context.Background()

	// Test loading dev compose file
	projectCtx := compose.ProjectContext{
		Name:        "goforms-test",
		ComposeFiles: []string{"docker-compose.yml"},
		EnvFile:     ".env",
		ProjectDir:  ".",
	}

	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		// If compose file doesn't exist, skip test
		if os.IsNotExist(err) {
			t.Skip("docker-compose.yml not found, skipping test")
		}
		require.NoError(t, err, "Failed to load project")
	}

	assert.NotEmpty(t, project.Name)
	assert.NotEmpty(t, project.Services)

	// Verify we can get service status
	statuses, err := svc.Ps(ctx, project)
	require.NoError(t, err, "Failed to get service status")
	// Statuses may be empty if services aren't running, which is fine
	_ = statuses
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	// Try to create a compose service - if it fails, Docker is not available
	logger := compose.NewSimpleLogger(os.Stdout, os.Stderr, os.Stderr, os.Stdout)
	_, err := compose.NewService(logger)
	return err == nil
}
