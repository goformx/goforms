package database

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/config/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite
	cfg    *config.Config
	rootDB *sqlx.DB
}

func (s *DatabaseTestSuite) SetupSuite() {
	s.cfg = &config.Config{
		Database: database.Config{
			Host: "localhost",
			Port: 3306,
			Credentials: database.Credentials{
				User:     "goforms",
				Password: "goforms",
				DBName:   "goforms_test",
			},
			ConnectionPool: database.PoolConfig{
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
			},
		},
	}

	// Create root connection
	var err error
	s.rootDB, err = sqlx.Connect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/",
		s.cfg.Database.Credentials.User,
		s.cfg.Database.Credentials.Password,
		s.cfg.Database.Host,
		s.cfg.Database.Port,
	))
	require.NoError(s.T(), err)

	// Create test database
	_, err = s.rootDB.Exec("CREATE DATABASE IF NOT EXISTS goforms_test")
	require.NoError(s.T(), err)
}

func (s *DatabaseTestSuite) TearDownSuite() {
	if s.rootDB != nil {
		_, err := s.rootDB.Exec("DROP DATABASE IF EXISTS goforms_test")
		if err != nil {
			s.T().Logf("Failed to drop test database: %v", err)
		}
		s.rootDB.Close()
	}
}

func (s *DatabaseTestSuite) TestNewDatabase() {
	db, err := New(s.cfg)
	require.NoError(s.T(), err)
	defer db.Close()

	// Test connection settings
	assert.Equal(s.T(), s.cfg.Database.ConnectionPool.MaxOpenConns, db.Stats().MaxOpenConnections)

	// Test connection is alive
	err = db.Ping()
	assert.NoError(s.T(), err)
}

func (s *DatabaseTestSuite) TestNewDatabaseError() {
	invalidCfg := &config.Config{
		Database: database.Config{
			Host: "invalid",
			Port: 3306,
			Credentials: database.Credentials{
				User:     "invalid",
				Password: "invalid",
				DBName:   "invalid",
			},
		},
	}

	_, err := New(invalidCfg)
	assert.Error(s.T(), err)
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
