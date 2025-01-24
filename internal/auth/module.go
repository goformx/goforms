package auth

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	v1 "github.com/jonesrussell/goforms/internal/api/v1"
	"github.com/jonesrussell/goforms/internal/core/user"
	"github.com/jonesrussell/goforms/internal/middleware"
	userstore "github.com/jonesrussell/goforms/internal/store/user"
)

//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Module("auth",
	fx.Provide(
		userstore.NewStore,
		user.NewService,
		v1.NewAuthHandler,
		middleware.NewJWTMiddleware,
	),
	fx.Invoke(registerAuthRoutes),
)

// AuthParams contains the dependencies needed for auth routes
type AuthParams struct {
	fx.In

	Echo          *echo.Echo
	AuthHandler   *v1.AuthHandler
	JWTMiddleware echo.MiddlewareFunc `name:"jwt"`
}

// registerAuthRoutes sets up the authentication routes
func registerAuthRoutes(p AuthParams) {
	// Apply JWT middleware to all routes except auth routes
	p.Echo.Use(p.JWTMiddleware)

	// Register auth routes
	p.AuthHandler.RegisterRoutes(p.Echo)
}
