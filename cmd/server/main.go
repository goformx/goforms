package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/app"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			// We can't use the logger here as we're shutting it down
			log.Printf("failed to sync logger: %v", err)
		}
	}()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Create and start application
	application := app.New(cfg, logger, db)

	// Register handlers
	application.RegisterHandlers()

	address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	logger.Info("starting server", zap.String("address", address))
	if err := application.Start(address); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
