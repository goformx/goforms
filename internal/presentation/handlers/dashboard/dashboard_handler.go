package dashboard

import (
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// DashboardHandler handles the /dashboard route
// Implements httpiface.Handler
type DashboardHandler struct {
	handlers.BaseHandler
	formService     form.Service
	sessionManager  *session.Manager
	renderer        view.Renderer
	config          *config.Config
	assetManager    web.AssetManagerInterface
	logger          logging.Logger
	responseBuilder *DashboardResponseBuilder
}

// NewDashboardHandler creates a new DashboardHandler and registers the /dashboard route
func NewDashboardHandler(
	formService form.Service,
	sessionManager *session.Manager,
	renderer view.Renderer,
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	logger logging.Logger,
) *DashboardHandler {
	h := &DashboardHandler{
		BaseHandler:     *handlers.NewBaseHandler("dashboard"),
		formService:     formService,
		sessionManager:  sessionManager,
		renderer:        renderer,
		config:          cfg,
		assetManager:    assetManager,
		logger:          logger,
		responseBuilder: NewDashboardResponseBuilder(cfg, assetManager, renderer, logger),
	}

	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/dashboard",
		Handler: h.Dashboard,
	})

	return h
}

// Dashboard handles GET /dashboard
func (h *DashboardHandler) Dashboard(ctx httpiface.Context) error {
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	user, err := h.getUserFromSession(echoCtx)
	if err != nil {
		h.logger.Warn("authentication required for dashboard access", "error", err)

		return h.responseBuilder.BuildAuthenticationErrorResponse(echoCtx)
	}

	forms, err := h.formService.ListForms(echoCtx.Request().Context(), user.ID)
	if err != nil {
		h.logger.Error("failed to fetch user forms", "user_id", user.ID, "error", err)

		return h.responseBuilder.BuildDashboardErrorResponse(echoCtx, "Failed to load your forms. Please try again.")
	}

	return h.responseBuilder.BuildDashboardResponse(echoCtx, user, forms)
}

// getUserFromSession extracts user information from the session
func (h *DashboardHandler) getUserFromSession(c echo.Context) (*entities.User, error) {
	// Get session cookie
	cookie, err := c.Cookie(h.sessionManager.GetCookieName())
	if err != nil {
		return nil, fmt.Errorf("no session cookie found")
	}

	// Get session from manager
	sess, exists := h.sessionManager.GetSession(cookie.Value)
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session is expired
	if time.Now().After(sess.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Create user entity from session data
	user := &entities.User{
		ID:    sess.UserID,
		Email: sess.Email,
		Role:  sess.Role,
	}

	return user, nil
}
