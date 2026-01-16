package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/goformx/goforms/internal/infrastructure/compose"
)

func handleDevUp(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext, dryRun bool) {
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	options := compose.UpOptions{
		Create: compose.CreateOptions{
			Recreate:      "missing",
			RemoveOrphans: false,
			Quiet:         false,
		},
		Start: compose.StartOptions{
			Wait:        true,
			WaitTimeout: 60,
		},
		DryRun: dryRun,
	}

	if err := svc.Up(ctx, project, options); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting services: %v\n", err)
		os.Exit(1)
	}
}

func handleDevDown(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext) {
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	options := compose.DownOptions{
		RemoveVolumes: false,
		RemoveOrphans: false,
		Timeout:       10,
	}

	if err := svc.Down(ctx, project, options); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping services: %v\n", err)
		os.Exit(1)
	}
}

func handleDevBuild(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext, services []string) {
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	options := compose.BuildOptions{
		Pull:     false,
		NoCache:  false,
		Quiet:    false,
		Services: services,
		Deps:     true,
	}

	if err := svc.Build(ctx, project, options); err != nil {
		fmt.Fprintf(os.Stderr, "Error building images: %v\n", err)
		os.Exit(1)
	}
}

func handleDevStatus(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext) {
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

	if len(statuses) == 0 {
		fmt.Println("No containers running")
		return
	}

	fmt.Printf("%-20s %-15s %-30s %-20s\n", "NAME", "STATE", "STATUS", "PORTS")
	fmt.Println(strings.Repeat("-", 85))
	for _, status := range statuses {
		fmt.Printf("%-20s %-15s %-30s %-20s\n", status.Name, status.State, status.Status, status.Ports)
	}
}

func handleDevLogs(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext, services []string) {
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

func handleDevHealth(ctx context.Context, svc compose.Service, logger compose.Logger, projectCtx compose.ProjectContext) {
	project, err := svc.LoadProject(ctx, projectCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(1)
	}

	config := compose.HealthWaitConfig{
		Timeout:       60,
		RetryInterval: 2,
		Jitter:        true,
	}

	if err := svc.WaitForHealthy(ctx, project, nil, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error waiting for health: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All services are healthy")
}
