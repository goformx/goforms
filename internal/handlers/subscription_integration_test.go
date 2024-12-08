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
	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
	"github.com/jonesrussell/goforms/internal/models"
)

var (
	testDB *sqlx.DB
)

func TestMain(m *testing.M) {
	// Setup
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	testDB, err = database.New(cfg)
	if err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

func TestSubscriptionIntegration(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	logger, _ := zap.NewDevelopment()
	store := models.NewSubscriptionStore(testDB)
	handler := NewSubscriptionHandler(logger, store)

	t.Run("Full subscription flow", func(t *testing.T) {
		// Clean up any existing test data
		_, err := testDB.Exec("DELETE FROM subscriptions WHERE email = $1", "integration@test.com")
		require.NoError(t, err)

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
		err = handler.CreateSubscription(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Verify subscription exists
		var exists bool
		err = testDB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = $1)", "integration@test.com")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Test duplicate subscription
		err = handler.CreateSubscription(c)
		assert.Error(t, err)
	})
}
