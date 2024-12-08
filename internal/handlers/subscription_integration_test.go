package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
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
	handler := NewSubscriptionHandler(testDB, logger)

	t.Run("Full subscription flow", func(t *testing.T) {
		// Clean up any existing test data
		_, err := testDB.Exec("DELETE FROM subscriptions WHERE email = $1", "integration@test.com")
		require.NoError(t, err)

		// Test subscription creation
		ctx := context.Background()
		err = handler.CreateSubscription(ctx, "integration@test.com")
		assert.NoError(t, err)

		// Verify subscription exists
		var exists bool
		err = testDB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = $1)", "integration@test.com")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Test duplicate subscription
		err = handler.CreateSubscription(ctx, "integration@test.com")
		assert.Error(t, err)
	})
}
