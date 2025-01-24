package application

import (
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/http"
	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
	"github.com/jonesrussell/goforms/internal/application/server"
	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// NewEcho creates a new Echo instance with common middleware and routes
func NewEcho(log logging.Logger, contactAPI *v1.ContactAPI, subscriptionAPI *v1.SubscriptionAPI, webHandler *v1.WebHandler) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Register API routes
	contactAPI.Register(e)
	subscriptionAPI.Register(e)
	webHandler.Register(e)

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

// Module provides application dependencies
//
//nolint:gochecknoglobals // fx modules are designed to be global
var Module = fx.Options(
	fx.Provide(
		NewEcho,
		server.New,
		NewServerConfig,
	),
	http.Module,
	domain.Module,
)
