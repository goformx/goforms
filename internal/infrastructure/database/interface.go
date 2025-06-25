// Package database provides database connection and ORM utilities for the application.
package database

import (
	"context"

	"gorm.io/gorm"
)

// DB defines the interface for database operations
type DB interface {
	// Close closes the database connection
	Close() error

	// MonitorConnectionPool monitors the database connection pool and logs metrics
	MonitorConnectionPool(ctx context.Context)

	// Ping pings the database to verify the connection
	Ping(ctx context.Context) error

	// GetDB returns the underlying GORM DB instance
	GetDB() *gorm.DB
}

// Ensure GormDB implements DB interface
var _ DB = (*GormDB)(nil)
