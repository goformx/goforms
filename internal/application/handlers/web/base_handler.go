package web

import (
	"context"
	"net/http"

	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	Logger         logging.Logger
	Config         *config.Config
	UserService    user.Service
	FormService    form.Service
	Renderer       view.Renderer
	SessionManager *session.Manager
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(
	logger logging.Logger,
	config *config.Config,
	userService user.Service,
	formService form.Service,
	renderer view.Renderer,
	sessionManager *session.Manager,
) *BaseHandler {
	return &BaseHandler{
		Logger:         logger,
		Config:         config,
		UserService:    userService,
		FormService:    formService,
		Renderer:       renderer,
		SessionManager: sessionManager,
	}
}

// RequireAuthenticatedUser ensures the user is authenticated and returns the user object
func (h *BaseHandler) RequireAuthenticatedUser(c echo.Context) (*entities.User, error) {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return nil, c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		h.Logger.Error("failed to get user", "error", err)
		return nil, response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, "Failed to get user")
	}

	return user, nil
}

// BuildPageData creates page data with common fields
func (h *BaseHandler) BuildPageData(c echo.Context, title string) shared.PageData {
	return shared.BuildPageData(h.Config, c, title)
}

// HandleError handles common error scenarios
func (h *BaseHandler) HandleError(c echo.Context, err error, message string) error {
	h.Logger.Error(message, "error", err)
	return response.WebErrorResponse(c, h.Renderer, http.StatusInternalServerError, message)
}

// HandleNotFound handles not found errors
func (h *BaseHandler) HandleNotFound(c echo.Context, message string) error {
	return response.WebErrorResponse(c, h.Renderer, http.StatusNotFound, message)
}

// HandleForbidden handles forbidden access errors
func (h *BaseHandler) HandleForbidden(c echo.Context, message string) error {
	return response.WebErrorResponse(c, h.Renderer, http.StatusForbidden, message)
}

// Start provides default lifecycle initialization
func (h *BaseHandler) Start(ctx context.Context) error {
	// Default implementation - no initialization needed
	return nil
}

// Stop provides default lifecycle cleanup
func (h *BaseHandler) Stop(ctx context.Context) error {
	// Default implementation - no cleanup needed
	return nil
}

// Register provides default route registration
func (h *BaseHandler) Register(e *echo.Echo) {
	// Default implementation - routes registered by RegisterHandlers
}
