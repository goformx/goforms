// Package server provides HTTP server interfaces and implementations.
package server

import (
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
)

// ServerInterface defines the interface for HTTP server operations
//
//go:generate mockgen -typed -source=interface.go -destination=../../../test/mocks/server/mock_server.go -package=server
type ServerInterface interface {
	// Start starts the server and returns when it's ready to accept connections
	Start() error

	// URL returns the server's full HTTP URL
	URL() string

	// Echo returns the underlying echo instance
	Echo() *echo.Echo

	// Config returns the server configuration
	Config() *config.Config
}
