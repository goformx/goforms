package commands

import (
	"bufio"
	"context"
	"os"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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

	// Create logger
	logger, err := logging.NewFactory().CreateLogger()
	if err != nil {
		return err
	}

	// Check database connection
	if err := db.PingContext(ctx); err != nil {
		return err
	}

	logger.Info("System status: OK")
	logger.Info("Database connection: OK")
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

	// Create logger
	logger, err := logging.NewFactory().CreateLogger()
	if err != nil {
		return err
	}

	// Check database connection
	if err := db.PingContext(ctx); err != nil {
		return err
	}

	logger.Info("Database connection: OK")
	return nil
}

func DropAllTables(c *cli.Context) error {
	// Create logger
	logger, err := logging.NewFactory().CreateLogger()
	if err != nil {
		return err
	}

	// Check if force flag is set
	if !c.Bool("force") {
		logger.Warn("WARNING: This will drop all database tables. This action cannot be undone.")
		logger.Info("Are you sure you want to continue? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		if response[0] != 'y' && response[0] != 'Y' {
			logger.Info("Operation cancelled")
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
	tableQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`
	rows, err := db.QueryContext(ctx, tableQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(tables) == 0 {
		logger.Info("No tables found in the database")
		return nil
	}

	// Disable foreign key checks temporarily
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}

	// Drop all tables
	for _, table := range tables {
		dropQuery := "DROP TABLE IF EXISTS " + table
		if _, err := db.ExecContext(ctx, dropQuery); err != nil {
			return err
		}
		logger.Info("Dropped table: " + table)
	}

	// Re-enable foreign key checks
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return err
	}

	logger.Info("Successfully dropped all tables", logging.IntField("count", len(tables)))
	return nil
}
