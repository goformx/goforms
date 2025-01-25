package http

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Handlers groups all HTTP handlers
type Handlers struct {
	logger              logging.Logger
	contactService      contact.Service
	subscriptionService subscription.Service
}

// NewHandlers creates all HTTP handlers
func NewHandlers(
	logger logging.Logger,
	contactService contact.Service,
	subscriptionService subscription.Service,
) *Handlers {
	return &Handlers{
		logger:              logger,
		contactService:      contactService,
		subscriptionService: subscriptionService,
	}
}

// Register registers all routes
func (h *Handlers) Register(e *echo.Echo) {
	// API routes
	api := e.Group("/api/v1")
	h.registerContact(api)
	h.registerSubscription(api)

	// Web routes
	e.GET("/", h.handleHome)
	e.GET("/contact", h.handleContact)
	// ... other web routes
}
