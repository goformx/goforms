package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/response"
)

// ContactAPI handles contact-related API endpoints
type ContactAPI struct {
	service contact.Service
	logger  logging.Logger
}

// NewContactAPI creates a new contact API handler
//
// Dependencies:
//   - service: contact.Service for handling contact business logic
//   - logger: logging.Logger for structured logging
//
// The handler implements RESTful endpoints for contact management:
//   - POST /api/v1/contacts - Create a new contact
//   - GET /api/v1/contacts - List all contacts
//   - GET /api/v1/contacts/:id - Get a specific contact
//   - PUT /api/v1/contacts/:id/status - Update contact status
func NewContactAPI(service contact.Service, logger logging.Logger) *ContactAPI {
	return &ContactAPI{
		service: service,
		logger:  logger,
	}
}

// Register registers the contact API routes with the given Echo instance
func (api *ContactAPI) Register(e *echo.Echo) {
	v1 := e.Group("/api/v1")
	contacts := v1.Group("/contacts")

	contacts.POST("", api.CreateContact)
	contacts.GET("", api.ListContacts)
	contacts.GET("/:id", api.GetContact)
	contacts.PUT("/:id/status", api.UpdateContactStatus)
}

// CreateContact handles contact submission creation
// @Summary Create a new contact submission
// @Description Creates a new contact submission with the provided details
// @Tags contacts
// @Accept json
// @Produce json
// @Param submission body contact.Submission true "Contact submission details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/contacts [post]
func (api *ContactAPI) CreateContact(c echo.Context) error {
	var submission contact.Submission
	if err := c.Bind(&submission); err != nil {
		api.logger.Error("failed to bind contact submission", logging.Error(err))
		return response.Error(c, http.StatusBadRequest, "invalid request")
	}

	if err := api.service.Submit(c.Request().Context(), &submission); err != nil {
		api.logger.Error("failed to create contact submission", logging.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to create contact submission")
	}

	return response.Success(c, http.StatusCreated, submission)
}

// ListContacts handles listing contact submissions
func (api *ContactAPI) ListContacts(c echo.Context) error {
	submissions, err := api.service.ListSubmissions(c.Request().Context())
	if err != nil {
		api.logger.Error("failed to list contact submissions", logging.Error(err))
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
		api.logger.Error("failed to get contact submission", logging.Error(err))
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
		api.logger.Error("failed to bind status update request", logging.Error(err))
		return response.Error(c, http.StatusBadRequest, "invalid request")
	}

	if err := api.service.UpdateSubmissionStatus(c.Request().Context(), id, req.Status); err != nil {
		api.logger.Error("failed to update contact submission status", logging.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to update contact submission status")
	}

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"id":     id,
		"status": req.Status,
	})
}
