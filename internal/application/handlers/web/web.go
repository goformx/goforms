package web

import (
	"errors"
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// WebHandler handles web page requests
type WebHandler struct {
	baseHandler       *BaseHandler
	userService       domain.UserService
	sessionManager    *middleware.SessionManager
	renderer          *view.Renderer
	middlewareManager *middleware.Manager
	config            *config.Config
	logger            logging.Logger
}

// NewWebHandler creates a new web handler
func NewWebHandler(
	baseHandler *BaseHandler,
	userService domain.UserService,
	sessionManager *middleware.SessionManager,
	renderer *view.Renderer,
	middlewareManager *middleware.Manager,
	cfg *config.Config,
	logger logging.Logger,
) *WebHandler {
	return &WebHandler{
		baseHandler:       baseHandler,
		userService:       userService,
		sessionManager:    sessionManager,
		renderer:          renderer,
		middlewareManager: middlewareManager,
		config:            cfg,
		logger:            logger,
	}
}

// Validate validates the handler's dependencies
func (h *WebHandler) Validate() error {
	if h.baseHandler == nil {
		return errors.New("base handler is required")
	}
	if h.userService == nil {
		return errors.New("user service is required")
	}
	if h.sessionManager == nil {
		return errors.New("session manager is required")
	}
	if h.renderer == nil {
		return errors.New("renderer is required")
	}
	if h.middlewareManager == nil {
		return errors.New("middleware manager is required")
	}
	if h.config == nil {
		return errors.New("config is required")
	}
	if h.logger == nil {
		return errors.New("logger is required")
	}
	return nil
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	e.GET("/", h.handleHome)
	e.GET("/dashboard", h.handleDashboard)
	e.GET("/forms/:id", h.handleFormView)
}

// handleHome handles the home page request
func (h *WebHandler) handleHome(c echo.Context) error {
	data := shared.PageData{
		Title: "Welcome to GoFormX",
	}
	return h.renderer.Render(c, pages.Home(data))
}

// handleDashboard handles the dashboard page request
func (h *WebHandler) handleDashboard(c echo.Context) error {
	// Get user ID from session
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	userID := userIDRaw

	// Get user's forms
	forms, err := h.baseHandler.formService.GetUserForms(userID)
	if err != nil {
		h.logger.Error("failed to get user forms", logging.ErrorField("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get forms",
		})
	}

	data := shared.PageData{
		Title: "Dashboard",
		Forms: forms,
	}
	return h.renderer.Render(c, pages.Dashboard(data))
}

// handleFormView handles the form view page request
func (h *WebHandler) handleFormView(c echo.Context) error {
	formID := c.Param("id")
	if formID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Form ID is required",
		})
	}

	form, err := h.baseHandler.formService.GetForm(formID)
	if err != nil {
		h.logger.Error("failed to get form", logging.ErrorField("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get form",
		})
	}

	data := shared.PageData{
		Title: form.Title,
		Form:  form,
	}
	return h.renderer.Render(c, pages.Forms(data))
}
