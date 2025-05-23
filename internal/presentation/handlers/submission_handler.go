package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/web"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// SubmissionHandler handles form submission-related HTTP requests
type SubmissionHandler struct {
	formService form.Service
	logger      logging.Logger
	Base        *BaseHandler
}

// NewSubmissionHandler creates a new submission handler
func NewSubmissionHandler(
	formService form.Service,
	logger logging.Logger,
	base *BaseHandler,
) *SubmissionHandler {
	return &SubmissionHandler{
		formService: formService,
		logger:      logger,
		Base:        base,
	}
}

// Register sets up the submission routes
func (h *SubmissionHandler) Register(e *echo.Echo) {
	submissions := e.Group("/dashboard/forms/:id/submissions")
	h.Base.SetupMiddleware(submissions)

	submissions.GET("", h.ShowFormSubmissions)
}

// ShowFormSubmissions handles viewing form submissions
func (h *SubmissionHandler) ShowFormSubmissions(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.Base.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	submissions, err := h.formService.GetFormSubmissions(formObj.ID)
	if err != nil {
		return h.Base.handleError(err, http.StatusInternalServerError, "Failed to get form submissions")
	}

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	// Create page data
	data := shared.PageData{
		Title:       "Form Submissions - GoForms",
		User:        currentUser,
		Form:        formObj,
		Submissions: submissions,
		CSRFToken:   csrfToken,
		AssetPath:   web.GetAssetPath,
	}

	// Set content
	data.Content = pages.FormSubmissionsContent(data)

	// Render the submissions page
	return pages.FormSubmissions(data).Render(c.Request().Context(), c.Response().Writer)
}
