package health

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/domain/services/health"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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
		return err
	}

	// Set a timeout for the ping
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ping the database
	if err := sqlDB.PingContext(pingCtx); err != nil {
		r.logger.Error("database health check failed",
			logging.ErrorField("error", err),
		)
		return err
	}

	r.logger.Debug("database health check passed")
	return nil
}
