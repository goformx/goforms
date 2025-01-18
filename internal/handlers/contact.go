package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/jonesrussell/goforms/internal/response"
	"github.com/jonesrussell/goforms/internal/validation"
)

// ContactHandler handles contact form submissions
type ContactHandler struct {
	logger logger.Logger
	store  models.ContactStore
}

// NewContactHandler creates a new contact form handler
func NewContactHandler(logger logger.Logger, store models.ContactStore) *ContactHandler {
	return &ContactHandler{
		logger: logger,
		store:  store,
	}
}

// Register registers the contact form routes with Echo
func (h *ContactHandler) Register(e *echo.Echo) {
	g := e.Group("/api")
	g.GET("/contact", h.GetContacts)
	g.POST("/contact", h.CreateContact)
}

// GetContacts returns all contact form submissions
func (h *ContactHandler) GetContacts(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	submissions, err := h.store.GetContacts(ctx)
	if err != nil {
		h.logger.Error("failed to get contact submissions",
			logger.Error(err),
		)
		return response.Error(c, http.StatusInternalServerError, "failed to get submissions")
	}

	return response.Success(c, http.StatusOK, submissions)
}

// CreateContact handles new contact form submissions
func (h *ContactHandler) CreateContact(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	var contact models.ContactSubmission
	if err := c.Bind(&contact); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request payload")
	}

	if err := validation.ValidateContact(&contact); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	if err := h.store.CreateContact(ctx, &contact); err != nil {
		h.logger.Error("failed to create contact submission",
			logger.Error(err),
			logger.String("email", contact.Email),
		)
		return response.Error(c, http.StatusInternalServerError, "failed to submit contact form")
	}

	return response.Success(c, http.StatusCreated, contact)
}
