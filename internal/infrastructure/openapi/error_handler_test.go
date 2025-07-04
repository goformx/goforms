package openapi_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/infrastructure/openapi"
	"github.com/goformx/goforms/test/mocks/logging"
)

// setupTest creates a test setup with logger and config
func setupTest(t *testing.T) (*gomock.Controller, *logging.MockLogger, *openapi.Config) {
	t.Helper()
	ctrl := gomock.NewController(t)
	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: true,
	}

	return ctrl, logger, config
}

func TestNewValidationErrorHandler(t *testing.T) {
	ctrl, logger, config := setupTest(t)
	defer ctrl.Finish()

	handler := openapi.NewValidationErrorHandler(logger, config)

	assert.NotNil(t, handler)
	assert.Implements(t, (*openapi.ValidationErrorHandler)(nil), handler)
}

func TestValidationErrorHandler_HandleError_RequestValidationError_Blocking(t *testing.T) {
	ctrl, logger, _ := setupTest(t)
	defer ctrl.Finish()

	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
		"ip":     "127.0.0.1",
	}

	// Expect logging
	logger.EXPECT().Warn("Request validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, openapi.RequestValidationError, metadata)

	require.Error(t, result)
	assert.IsType(t, &echo.HTTPError{}, result)

	httpError, ok := result.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
	assert.Contains(t, httpError.Message, "Request validation failed")
	assert.Contains(t, httpError.Message, "test validation error")
}

func TestValidationErrorHandler_HandleError_ResponseValidationError_Blocking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  false,
		BlockInvalidResponses: true,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
		"status": 500,
	}

	// Expect logging
	logger.EXPECT().Warn("Response validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, openapi.ResponseValidationError, metadata)

	require.Error(t, result)
	assert.IsType(t, &echo.HTTPError{}, result)

	httpError, ok := result.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpError.Code)
	assert.Contains(t, httpError.Message, "Response validation failed")
	assert.Contains(t, httpError.Message, "test validation error")
}

func TestValidationErrorHandler_HandleError_RequestValidationError_NonBlocking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  false,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
		"ip":     "127.0.0.1",
	}

	// Expect logging
	logger.EXPECT().Warn("Request validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, openapi.RequestValidationError, metadata)

	require.NoError(t, result) // Should not block, just log
}

func TestValidationErrorHandler_HandleError_ResponseValidationError_NonBlocking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  false,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
		"status": 500,
	}

	// Expect logging
	logger.EXPECT().Warn("Response validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, openapi.ResponseValidationError, metadata)

	require.NoError(t, result) // Should not block, just log
}

func TestValidationErrorHandler_HandleError_NoLogging(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   false,
		BlockInvalidRequests:  false,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
	}

	// Should not expect any logging calls

	result := handler.HandleError(ctx, testError, openapi.RequestValidationError, metadata)

	require.NoError(t, result)
}

func TestValidationErrorHandler_HandleError_UnknownErrorType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: true,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
	}

	// Expect logging
	logger.EXPECT().Warn("Validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, "unknown", metadata)

	require.NoError(t, result) // Unknown error types should not block
}

func TestValidationErrorHandler_HandleError_EmptyMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{}

	// Expect logging
	logger.EXPECT().Warn("Request validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, openapi.RequestValidationError, metadata)

	require.Error(t, result)
	assert.IsType(t, &echo.HTTPError{}, result)
}

func TestValidationErrorHandler_HandleError_NilMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")

	// Expect logging
	logger.EXPECT().Warn("Request validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, openapi.RequestValidationError, nil)

	require.Error(t, result)
	assert.IsType(t, &echo.HTTPError{}, result)
}

func TestValidationErrorHandler_HandleError_NilError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
	}

	// Expect logging
	logger.EXPECT().Warn("Request validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, nil, openapi.RequestValidationError, metadata)

	require.Error(t, result)
	assert.IsType(t, &echo.HTTPError{}, result)
}

func TestValidationErrorHandler_HandleError_ComplexMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: false,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":      "/api/v1/users/123",
		"method":    "POST",
		"ip":        "192.168.1.100",
		"user_id":   "user123",
		"timestamp": "2023-01-01T00:00:00Z",
		"headers": map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
	}

	// Expect logging
	logger.EXPECT().Warn("Request validation failed", gomock.Any()).Times(1)

	result := handler.HandleError(ctx, testError, openapi.RequestValidationError, metadata)

	require.Error(t, result)
	assert.IsType(t, &echo.HTTPError{}, result)
}

func TestValidationErrorHandler_HandleError_AllErrorTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := logging.NewMockLogger(ctrl)
	config := &openapi.Config{
		LogValidationErrors:   true,
		BlockInvalidRequests:  true,
		BlockInvalidResponses: true,
	}

	handler := openapi.NewValidationErrorHandler(logger, config)
	ctx := context.Background()
	testError := errors.New("test validation error")
	metadata := map[string]any{
		"path":   "/test",
		"method": "GET",
	}

	// Test RequestValidationError
	logger.EXPECT().Warn("Request validation failed", gomock.Any()).Times(1)

	result1 := handler.HandleError(ctx, testError, openapi.RequestValidationError, metadata)
	require.Error(t, result1)
	assert.IsType(t, &echo.HTTPError{}, result1)
	httpError, ok := result1.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)

	// Test ResponseValidationError
	logger.EXPECT().Warn("Response validation failed", gomock.Any()).Times(1)

	result2 := handler.HandleError(ctx, testError, openapi.ResponseValidationError, metadata)
	require.Error(t, result2)
	assert.IsType(t, &echo.HTTPError{}, result2)
	httpError2, ok := result2.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpError2.Code)

	// Test unknown error type
	logger.EXPECT().Warn("Validation failed", gomock.Any()).Times(1)

	result3 := handler.HandleError(ctx, testError, "unknown", metadata)
	require.NoError(t, result3)
}
