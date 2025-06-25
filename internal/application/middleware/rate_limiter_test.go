package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/goformx/goforms/internal/application/middleware"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
)

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	e := echo.New()
	cfg := &appconfig.SecurityConfig{
		RateLimit: appconfig.RateLimitConfig{
			Enabled:  true,
			Requests: 1, // 1 request per second
			Burst:    1,
			Window:   time.Second,
		},
	}
	e.Use(middleware.RateLimiter(cfg))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req1.Header.Set("X-Real-IP", "192.168.1.1")
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req2.Header.Set("X-Real-IP", "192.168.1.1")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}
