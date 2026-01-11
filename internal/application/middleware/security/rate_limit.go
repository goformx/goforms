package security

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// RateLimitExceededMsg is returned when rate limit is exceeded
	RateLimitExceededMsg = "Rate limit exceeded: too many requests from the same form or origin"
	// RateLimitDeniedMsg is returned when a request is denied
	RateLimitDeniedMsg = "Rate limit exceeded: please try again later"
)

// RateLimiter handles rate limiting middleware setup
type RateLimiter struct {
	logger      logging.Logger
	config      *appconfig.Config
	pathChecker PathChecker
}

// PathChecker interface for checking path types
type PathChecker interface {
	IsAuthPath(path string) bool
	IsFormPath(path string) bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(logger logging.Logger, config *appconfig.Config, pathChecker PathChecker) *RateLimiter {
	return &RateLimiter{
		logger:      logger,
		config:      config,
		pathChecker: pathChecker,
	}
}

// Setup creates and configures rate limiting middleware
func (rl *RateLimiter) Setup() echo.MiddlewareFunc {
	rateLimitConfig := rl.config.Security.RateLimit

	if err := rl.validateConfig(rateLimitConfig); err != nil {
		rl.logger.Error("Invalid rate limit configuration", "error", err)
		return noopMiddleware()
	}

	if rl.config.App.IsDevelopment() && !rateLimitConfig.Enabled {
		rl.logger.Info("Rate limiting disabled in development mode")
		return noopMiddleware()
	}

	rl.logger.Info("Setting up rate limiter",
		"enabled", rateLimitConfig.Enabled,
		"requests_per_second", rateLimitConfig.Requests,
		"burst", rateLimitConfig.Burst,
		"window", rateLimitConfig.Window,
		"skip_paths", rateLimitConfig.SkipPaths,
		"skip_methods", rateLimitConfig.SkipMethods,
	)

	return echomw.RateLimiterWithConfig(rl.createConfig(rateLimitConfig))
}

func (rl *RateLimiter) validateConfig(config appconfig.RateLimitConfig) error {
	if config.Requests <= 0 {
		return errors.New("requests per second must be positive")
	}
	if config.Burst <= 0 {
		return errors.New("burst must be positive")
	}
	if config.Window <= 0 {
		return errors.New("window duration must be positive")
	}
	return nil
}

func (rl *RateLimiter) createConfig(config appconfig.RateLimitConfig) echomw.RateLimiterConfig {
	return echomw.RateLimiterConfig{
		Skipper:             rl.createSkipper(config),
		Store:               rl.createStore(config),
		IdentifierExtractor: rl.createIdentifierExtractor(),
		ErrorHandler:        rl.createErrorHandler(),
		DenyHandler:         rl.createDenyHandler(),
	}
}

func (rl *RateLimiter) createSkipper(config appconfig.RateLimitConfig) echomw.Skipper {
	return func(c echo.Context) bool {
		path := c.Request().URL.Path
		method := c.Request().Method

		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				rl.logger.Debug("Rate limiter skipping path", "path", path, "skip_path", skipPath)
				return true
			}
		}

		for _, skipMethod := range config.SkipMethods {
			if method == skipMethod {
				rl.logger.Debug("Rate limiter skipping method", "method", method, "skip_method", skipMethod)
				return true
			}
		}

		return false
	}
}

func (rl *RateLimiter) createStore(config appconfig.RateLimitConfig) echomw.RateLimiterStore {
	return echomw.NewRateLimiterMemoryStoreWithConfig(
		echomw.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(config.Requests),
			Burst:     config.Burst,
			ExpiresIn: config.Window,
		},
	)
}

func (rl *RateLimiter) createIdentifierExtractor() echomw.Extractor {
	return func(c echo.Context) (string, error) {
		path := c.Request().URL.Path

		switch {
		case rl.pathChecker.IsAuthPath(path):
			return fmt.Sprintf("ip:%s", c.RealIP()), nil
		case rl.pathChecker.IsFormPath(path):
			return rl.getFormIdentifier(c), nil
		default:
			return fmt.Sprintf("default:%s", c.RealIP()), nil
		}
	}
}

func (rl *RateLimiter) getFormIdentifier(c echo.Context) string {
	formID := c.Param("formID")
	if formID == "" {
		formID = "unknown"
	}

	origin := c.Request().Header.Get("Origin")
	if origin == "" {
		origin = "unknown"
	}

	return fmt.Sprintf("form:%s:%s", formID, origin)
}

func (rl *RateLimiter) createErrorHandler() func(c echo.Context, err error) error {
	return func(c echo.Context, err error) error {
		rl.logger.Warn("Rate limit exceeded",
			"path", c.Request().URL.Path,
			"method", c.Request().Method,
			"ip", c.RealIP(),
			"error", err,
		)
		return echo.NewHTTPError(http.StatusTooManyRequests, RateLimitExceededMsg)
	}
}

func (rl *RateLimiter) createDenyHandler() func(c echo.Context, identifier string, err error) error {
	return func(c echo.Context, identifier string, err error) error {
		rl.logger.Warn("Rate limit denied",
			"path", c.Request().URL.Path,
			"method", c.Request().Method,
			"ip", c.RealIP(),
			"identifier", identifier,
			"error", err,
		)
		return echo.NewHTTPError(http.StatusTooManyRequests, RateLimitDeniedMsg)
	}
}

func noopMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}
