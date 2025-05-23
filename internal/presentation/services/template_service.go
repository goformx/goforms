package services

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// TemplateService handles template rendering
type TemplateService struct {
	logger logging.Logger
}

// NewTemplateService creates a new template service
func NewTemplateService(logger logging.Logger) *TemplateService {
	return &TemplateService{
		logger: logger,
	}
}

// RenderDashboard renders the dashboard page
func (s *TemplateService) RenderDashboard(c echo.Context, data shared.PageData) error {
	data.Content = pages.DashboardContent(data)
	return pages.Dashboard(data).Render(c.Request().Context(), c.Response().Writer)
}

// RenderNewForm renders the new form page
func (s *TemplateService) RenderNewForm(c echo.Context, data shared.PageData) error {
	data.Content = pages.NewFormContent(data)
	return pages.NewForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// RenderEditForm renders the edit form page
func (s *TemplateService) RenderEditForm(c echo.Context, data shared.PageData) error {
	return pages.EditForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// RenderFormSubmissions renders the form submissions page
func (s *TemplateService) RenderFormSubmissions(c echo.Context, form *form.Form, submissions []*model.FormSubmission) error {
	return c.Render(http.StatusOK, "form_submissions.html", map[string]any{
		"Form":        form,
		"Submissions": submissions,
	})
}
