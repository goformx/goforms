package handlers

import (
	"errors"
	"net/http"

	"github.com/goformx/goforms/internal/application/services/formops"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// FormHandler handles form-related HTTP requests
type FormHandler struct {
	Base           *BaseHandler
	formService    form.Service
	formOperations formops.Service
	logger         logging.Logger
}

// NewFormHandler creates a new form handler
func NewFormHandler(
	formService form.Service,
	formOperations formops.Service,
	logger logging.Logger,
	base *BaseHandler,
) *FormHandler {
	return &FormHandler{
		formService:    formService,
		formOperations: formOperations,
		logger:         logger,
		Base:           base,
	}
}

// Register sets up the form routes
func (h *FormHandler) Register(e *echo.Echo) {
	forms := e.Group("/dashboard/forms")

	forms.GET("/new", h.ShowNewForm)
	forms.POST("", h.CreateForm)
	forms.GET("/:id/edit", h.ShowEditForm)
	forms.PUT("/:id", h.UpdateForm)
	forms.DELETE("/:id", h.DeleteForm)
}

// ShowNewForm displays the form creation page
func (h *FormHandler) ShowNewForm(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	data := shared.PageData{
		Title:     "Create New Form - GoFormX",
		User:      currentUser,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}

	data.Content = pages.NewFormContent(data)
	return pages.NewForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// CreateForm handles form creation
func (h *FormHandler) CreateForm(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	title := c.FormValue("title")
	description := c.FormValue("description")

	if title == "" || description == "" {
		return h.Base.handleError(errors.New("missing fields"), http.StatusBadRequest, "Title and description are required")
	}

	// Create the form
	formObj, createErr := h.formService.CreateForm(
		currentUser.ID,
		title,
		description,
		form.JSON{
			"display":    "form",
			"components": []any{},
		},
	)
	if createErr != nil {
		return h.Base.handleError(createErr, http.StatusInternalServerError, "Failed to create form")
	}

	// Redirect to the form edit page
	return c.Redirect(http.StatusSeeOther, "/dashboard/forms/"+formObj.ID+"/edit")
}

// ShowEditForm displays the form editing page
func (h *FormHandler) ShowEditForm(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.formOperations.EnsureFormOwnership(c, currentUser, c.Param("id"))
	if err != nil {
		return err
	}

	submissions, err := h.formService.GetFormSubmissions(formObj.ID)
	if err != nil {
		h.logger.Error("failed to get form submissions", err)
		submissions = []*model.FormSubmission{}
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	data := shared.PageData{
		Title:                "Edit Form - GoFormX",
		User:                 currentUser,
		Form:                 formObj,
		Submissions:          submissions,
		CSRFToken:            csrfToken,
		AssetPath:            web.GetAssetPath,
		FormBuilderAssetPath: web.GetAssetPath("src/js/form-builder.ts"),
	}

	return pages.EditForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// UpdateForm handles form updates
func (h *FormHandler) UpdateForm(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.formOperations.EnsureFormOwnership(c, currentUser, c.Param("id"))
	if err != nil {
		return err
	}

	formData, err := h.formOperations.ValidateAndBindFormData(c)
	if err != nil {
		return h.Base.handleError(err, http.StatusBadRequest, "Invalid form data")
	}

	// Update form details
	formObj.Title = formData.Title
	formObj.Description = formData.Description

	if updateErr := h.formService.UpdateForm(formObj); updateErr != nil {
		return h.Base.handleError(updateErr, http.StatusInternalServerError, "Failed to update form")
	}

	return c.JSON(http.StatusOK, formObj)
}

// DeleteForm handles form deletion
func (h *FormHandler) DeleteForm(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.formOperations.EnsureFormOwnership(c, currentUser, c.Param("id"))
	if err != nil {
		return err
	}

	if deleteErr := h.formService.DeleteForm(formObj.ID); deleteErr != nil {
		return h.Base.handleError(deleteErr, http.StatusInternalServerError, "Failed to delete form")
	}

	return c.NoContent(http.StatusNoContent)
}
