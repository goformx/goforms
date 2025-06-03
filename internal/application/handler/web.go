package handler

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"

	amw "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/view"
)

// WebHandler handles web requests
type WebHandler struct {
	*handlers.BaseHandler
	renderer          *view.Renderer
	middlewareManager *amw.Manager
	config            *config.Config
	userService       user.Service
	sessionManager    *amw.SessionManager
	authHandler       *AuthHandler
	pageHandler       *PageHandler
}

// NewWebHandler creates a new web handler
func NewWebHandler(
	baseHandler *handlers.BaseHandler,
	userService user.Service,
	sessionManager *amw.SessionManager,
	renderer *view.Renderer,
	middlewareManager *amw.Manager,
	config *config.Config,
) *WebHandler {
	handler := &WebHandler{
		BaseHandler:       baseHandler,
		userService:       userService,
		sessionManager:    sessionManager,
		renderer:          renderer,
		middlewareManager: middlewareManager,
		config:            config,
	}

	// Initialize sub-handlers
	handler.authHandler = NewAuthHandler(
		baseHandler,
		userService,
		sessionManager,
		renderer,
		middlewareManager,
		config,
	)
	handler.pageHandler = NewPageHandler(
		baseHandler,
		userService,
		sessionManager,
		renderer,
		middlewareManager,
		config,
	)

	return handler
}

// Validate validates that required dependencies are set.
// Returns an error if any required dependency is missing.
//
// Required dependencies:
//   - renderer
//   - middlewareManager
//   - config
func (h *WebHandler) Validate() error {
	if err := h.BaseHandler.Validate(); err != nil {
		return fmt.Errorf("WebHandler validation failed: %w", err)
	}
	if h.renderer == nil {
		return errors.New("WebHandler validation failed: renderer is required")
	}
	if h.middlewareManager == nil {
		return errors.New("WebHandler validation failed: middleware manager is required")
	}
	if h.config == nil {
		return errors.New("WebHandler validation failed: config is required")
	}
	return nil
}

// registerRoutes registers the web routes
func (h *WebHandler) registerRoutes(e *echo.Echo) {
	// Register sub-handler routes
	h.authHandler.Register(e)
	h.pageHandler.Register(e)
}

// validateDependencies validates required dependencies for the handler
func (h *WebHandler) validateDependencies() {
	if err := h.Validate(); err != nil {
		h.LogError("failed to validate web handler", err)
		panic(fmt.Sprintf("failed to validate web handler: %v", err))
	}
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	h.validateDependencies()
	if h.config.App.IsDevelopment() {
		h.LogDebug("registering web routes")
	}

	h.middlewareManager.Setup(e) // Ensure middleware is loaded properly
	h.registerRoutes(e)

	if h.config.App.IsDevelopment() {
		h.LogDebug("web routes registration complete")
	}
}
