package models

import (
	"context"
	"database/sql"
)

// DB interface defines the database operations we need
type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
