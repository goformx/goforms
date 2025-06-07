package middleware

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// RecoveryMiddleware handles panic recovery and error mapping
type RecoveryMiddleware struct {
	logger logging.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger logging.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// WithRecovery returns a middleware that recovers from panics and maps errors
func (m *RecoveryMiddleware) WithRecovery() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					m.logger.Error("panic recovered",
						logging.String("operation", "recovery"),
						logging.Error(err),
					)
					if err := c.JSON(http.StatusInternalServerError, map[string]string{
						"error": "Internal server error",
					}); err != nil {
						m.logger.Error("failed to send error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
				}
			}()

			err := next(c)
			if err == nil {
				return nil
			}

			var domainErr *errors.DomainError
			if stderrors.As(err, &domainErr) {
				switch domainErr.Code {
				case errors.ErrCodeValidation:
					if err := c.JSON(http.StatusBadRequest, domainErr); err != nil {
						m.logger.Error("failed to send validation error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeRequired:
					if err := c.JSON(http.StatusBadRequest, domainErr); err != nil {
						m.logger.Error("failed to send required error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeInvalid:
					if err := c.JSON(http.StatusBadRequest, domainErr); err != nil {
						m.logger.Error("failed to send invalid error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeInvalidFormat:
					if err := c.JSON(http.StatusBadRequest, domainErr); err != nil {
						m.logger.Error("failed to send invalid format error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeInvalidInput:
					if err := c.JSON(http.StatusBadRequest, domainErr); err != nil {
						m.logger.Error("failed to send invalid input error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeInvalidToken:
					if err := c.JSON(http.StatusUnauthorized, domainErr); err != nil {
						m.logger.Error("failed to send invalid token error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeAuthentication:
					if err := c.JSON(http.StatusUnauthorized, domainErr); err != nil {
						m.logger.Error("failed to send authentication error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeInsufficientRole:
					if err := c.JSON(http.StatusForbidden, domainErr); err != nil {
						m.logger.Error("failed to send insufficient role error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeBadRequest:
					if err := c.JSON(http.StatusBadRequest, domainErr); err != nil {
						m.logger.Error("failed to send bad request error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeServerError:
					if err := c.JSON(http.StatusInternalServerError, domainErr); err != nil {
						m.logger.Error("failed to send server error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeAlreadyExists:
					if err := c.JSON(http.StatusConflict, domainErr); err != nil {
						m.logger.Error("failed to send already exists error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeStartup:
					if err := c.JSON(http.StatusServiceUnavailable, domainErr); err != nil {
						m.logger.Error("failed to send startup error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeShutdown:
					if err := c.JSON(http.StatusServiceUnavailable, domainErr); err != nil {
						m.logger.Error("failed to send shutdown error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeConfig:
					if err := c.JSON(http.StatusInternalServerError, domainErr); err != nil {
						m.logger.Error("failed to send config error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeDatabase:
					if err := c.JSON(http.StatusInternalServerError, domainErr); err != nil {
						m.logger.Error("failed to send database error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				case errors.ErrCodeTimeout:
					if err := c.JSON(http.StatusGatewayTimeout, domainErr); err != nil {
						m.logger.Error("failed to send timeout error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				default:
					if err := c.JSON(http.StatusInternalServerError, domainErr); err != nil {
						m.logger.Error("failed to send unknown error response",
							logging.String("operation", "recovery"),
							logging.Error(err),
						)
					}
					return nil
				}
			}

			var echoErr *echo.HTTPError
			if stderrors.As(err, &echoErr) {
				if err := c.JSON(echoErr.Code, echoErr); err != nil {
					m.logger.Error("failed to send echo error response",
						logging.String("operation", "recovery"),
						logging.Error(err),
					)
				}
				return nil
			}

			m.logger.Error("unhandled error",
				logging.String("operation", "recovery"),
				logging.Error(err),
			)
			if err := c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Internal server error",
			}); err != nil {
				m.logger.Error("failed to send error response",
					logging.String("operation", "recovery"),
					logging.Error(err),
				)
			}
			return nil
		}
	}
}
