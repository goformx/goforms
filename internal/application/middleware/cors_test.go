package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/goformx/goforms/internal/application/middleware"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
)

func TestCORS_AllowOrigin(t *testing.T) {
	e := echo.New()
	cfg := &appconfig.SecurityConfig{
		CORS: appconfig.CORSConfig{
			AllowedOrigins:   []string{"https://example.com"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
			MaxAge:           3600,
		},
	}
	e.Use(middleware.CORS(cfg))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Origin", "https://example.com")

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
}
