package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/goformx/goforms/internal/infrastructure/compose"
)

func handleDoctor(ctx context.Context, svc compose.Service, logger compose.Logger, _ []string) {
	issues := 0

	fmt.Println("🔍 Checking system health...")
	fmt.Println()

	// Check Docker daemon connectivity
	fmt.Print("Checking Docker daemon connectivity... ")
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		issues++
	} else {
		err = dockerCLI.Initialize(&flags.ClientOptions{})
		if err != nil {
			fmt.Println("❌ FAILED")
			fmt.Printf("  Error: %v\n", err)
			issues++
		} else {
			// Try to ping the daemon
			_, err = dockerCLI.Client().Ping(ctx)
			if err != nil {
				fmt.Println("❌ FAILED")
				fmt.Printf("  Error: Cannot connect to Docker daemon: %v\n", err)
				issues++
			} else {
				fmt.Println("✅ OK")
			}
		}
	}

	// Check compose files
	fmt.Print("Checking compose files... ")
	composeFiles := []string{"docker-compose.yml", "docker-compose.prod.yml"}
	missingFiles := []string{}
	for _, file := range composeFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
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
	_, err = svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		issues++
	} else {
		fmt.Println("✅ OK")
	}

	// Check environment file
	fmt.Print("Checking environment file... ")
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
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
