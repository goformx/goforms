package handlers

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/core/subscription"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/response"
	"github.com/jonesrussell/goforms/internal/validation"
)

// SubscriptionHandler handles subscription requests
type SubscriptionHandler struct {
	store  subscription.Store
	logger logger.Logger
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(store subscription.Store, log logger.Logger) *SubscriptionHandler {
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
	var sub subscription.Subscription
	if err := c.Bind(&sub); err != nil {
		h.logger.Error("failed to bind subscription request",
			logger.Error(err),
		)
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validation.ValidateSubscription(&sub); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.store.Create(c.Request().Context(), &sub); err != nil {
		h.logger.Error("failed to create subscription",
			logger.Error(err),
			logger.String("email", sub.Email),
		)
		return response.InternalError(c, "Failed to create subscription")
	}

	h.logger.Info("subscription created",
		logger.String("email", sub.Email),
	)

	return response.Created(c, sub)
}
