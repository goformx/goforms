package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// WebHandlerOption defines a web handler option.
// This type is used to implement the functional options pattern
// for configuring the WebHandler.
type WebHandlerOption func(*WebHandler)

// WithContactService sets the contact service.
// This is a required option for the WebHandler as it needs
// the contact service to function properly.
//
// Example:
//
//	handler := NewWebHandler(logger, WithContactService(contactService))
func WithContactService(svc contact.Service) WebHandlerOption {
	return func(h *WebHandler) {
		h.contactService = svc
	}
}

// WithRenderer sets the view renderer.
// This is a required option for the WebHandler as it needs
// the renderer to display web pages.
//
// Example:
//
//	handler := NewWebHandler(logger, WithRenderer(renderer))
func WithRenderer(renderer *view.Renderer) WebHandlerOption {
	return func(h *WebHandler) {
		h.renderer = renderer
	}
}

// WebHandler handles web page requests.
// It requires both a renderer and a contact service to function properly.
// Use the functional options pattern to configure these dependencies.
type WebHandler struct {
	Base
	contactService contact.Service
	renderer       *view.Renderer
}

// NewWebHandler creates a new web handler.
// It uses the functional options pattern to configure the handler.
// The logger is required as a direct parameter, while other dependencies
// are provided through options.
//
// Example:
//
//	handler := NewWebHandler(logger,
//	    WithRenderer(renderer),
//	    WithContactService(contactService),
//	)
func NewWebHandler(logger logging.Logger, opts ...WebHandlerOption) *WebHandler {
	h := &WebHandler{
		Base: Base{Logger: logger},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Validate validates that required dependencies are set.
// This ensures that all required dependencies have been properly
// configured through the functional options pattern.
func (h *WebHandler) Validate() error {
	if err := h.Base.Validate(); err != nil {
		return err
	}
	if h.renderer == nil {
		return fmt.Errorf("renderer is required")
	}
	if h.contactService == nil {
		return fmt.Errorf("contact service is required")
	}
	return nil
}

// Register registers the web routes.
// This method sets up all web page routes and static file serving.
// It validates that all required dependencies are available before
// registering any routes.
func (h *WebHandler) Register(e *echo.Echo) {
	if err := h.Validate(); err != nil {
		h.Logger.Error("failed to validate handler", logging.Error(err))
		return
	}

	h.Logger.Debug("registering web routes")

	// Web pages
	e.GET("/", h.handleHome)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/"))

	e.GET("/contact", h.handleContact)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/contact"))

	e.GET("/signup", h.handleSignup)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/signup"))

	e.GET("/login", h.handleLogin)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/login"))

	// Static files - Note: paths must be relative to the project root
	e.Static("/static", "./static")
	h.Logger.Debug("registered static directory", logging.String("path", "/static"), logging.String("root", "./static"))

	e.File("/favicon.ico", "./static/favicon.ico")
	h.Logger.Debug("registered favicon", logging.String("path", "/favicon.ico"))

	h.Logger.Debug("web routes registration complete")
}

// handleHome renders the home page
func (h *WebHandler) handleHome(c echo.Context) error {
	h.Logger.Debug("handling home page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	err := h.renderer.Render(c, pages.Home())
	if err != nil {
		h.Logger.Error("failed to render home page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
	}
	return err
}

// handleContact renders the contact page
func (h *WebHandler) handleContact(c echo.Context) error {
	h.Logger.Debug("handling contact page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	err := h.renderer.Render(c, pages.Contact())
	if err != nil {
		h.Logger.Error("failed to render contact page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
	}
	return err
}

// handleSignup renders the signup page
func (h *WebHandler) handleSignup(c echo.Context) error {
	h.Logger.Debug("handling signup page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	err := h.renderer.Render(c, pages.Signup())
	if err != nil {
		h.Logger.Error("failed to render signup page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
	}
	return err
}

// handleLogin renders the login page
func (h *WebHandler) handleLogin(c echo.Context) error {
	h.Logger.Debug("handling login page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	err := h.renderer.Render(c, pages.Login())
	if err != nil {
		h.Logger.Error("failed to render login page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
	}
	return err
}
