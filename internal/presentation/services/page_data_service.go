package services

import (
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/web"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
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
func (s *PageDataService) PrepareDashboardData(c echo.Context, user *user.User, forms []*form.Form) shared.PageData {
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return shared.PageData{
		Title:     "Dashboard - GoForms",
		User:      user,
		Forms:     forms,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}
}

// PrepareFormData prepares data for form-related pages
func (s *PageDataService) PrepareFormData(c echo.Context, user *user.User, form *form.Form, submissions []*model.FormSubmission) shared.PageData {
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return shared.PageData{
		Title:                "Edit Form - GoForms",
		User:                 user,
		Form:                 form,
		Submissions:          submissions,
		CSRFToken:            csrfToken,
		AssetPath:            web.GetAssetPath,
		FormBuilderAssetPath: web.GetAssetPath("src/js/form-builder.ts"),
	}
}

// PrepareNewFormData prepares data for the new form page
func (s *PageDataService) PrepareNewFormData(c echo.Context, user *user.User) shared.PageData {
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	return shared.PageData{
		Title:     "Create New Form - GoForms",
		User:      user,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}
}
