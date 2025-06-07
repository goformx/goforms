package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// RecoveryMiddleware provides panic recovery for HTTP requests
type RecoveryMiddleware struct {
	logger logging.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger logging.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// WithRecovery adds panic recovery to the request
func (m *RecoveryMiddleware) WithRecovery() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get request context and logger
			ctx := c.Request().Context()
			requestID := GetRequestIDFromContext(ctx)
			logger := GetLoggerFromContext(ctx)
			if logger == nil {
				logger = m.logger
			}

			// Add request timing
			startTime := time.Now()

			// Recover from panics
			defer func() {
				if r := recover(); r != nil {
					// Log the panic with stack trace
					logger.Error("Panic recovered",
						logging.String("request_id", requestID),
						logging.String("error", fmt.Sprintf("%v", r)),
						logging.String("stack", string(debug.Stack())),
						logging.Duration("duration", time.Since(startTime)),
					)

					// Create error response
					err := errors.New(errors.ErrCodeServerError, "Internal server error", nil)
					if err := c.JSON(http.StatusInternalServerError, map[string]string{
						"error": err.Error(),
					}); err != nil {
						logger.Error("Failed to send error response",
							logging.String("request_id", requestID),
							logging.Error(err),
						)
					}
				}
			}()

			// Handle request with timeout
			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			// Wait for request completion or timeout
			select {
			case err := <-done:
				if err != nil {
					// Log the error with context
					logger.Error("Request error",
						logging.String("request_id", requestID),
						logging.Error(err),
						logging.Duration("duration", time.Since(startTime)),
					)

					// Handle domain errors
					if domainErr, ok := err.(*errors.DomainError); ok {
						switch domainErr.Code {
						case errors.ErrCodeValidation:
							return c.JSON(http.StatusBadRequest, map[string]any{
								"error":   domainErr.Error(),
								"code":    domainErr.Code,
								"details": domainErr.Context,
							})
						case errors.ErrCodeUnauthorized:
							return c.JSON(http.StatusUnauthorized, map[string]any{
								"error":   domainErr.Error(),
								"code":    domainErr.Code,
								"details": domainErr.Context,
							})
						case errors.ErrCodeForbidden:
							return c.JSON(http.StatusForbidden, map[string]any{
								"error":   domainErr.Error(),
								"code":    domainErr.Code,
								"details": domainErr.Context,
							})
						case errors.ErrCodeNotFound:
							return c.JSON(http.StatusNotFound, map[string]any{
								"error":   domainErr.Error(),
								"code":    domainErr.Code,
								"details": domainErr.Context,
							})
						case errors.ErrCodeConflict:
							return c.JSON(http.StatusConflict, map[string]any{
								"error":   domainErr.Error(),
								"code":    domainErr.Code,
								"details": domainErr.Context,
							})
						default:
							return c.JSON(http.StatusInternalServerError, map[string]any{
								"error":   "Internal server error",
								"code":    errors.ErrCodeServerError,
								"details": domainErr.Context,
							})
						}
					}

					// Handle echo errors
					if echoErr, ok := err.(*echo.HTTPError); ok {
						var message string
						switch msg := echoErr.Message.(type) {
						case string:
							message = msg
						case error:
							message = msg.Error()
						default:
							message = fmt.Sprintf("%v", msg)
						}
						return c.JSON(echoErr.Code, map[string]any{
							"error": message,
							"code":  errors.ErrCodeServerError,
						})
					}

					// Handle unknown errors
					return c.JSON(http.StatusInternalServerError, map[string]any{
						"error": "Internal server error",
						"code":  errors.ErrCodeServerError,
					})
				}
				return nil
			case <-ctx.Done():
				// Handle timeout
				logger.Error("Request timeout",
					logging.String("request_id", requestID),
					logging.Duration("duration", time.Since(startTime)),
				)
				return c.JSON(http.StatusGatewayTimeout, map[string]any{
					"error": "Request timeout",
					"code":  errors.ErrCodeTimeout,
				})
			}
		}
	}
}
