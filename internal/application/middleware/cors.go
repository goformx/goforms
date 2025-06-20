package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
)

// CORS creates and configures CORS middleware using Echo's built-in implementation
func CORS(securityConfig *appconfig.SecurityConfig) echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     securityConfig.CORS.AllowedOrigins,
		AllowMethods:     securityConfig.CORS.AllowedMethods,
		AllowHeaders:     securityConfig.CORS.AllowedHeaders,
		AllowCredentials: securityConfig.CORS.AllowCredentials,
		MaxAge:           securityConfig.CORS.MaxAge,
	})
}
