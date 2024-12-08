package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
	"github.com/jonesrussell/goforms/internal/models"
)

type SubscriptionTestSuite struct {
	suite.Suite
	db      *sqlx.DB
	handler *SubscriptionHandler
	logger  *zap.Logger
}

func (s *SubscriptionTestSuite) SetupSuite() {
	// Set test environment variables
	os.Setenv("MYSQL_HOSTNAME", "localhost")
	os.Setenv("MYSQL_PORT", "3306")
	os.Setenv("MYSQL_USER", "root")
	os.Setenv("MYSQL_PASSWORD", "rootpassword")
	// Don't set MYSQL_DATABASE yet as we need to create it first
	os.Setenv("MYSQL_MAX_OPEN_CONNS", "25")
	os.Setenv("MYSQL_MAX_IDLE_CONNS", "5")
	os.Setenv("MYSQL_CONN_MAX_LIFETIME", "5m")

	// Connect to MySQL without specifying a database
	db, err := sqlx.Connect("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/",
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_HOSTNAME"),
			os.Getenv("MYSQL_PORT"),
		))
	require.NoError(s.T(), err)

	// Create test database
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS goforms_test")
	require.NoError(s.T(), err)

	// Grant privileges
	_, err = db.Exec("GRANT ALL PRIVILEGES ON goforms_test.* TO 'goforms'@'%'")
	require.NoError(s.T(), err)

	_, err = db.Exec("FLUSH PRIVILEGES")
	require.NoError(s.T(), err)

	// Close initial connection
	db.Close()

	// Now set the database name and create the real connection
	os.Setenv("MYSQL_DATABASE", "goforms_test")
	cfg, err := config.New()
	require.NoError(s.T(), err)

	s.db, err = database.New(cfg)
	require.NoError(s.T(), err)

	// Run migrations
	err = s.setupTestDatabase()
	require.NoError(s.T(), err)

	s.logger, _ = zap.NewDevelopment()
	store := models.NewSubscriptionStore(s.db)
	s.handler = NewSubscriptionHandler(s.logger, store)
}

func (s *SubscriptionTestSuite) TearDownSuite() {
	if s.db != nil {
		// Drop test database
		_, err := s.db.Exec("DROP DATABASE IF EXISTS goforms_test")
		if err != nil {
			s.T().Logf("Failed to drop test database: %v", err)
		}

		s.db.Close()
	}
}

func (s *SubscriptionTestSuite) SetupTest() {
	// Clean up any existing test data before each test
	_, err := s.db.Exec("DELETE FROM subscriptions")
	require.NoError(s.T(), err)
}

func (s *SubscriptionTestSuite) setupTestDatabase() error {
	migrationPath := "file://../../migrations"

	driver, err := mysql.WithInstance(s.db.DB, &mysql.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}

	// Drop all tables to ensure clean state
	_, err = s.db.Exec("DROP TABLE IF EXISTS schema_migrations, subscriptions")
	if err != nil {
		return err
	}

	// Create schema_migrations table manually
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL,
			dirty boolean NOT NULL,
			PRIMARY KEY (version)
		)
	`)
	if err != nil {
		return err
	}

	// Run migrations without trying to force version
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (s *SubscriptionTestSuite) TestSubscriptionIntegration() {
	// Test subscription creation
	requestBody := map[string]string{
		"email": "integration@test.com",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(s.T(), err)

	e := echo.New()
	// Create initial request
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	s.logger.Info("testing subscription creation")
	err = s.handler.CreateSubscription(c)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), requestBody["email"], response["email"])

	var exists bool
	err = s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = ?)", requestBody["email"])
	require.NoError(s.T(), err)
	assert.True(s.T(), exists)

	// Test duplicate subscription with a fresh request
	s.logger.Info("testing duplicate subscription")
	duplicateReq := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBuffer(body))
	duplicateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	duplicateRec := httptest.NewRecorder()
	duplicateC := e.NewContext(duplicateReq, duplicateRec)

	err = s.handler.CreateSubscription(duplicateC)

	// Handle the error using Echo's default error handler
	if err != nil {
		he, ok := err.(*echo.HTTPError)
		assert.True(s.T(), ok)
		assert.Equal(s.T(), http.StatusConflict, he.Code)
		assert.Equal(s.T(), "Email already subscribed", he.Message)

		// Set the status code in the recorder
		duplicateRec.Code = he.Code
		// Write the error response
		_ = json.NewEncoder(duplicateRec.Body).Encode(map[string]string{
			"error": he.Message.(string),
		})
	}

	// Verify the response matches what we expect
	assert.Equal(s.T(), http.StatusConflict, duplicateRec.Code)
	var errResponse map[string]string
	err = json.NewDecoder(duplicateRec.Body).Decode(&errResponse)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Email already subscribed", errResponse["error"])
}

func TestSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}
