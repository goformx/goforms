package store

import (
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// NewUserStore creates a new user store
func NewUserStore(db *database.Database, logger logging.Logger) user.Store {
	return &Store{
		db:  db.DB,
		log: logger,
	}
}
