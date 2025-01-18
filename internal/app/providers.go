package app

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/api"
	v1 "github.com/jonesrussell/goforms/internal/api/v1"
	"github.com/jonesrussell/goforms/internal/config/server"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/web"
)

// NewEcho creates a new Echo instance with common middleware and routes
func NewEcho(log logger.Logger, contactAPI *v1.ContactAPI, subscriptionAPI *v1.SubscriptionAPI, pageHandler *web.PageHandler) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Configure static file serving with proper caching and security
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")

	// Add cache control headers for static files
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Path(), "/static") || c.Path() == "/favicon.ico" {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000")
			}
			return next(c)
		}
	})

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())
	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())

	// Register API routes
	contactAPI.Register(e)
	subscriptionAPI.Register(e)

	// Register web routes
	pageHandler.RegisterRoutes(e)

	return e
}

// NewServerConfig creates a new server configuration
func NewServerConfig() *server.Config {
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		port = 8090 // default port
	}

	readTimeout, err := time.ParseDuration(os.Getenv("READ_TIMEOUT"))
	if err != nil {
		readTimeout = 5 * time.Second
	}

	writeTimeout, err := time.ParseDuration(os.Getenv("WRITE_TIMEOUT"))
	if err != nil {
		writeTimeout = 10 * time.Second
	}

	idleTimeout, err := time.ParseDuration(os.Getenv("IDLE_TIMEOUT"))
	if err != nil {
		idleTimeout = 120 * time.Second
	}

	return &server.Config{
		Host: os.Getenv("SERVER_HOST"),
		Port: port,
		Timeouts: server.TimeoutConfig{
			Read:  readTimeout,
			Write: writeTimeout,
			Idle:  idleTimeout,
		},
	}
}

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		NewEcho,
		NewServerConfig,
		v1.NewContactAPI,
		v1.NewSubscriptionAPI,
	),
	api.Module,
	web.Module,
)
