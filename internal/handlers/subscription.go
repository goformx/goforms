package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
)

// SubscriptionHandler handles subscription-related requests
type SubscriptionHandler struct {
	store models.SubscriptionStore
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(store models.SubscriptionStore) *SubscriptionHandler {
	return &SubscriptionHandler{
		store: store,
	}
}

// CreateSubscription handles the creation of new subscriptions
func (h *SubscriptionHandler) CreateSubscription(c echo.Context) error {
	// Add timeout context
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Use ctx for database operations
	c.SetRequest(c.Request().WithContext(ctx))

	var sub models.Subscription
	if err := c.Bind(&sub); err != nil {
		if sub.Email == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "email is required")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	if err := sub.Validate(); err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.store.CreateSubscription(c.Request().Context(), &sub); err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create subscription")
	}

	return c.JSON(http.StatusCreated, sub)
}

// Register registers the subscription routes with Echo
func (h *SubscriptionHandler) Register(e *echo.Echo) {
	api := e.Group("/api")
	api.POST("/subscriptions", h.CreateSubscription)
}
