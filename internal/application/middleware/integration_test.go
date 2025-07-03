//go:generate mockgen -source=../../infrastructure/logging/types.go -destination=mocks/mock_logger.go -package=mocks
//go:generate mockgen -source=../../infrastructure/sanitization/interface.go -destination=mocks/mock_sanitizer.go -package=mocks

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/mocks"
	"github.com/goformx/goforms/internal/infrastructure/config"
)

func TestManager_MiddlewareSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks.NewMockLogger(ctrl)
	sanitizer := mocks.NewMockServiceInterface(ctrl)

	// Set up mock expectations as needed
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().WithComponent(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().With(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithFields(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithFieldsStructured(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithOperation(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithRequestID(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithUserID(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().WithError(gomock.Any()).Return(logger).AnyTimes()
	logger.EXPECT().SanitizeField(gomock.Any(), gomock.Any()).Return("").AnyTimes()

	cfg := &config.Config{}

	manager := middleware.NewManager(&middleware.ManagerConfig{
		Logger:    logger,
		Config:    cfg,
		Sanitizer: sanitizer,
	})
	e := echo.New()
	manager.Setup(e)

	// Register a simple handler
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}
