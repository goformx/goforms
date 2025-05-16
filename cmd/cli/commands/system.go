package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"

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

func DropAllTables(c *cli.Context) error {
	// Check if force flag is set
	if !c.Bool("force") {
		fmt.Println("WARNING: This will drop all database tables. This action cannot be undone.")
		fmt.Print("Are you sure you want to continue? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		if response[0] != 'y' && response[0] != 'Y' {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	ctx := context.Background()

	// Initialize database connection
	db, err := getDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Get list of all tables
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get list of tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %w", err)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in the database")
		return nil
	}

	// Disable foreign key checks temporarily
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Drop all tables
	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		fmt.Printf("Dropped table: %s\n", table)
	}

	// Re-enable foreign key checks
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	fmt.Printf("Successfully dropped %d tables\n", len(tables))
	return nil
}
