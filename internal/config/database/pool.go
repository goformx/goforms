package database

import "time"

// PoolConfig contains database connection pool settings
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}
