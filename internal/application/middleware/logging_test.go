package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/middleware"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestLoggingMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocklogging.NewMockLogger(ctrl)
	loggingMiddleware := middleware.LoggingMiddleware(mockLogger)

	handler := func(c echo.Context) error {
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	mockLogger.EXPECT().Info("request completed", gomock.Any()).Times(1)

	if err := loggingMiddleware(handler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoggingMiddlewareWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocklogging.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info("http request", gomock.Any()).Times(1)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	}

	h := middleware.LoggingMiddleware(mockLogger)(handler)
	_ = h(c)
}

func TestLoggingMiddlewareWithPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocklogging.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info("http request", gomock.Any()).Times(1)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		panic("test panic")
	}

	h := middleware.LoggingMiddleware(mockLogger)(handler)
	_ = h(c)
}

func TestLoggingMiddleware_RealIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocklogging.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info("http request", gomock.Any()).Times(1)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := middleware.LoggingMiddleware(mockLogger)
	handler := mw(func(c echo.Context) error {
		c.Response().WriteHeader(http.StatusOK)
		return c.String(http.StatusOK, "success")
	})

	if err := handler(c); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
