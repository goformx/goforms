package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
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
	os.Setenv("MYSQL_USER", "goforms")
	os.Setenv("MYSQL_PASSWORD", "goforms")
	os.Setenv("MYSQL_DATABASE", "goforms")
	os.Setenv("MYSQL_MAX_OPEN_CONNS", "25")
	os.Setenv("MYSQL_MAX_IDLE_CONNS", "5")
	os.Setenv("MYSQL_CONN_MAX_LIFETIME", "5m")

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
		driver, err := mysql.WithInstance(s.db.DB, &mysql.Config{})
		if err != nil {
			s.T().Logf("Failed to create driver instance: %v", err)
			return
		}

		m, err := migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"mysql",
			driver,
		)
		if err != nil {
			s.T().Logf("Failed to create migrate instance: %v", err)
			return
		}

		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			s.T().Logf("Failed to run down migrations: %v", err)
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
	// Use golang-migrate to run migrations
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

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (s *SubscriptionTestSuite) TestSubscriptionIntegration() {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		s.T().Skip("Skipping integration test in CI environment")
	}

	// Test subscription creation
	requestBody := strings.NewReader(`{
		"email": "integration@test.com",
		"name": "Test User"
	}`)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", requestBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := s.handler.CreateSubscription(c)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	// Verify subscription exists
	var exists bool
	err = s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = ?)", "integration@test.com")
	require.NoError(s.T(), err)
	assert.True(s.T(), exists)

	// Test duplicate subscription
	requestBody = strings.NewReader(`{
		"email": "integration@test.com",
		"name": "Test User"
	}`)
	req = httptest.NewRequest(http.MethodPost, "/api/subscriptions", requestBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = s.handler.CreateSubscription(c)
	assert.Error(s.T(), err)
}

func TestSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}
