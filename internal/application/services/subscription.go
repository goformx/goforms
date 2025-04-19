package services

import (
	"errors"
	"fmt"
	"net/http"

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

// wrapResponseError wraps errors from the response package
func (h *SubscriptionHandler) wrapResponseError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// HandleSubscribe handles subscription creation requests
func (h *SubscriptionHandler) HandleSubscribe(c echo.Context) error {
	var sub subscription.Subscription
	if err := c.Bind(&sub); err != nil {
		h.logger.Error("failed to bind subscription request",
			logging.Error(err),
		)
		return h.wrapResponseError(response.BadRequest(c, "Invalid request body"), "failed to handle subscription request")
	}

	if err := sub.Validate(); err != nil {
		return h.wrapResponseError(response.BadRequest(c, err.Error()), "failed to validate subscription")
	}

	// Check if subscription already exists
	existing, err := h.store.GetByEmail(c.Request().Context(), sub.Email)
	if err != nil && !errors.Is(err, subscription.ErrSubscriptionNotFound) {
		h.logger.Error("failed to check existing subscription",
			logging.Error(err),
			logging.String("email", sub.Email),
		)
		return h.wrapResponseError(
			response.InternalError(c, "Failed to create subscription"),
			"failed to check existing subscription",
		)
	}
	if existing != nil {
		return h.wrapResponseError(response.BadRequest(c, "Email already subscribed"), "duplicate subscription")
	}

	// Set initial status to pending
	sub.Status = subscription.StatusPending

	createErr := h.store.Create(c.Request().Context(), &sub)
	if createErr != nil {
		h.logger.Error("failed to create subscription",
			logging.Error(createErr),
			logging.String("email", sub.Email),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create subscription")
	}

	h.logger.Info("subscription created",
		logging.String("email", sub.Email),
	)

	return h.wrapResponseError(response.Created(c, sub), "failed to send response")
}
