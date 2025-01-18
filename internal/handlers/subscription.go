package handlers

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/jonesrussell/goforms/internal/response"
	"github.com/jonesrussell/goforms/internal/validation"
)

// SubscriptionHandler handles subscription requests
type SubscriptionHandler struct {
	store  models.SubscriptionStore
	logger logger.Logger
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(store models.SubscriptionStore, log logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		store:  store,
		logger: log,
	}
}

// Register registers the subscription routes
func (h *SubscriptionHandler) Register(e *echo.Echo) {
	e.POST("/api/v1/subscribe", h.HandleSubscribe)
}

// HandleSubscribe handles subscription creation requests
func (h *SubscriptionHandler) HandleSubscribe(c echo.Context) error {
	var subscription models.Subscription
	if err := c.Bind(&subscription); err != nil {
		h.logger.Error("failed to bind subscription request",
			logger.Error(err),
		)
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateSubscription(&subscription); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.store.Create(&subscription); err != nil {
		h.logger.Error("failed to create subscription",
			logger.Error(err),
			logger.String("email", subscription.Email),
		)
		return response.InternalError(c, "Failed to create subscription")
	}

	h.logger.Info("subscription created",
		logger.String("email", subscription.Email),
	)

	return response.Created(c, subscription)
}
