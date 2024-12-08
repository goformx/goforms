package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// SubscriptionHandler handles subscription-related requests
type SubscriptionHandler struct {
	logger *zap.Logger
	store  models.SubscriptionStore
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(logger *zap.Logger, store models.SubscriptionStore) *SubscriptionHandler {
	return &SubscriptionHandler{
		logger: logger,
		store:  store,
	}
}

// CreateSubscription handles the creation of new subscriptions
func (h *SubscriptionHandler) CreateSubscription(c echo.Context) error {
	var sub models.Subscription
	if err := c.Bind(&sub); err != nil {
		h.logger.Error("failed to bind subscription", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	if err := sub.Validate(); err != nil {
		h.logger.Error("subscription validation failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.store.CreateSubscription(c.Request().Context(), &sub); err != nil {
		h.logger.Error("failed to create subscription", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create subscription")
	}

	return c.JSON(http.StatusCreated, sub)
}
