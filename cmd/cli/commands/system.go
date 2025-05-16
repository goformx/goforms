package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
)

func CheckSystemStatus(c *cli.Context) error {
	ctx := context.Background()

	// Initialize database connection
	db, err := getDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Check database connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	fmt.Println("System status: OK")
	fmt.Println("Database connection: OK")
	return nil
}

func CheckDatabaseConnection(c *cli.Context) error {
	ctx := context.Background()

	// Initialize database connection
	db, err := getDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Check database connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	fmt.Println("Database connection: OK")
	return nil
}
