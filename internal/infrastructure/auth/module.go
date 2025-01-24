package auth

import (
	"os"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/user"
	userstore "github.com/jonesrussell/goforms/internal/infrastructure/store"
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

// Result bundles auth-related dependencies for injection
type Result struct {
	fx.Out

	JWTMiddleware echo.MiddlewareFunc
}

// New creates auth-related dependencies
func New(cfg *Config, userService user.Service) Result {
	return Result{
		JWTMiddleware: provideJWTMiddleware(cfg, userService),
	}
}

//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Module("auth",
	fx.Provide(
		NewConfig,
		userstore.NewStore,
		New,
		v1.NewAuthHandler,
	),
	fx.Invoke(registerAuthRoutes),
)

// provideJWTMiddleware creates a new JWT middleware with configuration
func provideJWTMiddleware(cfg *Config, userService user.Service) echo.MiddlewareFunc {
	return middleware.NewJWTMiddleware(userService, cfg.JWTSecret)
}

// Params for route registration
type Params struct {
	fx.In

	Echo          *echo.Echo
	AuthHandler   *v1.AuthHandler
	JWTMiddleware echo.MiddlewareFunc
}

// registerAuthRoutes registers authentication routes
func registerAuthRoutes(p Params) {
	// Apply JWT middleware to all routes except auth routes
	p.Echo.Use(p.JWTMiddleware)

	// Register auth routes
	p.AuthHandler.RegisterRoutes(p.Echo)
}
