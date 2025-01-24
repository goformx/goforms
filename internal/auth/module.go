package auth

import (
	"os"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	v1 "github.com/jonesrussell/goforms/internal/api/v1"
	"github.com/jonesrussell/goforms/internal/core/user"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/middleware"
	"github.com/jonesrussell/goforms/internal/models"
	userstore "github.com/jonesrussell/goforms/internal/store/user"
)

// Config holds authentication configuration
type Config struct {
	JWTSecret string
}

// NewConfig creates a new auth configuration
func NewConfig() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-256-bit-secret" // Default for development
	}

	return &Config{
		JWTSecret: jwtSecret,
	}, nil
}

//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Module("auth",
	fx.Provide(
		NewConfig,
		userstore.NewStore,
		provideUserService,
		provideJWTMiddleware,
		v1.NewAuthHandler,
	),
	fx.Invoke(registerAuthRoutes),
)

// provideUserService creates a new user service with JWT configuration
func provideUserService(cfg *Config, log logger.Logger, store models.UserStore) user.Service {
	return user.NewService(log, store, cfg.JWTSecret)
}

// provideJWTMiddleware creates a new JWT middleware with configuration
func provideJWTMiddleware(cfg *Config, userService user.Service) echo.MiddlewareFunc {
	return middleware.NewJWTMiddleware(userService, cfg.JWTSecret)
}

// AuthParams contains the dependencies needed for auth routes
type AuthParams struct {
	fx.In

	Echo          *echo.Echo
	AuthHandler   *v1.AuthHandler
	JWTMiddleware echo.MiddlewareFunc
}

// registerAuthRoutes sets up the authentication routes
func registerAuthRoutes(p AuthParams) {
	// Apply JWT middleware to all routes except auth routes
	p.Echo.Use(p.JWTMiddleware)

	// Register auth routes
	p.AuthHandler.RegisterRoutes(p.Echo)
}
