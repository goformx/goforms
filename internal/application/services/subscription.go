package services

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/response"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// SubscriptionHandler handles subscription requests
type SubscriptionHandler struct {
	store  subscription.Store
	logger logging.Logger
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(store subscription.Store, log logging.Logger) *SubscriptionHandler {
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
			logging.Error(err),
		)
		return response.BadRequest(c, "Invalid request body")
	}

	if err := sub.Validate(); err != nil {
		return response.BadRequest(c, err.Error())
	}

	if err := h.store.Create(c.Request().Context(), &sub); err != nil {
		h.logger.Error("failed to create subscription",
			logging.Error(err),
			logging.String("email", sub.Email),
		)
		return response.InternalError(c, "Failed to create subscription")
	}

	h.logger.Info("subscription created",
		logging.String("email", sub.Email),
	)

	return response.Created(c, sub)
}
