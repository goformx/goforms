package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/models"
)

// AsDB converts *sqlx.DB to models.DB interface
func AsDB(db *sqlx.DB) models.DB {
	return db
}

// AsPingContexter converts *sqlx.DB to handlers.PingContexter interface
func AsPingContexter(db *sqlx.DB) handlers.PingContexter {
	return db
}
