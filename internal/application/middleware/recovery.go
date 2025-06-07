package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	stderrors "errors"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Recovery returns a middleware that recovers from panics.
func Recovery(logger logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					stack := debug.Stack()
					logger.Error("panic recovered",
						logging.Error(err),
						logging.String("stack", string(stack)),
					)

					handleError(c, err, logger)
				}
			}()

			return next(c)
		}
	}
}

// handleError processes the error and sends an appropriate response
func handleError(c echo.Context, err error, logger logging.Logger) {
	var domainErr *domainerrors.DomainError
	if stderrors.As(err, &domainErr) {
		switch domainErr.Code {
		case domainerrors.ErrCodeNotFound:
			if err := c.JSON(http.StatusNotFound, map[string]string{
				"error": domainErr.Error(),
			}); err != nil {
				logger.Error("failed to send error response",
					logging.Error(err),
				)
			}
			return
		case domainerrors.ErrCodeValidation:
			if err := c.JSON(http.StatusBadRequest, map[string]string{
				"error": domainErr.Error(),
			}); err != nil {
				logger.Error("failed to send error response",
					logging.Error(err),
				)
			}
			return
		case domainerrors.ErrCodeUnauthorized:
			if err := c.JSON(http.StatusUnauthorized, map[string]string{
				"error": domainErr.Error(),
			}); err != nil {
				logger.Error("failed to send error response",
					logging.Error(err),
				)
			}
			return
		case domainerrors.ErrCodeForbidden:
			if err := c.JSON(http.StatusForbidden, map[string]string{
				"error": domainErr.Error(),
			}); err != nil {
				logger.Error("failed to send error response",
					logging.Error(err),
				)
			}
			return
		}
	}

	// Default error response
	if err := c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	}); err != nil {
		logger.Error("failed to send error response",
			logging.Error(err),
		)
	}
}
