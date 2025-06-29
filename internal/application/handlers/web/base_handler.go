package web

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	mwcontext "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/application/response"
	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/view"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	Logger         logging.Logger
	Config         *config.Config
	UserService    user.Service
	FormService    form.Service
	Renderer       view.Renderer
	SessionManager *session.Manager
	ErrorHandler   response.ErrorHandlerInterface
	AssetManager   web.AssetManagerInterface // Use interface instead of concrete type
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(
	logger logging.Logger,
	cfg *config.Config,
	userService user.Service,
	formService form.Service,
	renderer view.Renderer,
	sessionManager *session.Manager,
	errorHandler response.ErrorHandlerInterface,
	assetManager web.AssetManagerInterface, // Use interface
) *BaseHandler {
	return &BaseHandler{
		Logger:         logger,
		Config:         cfg,
		UserService:    userService,
		FormService:    formService,
		Renderer:       renderer,
		SessionManager: sessionManager,
		ErrorHandler:   errorHandler,
		AssetManager:   assetManager,
	}
}

// RequireAuthenticatedUser ensures the user is authenticated and returns the user object
func (h *BaseHandler) RequireAuthenticatedUser(c echo.Context) (*entities.User, error) {
	userID, ok := mwcontext.GetUserID(c)
	if !ok {
		return nil, fmt.Errorf("redirect to login: %w", c.Redirect(constants.StatusSeeOther, constants.PathLogin))
	}

	userEntity, err := h.UserService.GetUserByID(c.Request().Context(), userID)
	if err != nil || userEntity == nil {
		h.Logger.Error("failed to get user", "error", err)

		return nil, h.HandleError(c, err, "Failed to get user")
	}

	return userEntity, nil
}

// NewPageData creates page data with common fields
func (h *BaseHandler) NewPageData(c echo.Context, title string) *view.PageData {
	return view.NewPageData(h.Config, h.AssetManager, c, title)
}

// HandleError handles common error scenarios
func (h *BaseHandler) HandleError(c echo.Context, err error, message string) error {
	// Use the error handler for sanitized logging instead of logging raw error
	if handleErr := h.ErrorHandler.HandleError(err, c, message); handleErr != nil {
		return fmt.Errorf("handle error: %w", handleErr)
	}

	return nil
}

// HandleNotFound handles not found errors
func (h *BaseHandler) HandleNotFound(c echo.Context, message string) error {
	if notFoundErr := h.ErrorHandler.HandleNotFoundError(message, c); notFoundErr != nil {
		return fmt.Errorf("handle not found error: %w", notFoundErr)
	}

	return nil
}

// HandleForbidden handles forbidden access errors
func (h *BaseHandler) HandleForbidden(c echo.Context, message string) error {
	if forbiddenErr := h.ErrorHandler.HandleDomainError(
		domainerrors.New(domainerrors.ErrCodeForbidden, message, nil), c,
	); forbiddenErr != nil {
		return fmt.Errorf("handle forbidden error: %w", forbiddenErr)
	}

	return nil
}

// GetAssetBaseURL returns the base URL for assets (convenience method)
func (h *BaseHandler) GetAssetBaseURL() string {
	return h.AssetManager.GetBaseURL()
}

// ValidateAssetPath validates an asset path (convenience method)
func (h *BaseHandler) ValidateAssetPath(path string) error {
	if err := h.AssetManager.ValidatePath(path); err != nil {
		return fmt.Errorf("validate asset path: %w", err)
	}

	return nil
}

// Start initializes the base handler.
// This is called during application startup.
func (h *BaseHandler) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the base handler.
// This is called during application shutdown.
func (h *BaseHandler) Stop(_ context.Context) error {
	return nil // No cleanup needed
}

// Register provides default route registration
func (h *BaseHandler) Register(_ *echo.Echo) {
	// Default implementation - routes registered by RegisterHandlers
}
