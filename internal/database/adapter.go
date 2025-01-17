package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/models"
)

// AsDB converts *sqlx.DB to models.DB interface
func AsDB(db *sqlx.DB) models.DB {
	return db
}
