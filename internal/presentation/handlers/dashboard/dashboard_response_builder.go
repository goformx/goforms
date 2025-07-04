package dashboard

import (
	"net/http"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// DashboardResponseBuilder builds dashboard responses
type DashboardResponseBuilder struct {
	config       *config.Config
	assetManager web.AssetManagerInterface
	renderer     view.Renderer
	logger       logging.Logger
}

// NewDashboardResponseBuilder creates a new dashboard response builder
func NewDashboardResponseBuilder(
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	renderer view.Renderer,
	logger logging.Logger,
) *DashboardResponseBuilder {
	return &DashboardResponseBuilder{
		config:       cfg,
		assetManager: assetManager,
		renderer:     renderer,
		logger:       logger,
	}
}

// BuildDashboardResponse builds a successful dashboard response
func (b *DashboardResponseBuilder) BuildDashboardResponse(
	c echo.Context,
	user *entities.User,
	forms []*model.Form,
) error {
	b.logger.Info("dashboard rendered successfully", "user_id", user.ID, "forms_count", len(forms))

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    user.ID,
					"email": user.Email,
					"role":  user.Role,
				},
				"forms": forms,
				"count": len(forms),
			},
		})
	}

	// For web requests, render dashboard page
	return b.renderDashboard(c, user, forms)
}

// BuildDashboardErrorResponse builds an error response for dashboard failures
func (b *DashboardResponseBuilder) BuildDashboardErrorResponse(c echo.Context, errorMessage string) error {
	b.logger.Warn("dashboard error", "error", errorMessage)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": errorMessage,
		})
	}

	// For web requests, render error page or redirect to login if auth error
	if strings.Contains(errorMessage, "authentication") || strings.Contains(errorMessage, "unauthorized") {
		return c.Redirect(http.StatusSeeOther, constants.PathLogin)
	}

	// Render dashboard with error message
	return b.renderDashboardWithError(c, errorMessage)
}

// BuildAuthenticationErrorResponse builds an authentication error response
func (b *DashboardResponseBuilder) BuildAuthenticationErrorResponse(c echo.Context) error {
	b.logger.Warn("authentication required for dashboard access")

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status":  "error",
			"message": "Authentication required",
		})
	}

	// For web requests, redirect to login
	return c.Redirect(http.StatusSeeOther, constants.PathLogin)
}

// BuildFormsListResponse builds a response for forms list (API endpoint)
func (b *DashboardResponseBuilder) BuildFormsListResponse(
	c echo.Context,
	forms []*model.Form,
	total int,
	page,
	limit int,
) error {
	b.logger.Info("forms list retrieved", "count", len(forms), "total", total, "page", page, "limit", limit)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"forms": forms,
			"pagination": map[string]interface{}{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"totalPages": (total + limit - 1) / limit,
			},
		},
	})
}

// BuildFormsErrorResponse builds an error response for forms operations
func (b *DashboardResponseBuilder) BuildFormsErrorResponse(c echo.Context, errorMessage string, statusCode int) error {
	b.logger.Warn("forms operation failed", "error", errorMessage, "status_code", statusCode)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(statusCode, map[string]interface{}{
			"status":  "error",
			"message": errorMessage,
		})
	}

	// For web requests, render dashboard with error
	return b.renderDashboardWithError(c, errorMessage)
}

// Helper methods

// isAPIRequest checks if the request is an API request
func (b *DashboardResponseBuilder) isAPIRequest(c echo.Context) bool {
	accept := c.Request().Header.Get("Accept")
	contentType := c.Request().Header.Get("Content-Type")

	return strings.Contains(accept, "application/json") ||
		strings.Contains(contentType, "application/json") ||
		strings.HasPrefix(c.Request().URL.Path, "/api/")
}

// renderDashboard renders the dashboard page with user and forms data
func (b *DashboardResponseBuilder) renderDashboard(c echo.Context, user *entities.User, forms []*model.Form) error {
	// Create page data for the dashboard template
	pageData := view.NewPageData(b.config, b.assetManager, c, "Dashboard")
	pageData = pageData.WithUser(user).WithForms(forms)

	// Render the dashboard page using the generated template
	dashboardComponent := pages.Dashboard(*pageData, forms)

	return b.renderer.Render(c, dashboardComponent)
}

// renderDashboardWithError renders the dashboard page with an error message
func (b *DashboardResponseBuilder) renderDashboardWithError(c echo.Context, errorMessage string) error {
	// Create page data for the dashboard template
	pageData := view.NewPageData(b.config, b.assetManager, c, "Dashboard")
	pageData = pageData.WithMessage("error", errorMessage)

	// For now, render with empty forms list - in a real implementation,
	// you might want to try to get the user and forms again, or show a different error page
	dashboardComponent := pages.Dashboard(*pageData, []*model.Form{})

	return b.renderer.Render(c, dashboardComponent)
}
