package handlers

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/services"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// FormHandler handles form-related HTTP requests
type FormHandler struct {
	Base           *BaseHandler
	formService    form.Service
	formOperations *services.FormOperations
	logger         logging.Logger
}

// NewFormHandler creates a new form handler
func NewFormHandler(
	formService form.Service,
	formOperations *services.FormOperations,
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
	h.Base.SetupMiddleware(forms)

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
		Title:     "Create New Form - GoForms",
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

	var formData services.FormData
	if bindErr := c.Bind(&formData); bindErr != nil {
		return h.Base.handleError(bindErr, http.StatusBadRequest, "Invalid form data")
	}

	if validateErr := c.Validate(&formData); validateErr != nil {
		return h.Base.handleError(validateErr, http.StatusUnprocessableEntity, "Form validation failed")
	}

	// Create the form
	formObj, createErr := h.formService.CreateForm(
		currentUser.ID,
		formData.Title,
		formData.Description,
		h.formOperations.CreateDefaultSchema(),
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

	formObj, err := h.Base.getOwnedForm(c, currentUser)
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
		Title:                "Edit Form - GoForms",
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

	formObj, err := h.Base.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	var formData services.FormData
	if bindErr := c.Bind(&formData); bindErr != nil {
		return h.Base.handleError(bindErr, http.StatusBadRequest, "Invalid form data")
	}

	if validateErr := c.Validate(&formData); validateErr != nil {
		return h.Base.handleError(validateErr, http.StatusUnprocessableEntity, "Form validation failed")
	}

	// Update form details
	h.formOperations.UpdateFormDetails(formObj, &formData)

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

	formObj, err := h.Base.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	if deleteErr := h.formService.DeleteForm(formObj.ID); deleteErr != nil {
		return h.Base.handleError(deleteErr, http.StatusInternalServerError, "Failed to delete form")
	}

	return c.NoContent(http.StatusNoContent)
}
