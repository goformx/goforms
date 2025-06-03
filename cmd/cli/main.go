package main

import (
	"log"
	"os"

	"github.com/goformx/goforms/cmd/cli/commands"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	app := &cli.App{
		Name:  "goforms-cli",
		Usage: "CLI tool for managing GoFormX system",
		Commands: []*cli.Command{
			{
				Name:  "user",
				Usage: "User management commands",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a new user",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "email",
								Usage:    "User email",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "password",
								Usage:    "User password",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "first-name",
								Usage:    "User's first name",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "last-name",
								Usage:    "User's last name",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "role",
								Usage: "User role (default: user)",
								Value: "user",
							},
						},
						Action: commands.CreateUser,
					},
					{
						Name:   "list",
						Usage:  "List all users",
						Action: commands.ListUsers,
					},
					{
						Name:  "delete",
						Usage: "Delete a user",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:     "id",
								Usage:    "User ID to delete",
								Required: true,
							},
						},
						Action: commands.DeleteUser,
					},
				},
			},
			{
				Name:  "system",
				Usage: "System management commands",
				Subcommands: []*cli.Command{
					{
						Name:   "status",
						Usage:  "Check system status",
						Action: commands.CheckSystemStatus,
					},
					{
						Name:   "db-check",
						Usage:  "Check database connection",
						Action: commands.CheckDatabaseConnection,
					},
					{
						Name:   "drop-tables",
						Usage:  "Drop all database tables",
						Action: commands.DropAllTables,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "force",
								Usage: "Force drop without confirmation",
								Value: false,
							},
						},
					},
					{
						Name:   "routes",
						Usage:  "List all registered routes",
						Action: commands.ListRoutes,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
