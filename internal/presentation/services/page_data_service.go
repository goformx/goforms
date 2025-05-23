package services

import (
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// PageDataService handles template data preparation
type PageDataService struct {
	logger logging.Logger
}

// NewPageDataService creates a new page data service
func NewPageDataService(logger logging.Logger) *PageDataService {
	return &PageDataService{
		logger: logger,
	}
}

// PrepareDashboardData prepares data for the dashboard page
func (s *PageDataService) PrepareDashboardData(
	c echo.Context,
	currentUser *user.User,
	forms []*form.Form,
) shared.PageData {
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return shared.PageData{
		Title:     "Dashboard - GoFormX",
		User:      currentUser,
		Forms:     forms,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}
}

// PrepareFormData prepares data for form-related pages
func (s *PageDataService) PrepareFormData(
	c echo.Context,
	currentUser *user.User,
	formObj *form.Form,
	submissions []*model.FormSubmission,
) shared.PageData {
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return shared.PageData{
		Title:                "Edit Form - GoFormX",
		User:                 currentUser,
		Form:                 formObj,
		Submissions:          submissions,
		CSRFToken:            csrfToken,
		AssetPath:            web.GetAssetPath,
		FormBuilderAssetPath: web.GetAssetPath("src/js/form-builder.ts"),
	}
}

// PrepareNewFormData prepares data for the new form page
func (s *PageDataService) PrepareNewFormData(c echo.Context, currentUser *user.User) shared.PageData {
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return shared.PageData{
		Title:     "Create New Form - GoFormX",
		User:      currentUser,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}
}
