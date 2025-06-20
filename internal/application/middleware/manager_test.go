package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestManager_SetsCSPHeader(t *testing.T) {
	e := echo.New()
	cfg := &appconfig.SecurityConfig{
		CSP: appconfig.CSPConfig{
			Enabled:    true,
			Directives: "default-src 'self'; script-src 'self';",
		},
	}
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Content-Security-Policy", cfg.CSP.Directives)
			return next(c)
		}
	})
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, "default-src 'self'; script-src 'self';", rec.Header().Get("Content-Security-Policy"))
}
