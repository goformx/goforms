package commands

import (
	"fmt"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// getDB creates a new database connection
func getDB() (*database.Database, error) {
	// Create logger factory
	factory := logging.NewFactory()

	// Create logger
	logger, err := factory.CreateLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Load configuration
	cfg, err := config.New(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection
	db, err := database.NewDB(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
