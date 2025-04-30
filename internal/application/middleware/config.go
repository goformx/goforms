package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Config holds middleware configuration
type Config struct {
	Logger      logging.Logger
	JWTSecret   string
	UserService user.Service
	EnableCSRF  bool
}

// Setup configures all middleware for an Echo instance
func Setup(e *echo.Echo, cfg *Config) {
	// Security
	m := New(&ManagerConfig{
		Logger:      cfg.Logger,
		JWTSecret:   cfg.JWTSecret,
		UserService: cfg.UserService,
		EnableCSRF:  cfg.EnableCSRF,
	})
	m.Setup(e)

	// Logging must be first to capture all requests
	e.Use(LoggingMiddleware(cfg.Logger))

	// Auth if secret provided
	if cfg.JWTSecret != "" && cfg.UserService != nil {
		middleware, err := NewJWTMiddleware(cfg.UserService, cfg.JWTSecret)
		if err != nil {
			cfg.Logger.Error("failed to create JWT middleware", logging.Error(err))
			return
		}
		e.Use(middleware)
	}
}
