package middleware

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/goformx/goforms/internal/application/constants"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
)

// RateLimiter creates and configures rate limiter middleware
func RateLimiter(securityConfig *appconfig.SecurityConfig) echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(securityConfig.RateLimit.Requests),
				Burst:     securityConfig.RateLimit.Burst,
				ExpiresIn: securityConfig.RateLimit.Window,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// For login and signup pages, use IP address as identifier
			path := c.Request().URL.Path
			if path == constants.PathLogin || path == constants.PathSignup {
				return c.RealIP(), nil
			}

			// For form submissions, use form ID and origin
			formID := c.Param("formID")
			origin := c.Request().Header.Get("Origin")
			if formID == "" {
				formID = constants.DefaultUnknown
			}
			if origin == "" {
				origin = constants.DefaultUnknown
			}

			return fmt.Sprintf("%s:%s", formID, origin), nil
		},
		ErrorHandler: func(_ echo.Context, _ error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests,
				"Rate limit exceeded: too many requests from the same form or origin")
		},
		DenyHandler: func(_ echo.Context, _ string, _ error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests,
				"Rate limit exceeded: please try again later")
		},
	})
}
