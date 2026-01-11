package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/access"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
)

// Test helpers
func createTestLogger(ctrl *gomock.Controller) *mocklogging.MockLogger {
	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().WithComponent(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithRequestID(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	return logger
}

func createTestConfig() *appconfig.Config {
	return &appconfig.Config{
		App: appconfig.AppConfig{
			Environment:    "test",
			RequestTimeout: 30 * time.Second,
		},
		Security: appconfig.SecurityConfig{
			RateLimit: appconfig.RateLimitConfig{
				Enabled:  true,
				Requests: 1, // 1 request per second
				Burst:    1,
				Window:   time.Second,
			},
		},
	}
}

func createTestAccessManager() *access.Manager {
	return access.NewManager(&access.Config{
		DefaultAccess: access.Public,
		PublicPaths:   []string{"/"},
	}, []access.Rule{})
}

func createTestManager(t *testing.T, ctrl *gomock.Controller) (*middleware.Manager, *echo.Echo) {
	t.Helper()

	e := echo.New()
	cfg := createTestConfig()
	logger := createTestLogger(ctrl)
	sanitizer := sanitization.NewService()
	accessManager := createTestAccessManager()

	manager := middleware.NewManager(&middleware.ManagerConfig{
		Logger:         logger,
		Config:         cfg,
		UserService:    nil, // Not needed for this test
		FormService:    nil, // Not needed for this test
		SessionManager: nil, // Not needed for rate limiting test
		AccessManager:  accessManager,
		Sanitizer:      sanitizer,
	})

	manager.Setup(e)

	return manager, e
}

func TestManager_RateLimiter_BlocksAfterBurst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, e := createTestManager(t, ctrl)

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

func TestManager_RateLimiter_Scenarios(t *testing.T) {
	tests := []struct {
		name           string
		requests       int
		burst          int
		window         time.Duration
		requestCount   int
		expectedStatus []int
	}{
		{
			name:           "single request allowed",
			requests:       1,
			burst:          1,
			window:         time.Second,
			requestCount:   1,
			expectedStatus: []int{http.StatusOK},
		},
		{
			name:           "burst exceeded",
			requests:       1,
			burst:          1,
			window:         time.Second,
			requestCount:   2,
			expectedStatus: []int{http.StatusOK, http.StatusTooManyRequests},
		},
		{
			name:           "higher burst allows more requests",
			requests:       1,
			burst:          3,
			window:         time.Second,
			requestCount:   3,
			expectedStatus: []int{http.StatusOK, http.StatusOK, http.StatusOK},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create config with test parameters
			cfg := createTestConfig()
			cfg.Security.RateLimit.Requests = tt.requests
			cfg.Security.RateLimit.Burst = tt.burst
			cfg.Security.RateLimit.Window = tt.window

			e := echo.New()
			logger := createTestLogger(ctrl)
			sanitizer := sanitization.NewService()
			accessManager := createTestAccessManager()

			manager := middleware.NewManager(&middleware.ManagerConfig{
				Logger:         logger,
				Config:         cfg,
				UserService:    nil,
				FormService:    nil,
				SessionManager: nil,
				AccessManager:  accessManager,
				Sanitizer:      sanitizer,
			})

			manager.Setup(e)

			e.GET("/", func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			// Make requests and check status codes
			for i := range tt.requestCount {
				req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				req.Header.Set("X-Real-IP", "192.168.1.1")

				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)

				expectedStatus := tt.expectedStatus[i]
				assert.Equal(t, expectedStatus, rec.Code,
					"Request %d: expected status %d, got %d", i+1, expectedStatus, rec.Code)
			}
		})
	}
}
