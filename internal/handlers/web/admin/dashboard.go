package admin

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

const (
	// UnauthorizedErrorCode is the HTTP status code for unauthorized access
	UnauthorizedErrorCode = http.StatusUnauthorized
	// InternalServerErrorCode is the HTTP status code for internal server errors
	InternalServerErrorCode = http.StatusInternalServerError
)

// AdminDashboardHandler handles the admin dashboard routes
type AdminDashboardHandler struct {
	*handlers.BaseHandler
	renderer    *view.Renderer
	UserService user.Service
	FormService form.Service
}

// NewAdminDashboardHandler creates a new AdminDashboardHandler
func NewAdminDashboardHandler(
	logger logging.Logger,
	renderer *view.Renderer,
	userService user.Service,
	formService form.Service,
) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		BaseHandler: handlers.NewBaseHandler(formService, logger),
		renderer:    renderer,
		UserService: userService,
		FormService: formService,
	}
}

// Register sets up the routes for the admin dashboard handler
func (h *AdminDashboardHandler) Register(e *echo.Echo) {
	e.GET("/dashboard", h.showDashboard)
}

// showDashboard renders the admin dashboard page
func (h *AdminDashboardHandler) showDashboard(c echo.Context) error {
	h.LogDebug("handling admin dashboard page request")

	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(UnauthorizedErrorCode, "User not found")
	}

	forms, err := h.FormService.GetUserForms(currentUser.ID)
	if err != nil {
		h.LogError("failed to get user forms", err)
		return echo.NewHTTPError(InternalServerErrorCode, "Failed to get forms")
	}

	data := shared.PageData{
		Title: "Dashboard - GoFormX",
		User:  currentUser,
		Forms: forms,
	}

	return h.renderer.Render(c, pages.Dashboard(data))
}
