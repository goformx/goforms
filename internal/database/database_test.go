package database

import (
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/config/database"
	"github.com/jonesrussell/goforms/test/setup"
	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite
	db *sqlx.DB
}

func (s *DatabaseTestSuite) SetupSuite() {
	testDB, err := setup.NewTestDB()
	s.Require().NoError(err)
	s.db = testDB.DB
}

func (s *DatabaseTestSuite) TestNewDatabase() {
	cfg := &config.Config{
		Database: database.Config{
			Host:           os.Getenv("DB_HOST"),
			Port:           3306,
			User:           os.Getenv("DB_USER"),
			Password:       os.Getenv("DB_PASSWORD"),
			DBName:         os.Getenv("DB_DATABASE"),
			MaxOpenConns:   10,
			MaxIdleConns:   5,
			ConnMaxLifetme: time.Hour,
		},
	}

	db, err := New(cfg)
	s.Require().NoError(err)
	s.NotNil(db)
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
