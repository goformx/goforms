package health

import (
	"context"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/services/health"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	pingTimeout = 5 * time.Second
)

// Repository implements health.Repository interface
type Repository struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewRepository creates a new health repository
func NewRepository(db *database.GormDB, logger logging.Logger) health.Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// PingContext checks if the database is accessible
func (r *Repository) PingContext(ctx context.Context) error {
	r.logger.Debug("performing database health check")

	// Get underlying *sql.DB
	sqlDB, err := r.db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set a timeout for the ping
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	// Ping the database
	if pingErr := sqlDB.PingContext(pingCtx); pingErr != nil {
		r.logger.Error("database health check failed",
			logging.ErrorField("error", pingErr),
		)
		return fmt.Errorf("failed to ping database: %w", pingErr)
	}

	r.logger.Debug("database health check passed")
	return nil
}

// Check checks the health of the database
func (r *Repository) Check(ctx context.Context) error {
	return r.PingContext(ctx)
}
