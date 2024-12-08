package database

import (
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/config/database"
	"github.com/jonesrussell/goforms/test/setup"
	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite
	db *sqlx.DB
}

func init() {
	// Load .env.test file
	if err := godotenv.Load("../../.env.test"); err != nil {
		// Don't fail if .env.test doesn't exist, we might be using environment variables
		if !os.IsNotExist(err) {
			panic("Error loading .env.test file: " + err.Error())
		}
	}
}

func (s *DatabaseTestSuite) SetupSuite() {
	testDB, err := setup.NewTestDB()
	s.Require().NoError(err)
	s.db = testDB.DB
}

func (s *DatabaseTestSuite) TestNewDatabase() {
	port, _ := strconv.Atoi(os.Getenv("TEST_DB_PORT"))
	if port == 0 {
		port = 3306 // default port if not set
	}

	cfg := &config.Config{
		Database: database.Config{
			Host:           os.Getenv("TEST_DB_HOST"),
			Port:           port,
			User:           os.Getenv("TEST_DB_USER"),
			Password:       os.Getenv("TEST_DB_PASSWORD"),
			Name:           os.Getenv("TEST_DB_NAME"),
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
