package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/goformx/goforms/internal/infrastructure/compose"
)

func handleDoctor(ctx context.Context, svc compose.Service, _ compose.Logger, _ []string) {
	issues := 0

	fmt.Println("🔍 Checking system health...")
	fmt.Println()

	// Check Docker daemon connectivity
	fmt.Print("Checking Docker daemon connectivity... ")
	if err := checkDockerDaemon(ctx); err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		issues++
	} else {
		fmt.Println("✅ OK")
	}

	// Check compose files
	fmt.Print("Checking compose files... ")
	composeFiles := []string{"docker-compose.yml", "docker-compose.prod.yml"}
	missingFiles := []string{}
	for _, file := range composeFiles {
		if _, statErr := os.Stat(file); os.IsNotExist(statErr) {
			missingFiles = append(missingFiles, file)
		}
	}
	if len(missingFiles) > 0 {
		fmt.Println("⚠️  WARNING")
		for _, file := range missingFiles {
			fmt.Printf("  Missing: %s\n", file)
		}
	} else {
		fmt.Println("✅ OK")
	}

	// Try to parse compose files
	fmt.Print("Validating compose file syntax... ")
	projectCtx := compose.ProjectContext{
		Name:         "goforms-dev",
		ComposeFiles: []string{"docker-compose.yml"},
		EnvFile:      ".env",
	}
	_, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		issues++
	} else {
		fmt.Println("✅ OK")
	}

	// Check environment file
	fmt.Print("Checking environment file... ")
	if _, statErr := os.Stat(".env"); os.IsNotExist(statErr) {
		fmt.Println("⚠️  WARNING")
		fmt.Println("  .env file not found (may be optional)")
	} else {
		fmt.Println("✅ OK")
	}

	fmt.Println()
	if issues == 0 {
		fmt.Println("✅ All checks passed!")
		os.Exit(0)
	} else {
		fmt.Printf("❌ Found %d issue(s)\n", issues)
		os.Exit(1)
	}
}

// checkDockerDaemon checks Docker daemon connectivity.
func checkDockerDaemon(ctx context.Context) error {
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		return err
	}

	err = dockerCLI.Initialize(&flags.ClientOptions{})
	if err != nil {
		return err
	}

	_, err = dockerCLI.Client().Ping(ctx)
	if err != nil {
		return fmt.Errorf("cannot connect to Docker daemon: %w", err)
	}

	return nil
}
