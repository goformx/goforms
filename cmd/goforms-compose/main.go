package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/goformx/goforms/internal/infrastructure/compose"
)

const (
	appName    = "goforms-compose"
	appVersion = "0.1.0"
)

// Build metadata - populated via -ldflags at build time
var (
	buildCommit = "unknown"
	buildDate   = "unknown"
	buildGoVer  = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	ctx := context.Background()

	// Create a simple logger for the CLI
	logger := compose.NewSimpleLogger(os.Stdout, os.Stderr, os.Stderr, os.Stdout)

	// Create compose service
	composeService, err := compose.NewService(logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create compose service: %v\n", err)
		os.Exit(1)
	}

	// Route to appropriate command
	switch command {
	case "dev":
		handleDev(ctx, composeService, logger, args)
	case "prod":
		handleProd(ctx, composeService, logger, args)
	case "doctor":
		handleDoctor(ctx, composeService, logger, args)
	case "version":
		handleVersion()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s <command> [options]

Commands:
  dev <subcommand>     Manage development environment
  prod <subcommand>    Manage production environment
  doctor               Check system health and configuration
  version              Show version information

Examples:
  %s dev up
  %s prod deploy --tag v1.0.0
  %s doctor

Run '%s <command> --help' for command-specific help.
`, appName, appName, appName, appName, appName)
}

func handleDev(ctx context.Context, svc compose.Service, logger compose.Logger, args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: dev command requires a subcommand (up, down, build, status, logs, health)\n")
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	// Parse common flags
	fs := flag.NewFlagSet("dev", flag.ExitOnError)
	projectName := fs.String("project-name", "goforms-dev", "Compose project name")
	composeFilesStr := fs.String("compose-file", "docker-compose.yml", "Compose file path (comma-separated for multiple files)")
	envFile := fs.String("env-file", ".env", "Environment file path")
	projectDir := fs.String("project-dir", "", "Project directory (defaults to compose file directory)")
	dryRun := fs.Bool("dry-run", false, "Perform a dry run without making changes")

	if err := fs.Parse(subArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Support multiple compose files (comma-separated)
	composeFiles := strings.Split(*composeFilesStr, ",")
	for i, file := range composeFiles {
		composeFiles[i] = strings.TrimSpace(file)
	}

	projectCtx := compose.ProjectContext{
		Name:         *projectName,
		ComposeFiles: composeFiles,
		EnvFile:      *envFile,
		ProjectDir:   *projectDir,
	}

	switch subcommand {
	case "up":
		handleDevUp(ctx, svc, logger, projectCtx, *dryRun)
	case "down":
		handleDevDown(ctx, svc, logger, projectCtx)
	case "build":
		handleDevBuild(ctx, svc, logger, projectCtx, fs.Args())
	case "status":
		handleDevStatus(ctx, svc, logger, projectCtx)
	case "logs":
		handleDevLogs(ctx, svc, logger, projectCtx, fs.Args())
	case "health":
		handleDevHealth(ctx, svc, logger, projectCtx)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown dev subcommand '%s'\n", subcommand)
		fmt.Fprintf(os.Stderr, "Valid subcommands: up, down, build, status, logs, health\n")
		os.Exit(1)
	}
}

func handleProd(ctx context.Context, svc compose.Service, logger compose.Logger, args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: prod command requires a subcommand (deploy, rollback, status, logs, health)\n")
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	// Parse common flags
	fs := flag.NewFlagSet("prod", flag.ExitOnError)
	projectName := fs.String("project-name", "goforms", "Compose project name")
	composeFilesStr := fs.String("compose-file", "docker-compose.prod.yml", "Compose file path (comma-separated for multiple files)")
	envFile := fs.String("env-file", ".env", "Environment file path")
	projectDir := fs.String("project-dir", "", "Project directory (defaults to compose file directory)")
	dryRun := fs.Bool("dry-run", false, "Perform a dry run without making changes")
	tag := fs.String("tag", "", "Image tag for deployment")
	pull := fs.Bool("pull", true, "Pull images before starting")

	if err := fs.Parse(subArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Support multiple compose files (comma-separated)
	composeFiles := strings.Split(*composeFilesStr, ",")
	for i, file := range composeFiles {
		composeFiles[i] = strings.TrimSpace(file)
	}

	projectCtx := compose.ProjectContext{
		Name:         *projectName,
		ComposeFiles: composeFiles,
		EnvFile:      *envFile,
		ProjectDir:   *projectDir,
	}

	switch subcommand {
	case "deploy":
		handleProdDeploy(ctx, svc, logger, projectCtx, *tag, *pull, *dryRun)
	case "rollback":
		handleProdRollback(ctx, svc, logger, projectCtx, *dryRun)
	case "status":
		handleProdStatus(ctx, svc, logger, projectCtx)
	case "logs":
		handleProdLogs(ctx, svc, logger, projectCtx, fs.Args())
	case "health":
		handleProdHealth(ctx, svc, logger, projectCtx)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown prod subcommand '%s'\n", subcommand)
		fmt.Fprintf(os.Stderr, "Valid subcommands: deploy, rollback, status, logs, health\n")
		os.Exit(1)
	}
}

func handleVersion() {
	fmt.Printf("%s version %s\n", appName, appVersion)
	if buildCommit != "unknown" || buildDate != "unknown" || buildGoVer != "unknown" {
		fmt.Printf("Build commit: %s\n", buildCommit)
		fmt.Printf("Build date: %s\n", buildDate)
		fmt.Printf("Go version: %s\n", buildGoVer)
	}
}
