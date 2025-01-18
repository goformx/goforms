package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/config/server"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestAppIntegration(t *testing.T) {
	var app *App
	var e *echo.Echo

	// Create mock implementations
	subscriptionStore := &models.MockSubscriptionStore{}
	contactStore := &models.MockContactStore{}
	pingContexter := &models.MockPingContexter{}

	testApp := fxtest.New(t,
		fx.Provide(
			func() *config.Config {
				return &config.Config{
					Server: server.Config{
						Port: 8081,
					},
				}
			},
			func() *echo.Echo { return echo.New() },
			logger.GetLogger,
			func() models.SubscriptionStore { return subscriptionStore },
			func() handlers.PingContexter { return pingContexter },
			func() models.ContactStore { return contactStore },
			handlers.NewSubscriptionHandler,
			handlers.NewHealthHandler,
			handlers.NewContactHandler,
			handlers.NewMarketingHandler,
			NewApp,
		),
		fx.Populate(&app, &e),
	)

	require.NoError(t, testApp.Start(context.Background()))
	defer func() {
		err := testApp.Stop(context.Background())
		require.NoError(t, err)
	}()

	// Test health check endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test home page
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Welcome to Goforms")

	// Test contact page
	req = httptest.NewRequest(http.MethodGet, "/contact", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Contact Form Demo")
}
