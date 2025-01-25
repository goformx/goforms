package store

import (
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// NewUserStore creates a new user store
func NewUserStore(db *database.Database, logger logging.Logger) user.Store {
	return &Store{
		db:  db.DB,
		log: logger,
	}
}
