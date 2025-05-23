package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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
	submissions.Use(h.Base.authMiddleware.RequireAuth)

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

	return c.Render(http.StatusOK, "form_submissions.html", map[string]any{
		"Form":        formObj,
		"Submissions": submissions,
	})
}
