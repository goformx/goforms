package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ContactHandler handles contact form submissions
type ContactHandler struct {
	logger *zap.Logger
	store  models.ContactStore
}

// NewContactHandler creates a new contact form handler
func NewContactHandler(logger *zap.Logger, store models.ContactStore) *ContactHandler {
	return &ContactHandler{
		logger: logger,
		store:  store,
	}
}

// CreateContact handles new contact form submissions
func (h *ContactHandler) CreateContact(c echo.Context) error {
	// Add timeout context
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Use ctx for database operations
	c.SetRequest(c.Request().WithContext(ctx))

	var contact models.ContactSubmission
	if err := c.Bind(&contact); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	if err := contact.Validate(); err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.store.CreateContact(ctx, &contact); err != nil {
		h.logger.Error("failed to create contact submission",
			zap.Error(err),
			zap.String("email", contact.Email),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to submit contact form")
	}

	return c.JSON(http.StatusCreated, contact)
}

// Register registers the contact form routes with Echo
func (h *ContactHandler) Register(e *echo.Echo) {
	g := e.Group("/app")
	g.POST("/contact", h.CreateContact)
}
