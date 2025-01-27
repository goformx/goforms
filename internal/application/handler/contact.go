package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// ContactHandlerOption defines a contact handler option
type ContactHandlerOption func(*ContactHandler)

// WithContactServiceOpt sets the contact service
func WithContactServiceOpt(svc contact.Service) ContactHandlerOption {
	return func(h *ContactHandler) {
		h.contactService = svc
	}
}

// ContactHandler handles contact form submissions
type ContactHandler struct {
	Base
	contactService contact.Service
}

// NewContactHandler creates a new contact handler
func NewContactHandler(logger logging.Logger, opts ...ContactHandlerOption) *ContactHandler {
	h := &ContactHandler{
		Base: Base{Logger: logger},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Validate validates that required dependencies are set
func (h *ContactHandler) Validate() error {
	if err := h.Base.Validate(); err != nil {
		return err
	}
	if h.contactService == nil {
		return fmt.Errorf("contact service is required")
	}
	return nil
}

// Register registers the contact routes
func (h *ContactHandler) Register(e *echo.Echo) {
	if err := h.Validate(); err != nil {
		h.Logger.Error("failed to validate handler", logging.Error(err))
		return
	}

	g := e.Group("/api/v1/contact")
	g.POST("", h.handleSubmit)
	g.GET("", h.handleList)
	g.GET("/:id", h.handleGet)
	g.PUT("/:id/status", h.handleUpdateStatus)
}

// handleSubmit handles contact form submissions
// @Summary Submit contact form
// @Description Submit a new contact form
// @Tags contact
// @Accept json
// @Produce json
// @Param submission body contact.Submission true "Contact form submission"
// @Success 201 {object} contact.Submission
// @Failure 400 {object} echo.HTTPError
// @Router /api/v1/contact [post]
func (h *ContactHandler) handleSubmit(c echo.Context) error {
	var submission contact.Submission
	if err := c.Bind(&submission); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(submission); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.contactService.Submit(c.Request().Context(), &submission); err != nil {
		h.LogError("failed to submit contact form", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit contact form")
	}

	return c.JSON(http.StatusCreated, submission)
}

// handleList handles listing contact form submissions
// @Summary List contact form submissions
// @Description Get a list of all contact form submissions
// @Tags contact
// @Produce json
// @Success 200 {array} contact.Submission
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/contact [get]
func (h *ContactHandler) handleList(c echo.Context) error {
	submissions, err := h.contactService.ListSubmissions(c.Request().Context())
	if err != nil {
		h.LogError("failed to list submissions", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list submissions")
	}

	return c.JSON(http.StatusOK, submissions)
}

// handleGet handles getting a single contact form submission
// @Summary Get contact form submission
// @Description Get a specific contact form submission by ID
// @Tags contact
// @Produce json
// @Param id path int true "Submission ID"
// @Success 200 {object} contact.Submission
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Router /api/v1/contact/{id} [get]
func (h *ContactHandler) handleGet(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	submission, err := h.contactService.GetSubmission(c.Request().Context(), id)
	if err != nil {
		h.LogError("failed to get submission", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get submission")
	}

	if submission == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Submission not found")
	}

	return c.JSON(http.StatusOK, submission)
}

// handleUpdateStatus handles updating a submission's status
// @Summary Update submission status
// @Description Update the status of a contact form submission
// @Tags contact
// @Accept json
// @Produce json
// @Param id path int true "Submission ID"
// @Param status body contact.Status true "New status"
// @Success 200 {object} contact.Submission
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Router /api/v1/contact/{id}/status [put]
func (h *ContactHandler) handleUpdateStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	var status contact.Status
	if err := c.Bind(&status); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid status format")
	}

	if err := h.contactService.UpdateSubmissionStatus(c.Request().Context(), id, status); err != nil {
		h.LogError("failed to update submission status", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update status")
	}

	return c.NoContent(http.StatusOK)
}
