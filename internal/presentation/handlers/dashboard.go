package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/presentation/middleware"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	authMiddleware *middleware.AuthMiddleware
}

func NewDashboardHandler(userService user.Service) *DashboardHandler {
	return &DashboardHandler{
		authMiddleware: middleware.NewAuthMiddleware(userService),
	}
}

func (h *DashboardHandler) RegisterRoutes(e *echo.Echo) {
	// Dashboard routes
	dashboard := e.Group("/dashboard")
	dashboard.Use(h.authMiddleware.RequireAuth) // Middleware to ensure user is authenticated

	dashboard.GET("", h.ShowDashboard)
}

func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	// Get user from context (set by auth middleware)
	user, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	// Create page data
	data := shared.PageData{
		Title: "Dashboard - GoForms",
		User:  user,
	}

	// Render dashboard page
	return pages.Dashboard(data).Render(c.Request().Context(), c.Response().Writer)
} 