package app

import (
	"context"
	"html/template"
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
	tmpl := template.Must(template.New("test").Parse("{{.}}"))

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
			func() *template.Template { return tmpl },
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
	defer testApp.Stop(context.Background())

	// Test health check endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTemplateRendering(t *testing.T) {
	tmpl := template.Must(template.New("test").Parse("{{.}}"))
	rec := httptest.NewRecorder()
	err := tmpl.Execute(rec, "test content")
	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "test content")
}
