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
	cfg, err := config.Load()
	require.NoError(s.T(), err)

	s.db, err = database.New(cfg)
	require.NoError(s.T(), err)

	s.logger, _ = zap.NewDevelopment()
	store := models.NewSubscriptionStore(s.db)
	s.handler = NewSubscriptionHandler(s.logger, store)
}

func (s *SubscriptionTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *SubscriptionTestSuite) TestSubscriptionIntegration() {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		s.T().Skip("Skipping integration test in CI environment")
	}

	// Clean up any existing test data
	_, err := s.db.Exec("DELETE FROM subscriptions WHERE email = $1", "integration@test.com")
	require.NoError(s.T(), err)

	// Create a mock echo.Context
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(`{
		"email": "integration@test.com",
		"name": "Test User"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test subscription creation
	err = s.handler.CreateSubscription(c)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	// Verify subscription exists
	var exists bool
	err = s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = $1)", "integration@test.com")
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)

	// Test duplicate subscription
	err = s.handler.CreateSubscription(c)
	assert.Error(s.T(), err)
}

func TestSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}
