package middleware

import (
	"fmt"
	"net/http"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					var err error
					switch x := r.(type) {
					case string:
						err = fmt.Errorf("%s", x)
					case error:
						err = x
					default:
						err = fmt.Errorf("unknown panic")
					}

					logger.Error("panic recovered",
						logging.String("error", err.Error()),
						logging.String("path", c.Request().URL.Path),
					)

					// Check if it's a domain error
					if domainErr, ok := err.(*domainerrors.DomainError); ok {
						switch domainErr.Code {
						case domainerrors.ErrCodeNotFound:
							if jsonErr := c.JSON(http.StatusNotFound, map[string]string{
								"error": domainErr.Message,
							}); jsonErr != nil {
								logger.Error("failed to send error response", logging.Error(jsonErr))
							}
							return
						case domainerrors.ErrCodeValidation:
							if jsonErr := c.JSON(http.StatusBadRequest, map[string]string{
								"error": domainErr.Message,
							}); jsonErr != nil {
								logger.Error("failed to send error response", logging.Error(jsonErr))
							}
							return
						case domainerrors.ErrCodeUnauthorized:
							if jsonErr := c.JSON(http.StatusUnauthorized, map[string]string{
								"error": domainErr.Message,
							}); jsonErr != nil {
								logger.Error("failed to send error response", logging.Error(jsonErr))
							}
							return
						case domainerrors.ErrCodeForbidden:
							if jsonErr := c.JSON(http.StatusForbidden, map[string]string{
								"error": domainErr.Message,
							}); jsonErr != nil {
								logger.Error("failed to send error response", logging.Error(jsonErr))
							}
							return
						}
					}

					// Default error response
					if jsonErr := c.JSON(http.StatusInternalServerError, map[string]string{
						"error": "Internal server error",
					}); jsonErr != nil {
						logger.Error("failed to send error response", logging.Error(jsonErr))
					}
				}
			}()

			return next(c)
		}
	}
}
