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
	logger, logErr := logging.NewFactory().CreateLogger()
	if logErr != nil {
		return logErr
	}

	// Check database connection
	if pingErr := db.PingContext(ctx); pingErr != nil {
		return pingErr
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
	logger, logErr := logging.NewFactory().CreateLogger()
	if logErr != nil {
		return logErr
	}

	// Check database connection
	if pingErr := db.PingContext(ctx); pingErr != nil {
		return pingErr
	}

	logger.Info("Database connection: OK")
	return nil
}

func DropAllTables(c *cli.Context) error {
	// Create logger
	logger, logErr := logging.NewFactory().CreateLogger()
	if logErr != nil {
		return logErr
	}

	// Check if force flag is set
	if !c.Bool("force") {
		logger.Warn("WARNING: This will drop all database tables. This action cannot be undone.")
		logger.Info("Are you sure you want to continue? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, readErr := reader.ReadString('\n')
		if readErr != nil {
			return readErr
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
	rows, queryErr := db.QueryContext(ctx, tableQuery)
	if queryErr != nil {
		return queryErr
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if scanErr := rows.Scan(&tableName); scanErr != nil {
			return scanErr
		}
		tables = append(tables, tableName)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return rowsErr
	}

	if len(tables) == 0 {
		logger.Info("No tables found in the database")
		return nil
	}

	// Disable foreign key checks temporarily
	if _, fkErr := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); fkErr != nil {
		return fkErr
	}

	// Drop all tables
	for _, table := range tables {
		dropQuery := "DROP TABLE IF EXISTS " + table
		if _, dropErr := db.ExecContext(ctx, dropQuery); dropErr != nil {
			return dropErr
		}
		logger.Info("Dropped table: " + table)
	}

	// Re-enable foreign key checks
	if _, fkErr := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); fkErr != nil {
		return fkErr
	}

	logger.Info("Successfully dropped all tables", logging.IntField("count", len(tables)))
	return nil
}
