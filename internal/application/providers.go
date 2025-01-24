package application

import (
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo, handlers ...interface{ Register(e *echo.Echo) }) {
	for _, handler := range handlers {
		handler.Register(e)
	}
}

// NewEcho creates a new Echo instance with common middleware and routes
func NewEcho(log logging.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

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
