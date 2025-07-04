package openapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// validationErrorHandler implements ValidationErrorHandler
type validationErrorHandler struct {
	logger logging.Logger
	config *Config
}

// NewValidationErrorHandler creates a new error handler
func NewValidationErrorHandler(logger logging.Logger, config *Config) ValidationErrorHandler {
	return &validationErrorHandler{logger: logger, config: config}
}

// HandleError handles validation errors consistently
func (h *validationErrorHandler) HandleError(
	ctx context.Context,
	err error,
	errorType ValidationErrorType,
	metadata map[string]interface{},
) error {
	var (
		shouldBlock  bool
		errorMessage string
	)

	switch errorType {
	case RequestValidationError:
		shouldBlock = h.config.BlockInvalidRequests
		errorMessage = "Request validation failed"
	case ResponseValidationError:
		shouldBlock = h.config.BlockInvalidResponses
		errorMessage = "Response validation failed"
	default:
		shouldBlock = false
		errorMessage = "Validation failed"
	}

	if shouldBlock {
		statusCode := http.StatusBadRequest
		if errorType == ResponseValidationError {
			statusCode = http.StatusInternalServerError
		}

		return echo.NewHTTPError(statusCode, fmt.Sprintf("%s: %v", errorMessage, err))
	}

	if h.config.LogValidationErrors {
		logFields := []interface{}{"error", err}
		for key, value := range metadata {
			logFields = append(logFields, key, value)
		}

		h.logger.Warn(errorMessage, logFields...)
	}

	return nil
}
