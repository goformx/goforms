package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/response"
)

type ContactHandler struct {
	log   logger.Logger
	store contact.Store
}

func NewContactHandler(log logger.Logger, store contact.Store) *ContactHandler {
	return &ContactHandler{
		log:   log,
		store: store,
	}
}

// Register registers the contact routes
func (h *ContactHandler) Register(e *echo.Echo) {
	e.POST("/api/contact", h.CreateContact)
	e.GET("/api/contact", h.GetContacts)
	e.GET("/api/contact/:id", h.GetContact)
	e.PUT("/api/contact/:id/status", h.UpdateContactStatus)
}

func (h *ContactHandler) validateSubmission(s *contact.Submission) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(s.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(s.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	if strings.TrimSpace(s.Message) == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

func (h *ContactHandler) CreateContact(c echo.Context) error {
	var submission contact.Submission
	if err := c.Bind(&submission); err != nil {
		h.log.Error("failed to bind contact submission", logger.Error(err))
		_ = response.Error(c, http.StatusBadRequest, "invalid request")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := h.validateSubmission(&submission); err != nil {
		h.log.Error("invalid contact submission", logger.Error(err))
		_ = response.Error(c, http.StatusBadRequest, "invalid request")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	submission.Status = contact.StatusPending

	if err := h.store.Create(c.Request().Context(), &submission); err != nil {
		h.log.Error("failed to create contact submission", logger.Error(err))
		_ = response.Error(c, http.StatusInternalServerError, "failed to create contact submission")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create contact submission")
	}

	return response.Success(c, http.StatusCreated, submission)
}

func (h *ContactHandler) GetContacts(c echo.Context) error {
	submissions, err := h.store.List(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get contact submissions", logger.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to get contact submissions")
	}

	return c.JSON(http.StatusOK, submissions)
}

func (h *ContactHandler) GetContact(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return response.Error(c, http.StatusBadRequest, "missing contact submission id")
	}

	var submissionID int64
	if _, err := fmt.Sscanf(id, "%d", &submissionID); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid contact submission id")
	}

	submission, err := h.store.GetByID(c.Request().Context(), submissionID)
	if err != nil {
		h.log.Error("failed to get contact submission", logger.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to get contact submission")
	}

	if submission == nil {
		return response.Error(c, http.StatusNotFound, "contact submission not found")
	}

	return response.Success(c, http.StatusOK, submission)
}

func (h *ContactHandler) UpdateContactStatus(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return response.Error(c, http.StatusBadRequest, "missing contact submission id")
	}

	var submissionID int64
	if _, err := fmt.Sscanf(id, "%d", &submissionID); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid contact submission id")
	}

	var req struct {
		Status contact.Status `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		h.log.Error("failed to bind status update request", logger.Error(err))
		return response.Error(c, http.StatusBadRequest, "invalid request")
	}

	if err := h.store.UpdateStatus(c.Request().Context(), submissionID, req.Status); err != nil {
		h.log.Error("failed to update contact submission status", logger.Error(err))
		return response.Error(c, http.StatusInternalServerError, "failed to update contact submission status")
	}

	return response.Success(c, http.StatusOK, nil)
}
