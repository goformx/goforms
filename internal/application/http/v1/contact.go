package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/response"
)

// ContactAPI handles contact-related API endpoints
type ContactAPI struct {
	service contact.Service
	logger  logger.Logger
}

// NewContactAPI creates a new contact API handler
func NewContactAPI(service contact.Service, log logger.Logger) *ContactAPI {
	return &ContactAPI{
		service: service,
		logger:  log,
	}
}

// Register registers the contact API routes
func (api *ContactAPI) Register(e *echo.Echo) {
	v1 := e.Group("/api/v1")
	contacts := v1.Group("/contacts")

	contacts.POST("", api.CreateContact)
	contacts.GET("", api.ListContacts)
	contacts.GET("/:id", api.GetContact)
	contacts.PUT("/:id/status", api.UpdateContactStatus)
}

// CreateContact handles contact submission creation
func (api *ContactAPI) CreateContact(c echo.Context) error {
	var submission contact.Submission
	if err := c.Bind(&submission); err != nil {
		api.logger.Error("failed to bind contact submission", logger.Error(err))
		return response.Error(c, http.StatusBadRequest, "invalid request")
	}

	if err := api.service.CreateSubmission(c.Request().Context(), &submission); err != nil {
		api.logger.Error("failed to create contact submission", logger.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to create contact submission")
	}

	return response.Success(c, http.StatusCreated, submission)
}

// ListContacts handles listing contact submissions
func (api *ContactAPI) ListContacts(c echo.Context) error {
	submissions, err := api.service.ListSubmissions(c.Request().Context())
	if err != nil {
		api.logger.Error("failed to list contact submissions", logger.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to list contact submissions")
	}

	return response.Success(c, http.StatusOK, submissions)
}

// GetContact handles retrieving a single contact submission
func (api *ContactAPI) GetContact(c echo.Context) error {
	id, err := response.ParseInt64Param(c, "id")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid contact submission id")
	}

	submission, err := api.service.GetSubmission(c.Request().Context(), id)
	if err != nil {
		api.logger.Error("failed to get contact submission", logger.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to get contact submission")
	}

	if submission == nil {
		return response.Error(c, http.StatusNotFound, "contact submission not found")
	}

	return response.Success(c, http.StatusOK, submission)
}

// UpdateContactStatus handles updating a contact submission's status
func (api *ContactAPI) UpdateContactStatus(c echo.Context) error {
	id, err := response.ParseInt64Param(c, "id")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid contact submission id")
	}

	var req struct {
		Status contact.Status `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		api.logger.Error("failed to bind status update request", logger.Error(err))
		return response.Error(c, http.StatusBadRequest, "invalid request")
	}

	if err := api.service.UpdateSubmissionStatus(c.Request().Context(), id, req.Status); err != nil {
		api.logger.Error("failed to update contact submission status", logger.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to update contact submission status")
	}

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"id":     id,
		"status": req.Status,
	})
}
