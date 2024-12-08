package handlers

import (
	"net/http"
	"testing"

	"github.com/jonesrussell/goforms/internal/models"
	"github.com/jonesrussell/goforms/test/fixtures"
	"github.com/jonesrussell/goforms/test/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type SubscriptionTestSuite struct {
	suite.Suite
	handler *SubscriptionHandler
	testDB  *setup.TestDB
	fixture *fixtures.SubscriptionFixture
}

func (s *SubscriptionTestSuite) SetupSuite() {
	var err error

	// Setup test database
	s.testDB, err = setup.NewTestDB()
	require.NoError(s.T(), err)

	// Run migrations
	err = s.testDB.RunMigrations()
	require.NoError(s.T(), err)

	// Create a production logger for tests (less verbose)
	logger := zap.NewProductionConfig()
	logger.Level = zap.NewAtomicLevelAt(zap.WarnLevel) // Only show warnings and errors
	log, _ := logger.Build()

	// Setup handler
	store := models.NewSubscriptionStore(s.testDB.DB)
	s.handler = NewSubscriptionHandler(log, store)

	// Setup fixture
	s.fixture = fixtures.NewSubscriptionFixture(s.handler.CreateSubscription)
}

func (s *SubscriptionTestSuite) TearDownSuite() {
	if s.testDB != nil {
		if err := s.testDB.Cleanup(); err != nil {
			s.T().Logf("Failed to cleanup test data: %v", err)
		}
	}
}

func (s *SubscriptionTestSuite) SetupTest() {
	err := s.testDB.ClearData()
	require.NoError(s.T(), err)
}

func (s *SubscriptionTestSuite) TestSubscriptionIntegration() {
	// Test successful subscription
	rec, err := s.fixture.CreateSubscriptionRequest("integration@test.com")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = fixtures.ParseResponse(rec, &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "integration@test.com", response["email"])

	// Verify database record
	var exists bool
	err = s.testDB.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = ?)", "integration@test.com")
	require.NoError(s.T(), err)
	assert.True(s.T(), exists)

	// Test duplicate subscription
	rec, err = s.fixture.CreateSubscriptionRequest("integration@test.com")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusConflict, rec.Code)

	var errResponse map[string]string
	err = fixtures.ParseResponse(rec, &errResponse)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Email already subscribed", errResponse["error"])

	// Test invalid origin
	rec, err = s.fixture.CreateSubscriptionRequestWithOrigin("new@test.com", "https://invalid-origin.com")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusForbidden, rec.Code)

	err = fixtures.ParseResponse(rec, &errResponse)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "invalid origin", errResponse["error"])
}

func TestSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}
