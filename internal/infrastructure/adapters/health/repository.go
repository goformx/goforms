package health

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/goformx/goforms/internal/domain/services/health"
)

// Repository implements the health.Repository interface
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new health check repository
func NewRepository(db *sqlx.DB) health.Repository {
	return &Repository{
		db: db,
	}
}

// PingContext checks if the database is accessible
func (r *Repository) PingContext(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
