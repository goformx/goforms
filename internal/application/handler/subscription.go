package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// SubscriptionHandler handles subscription-related requests
type SubscriptionHandler struct {
	Base
	subscriptionService subscription.Service
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(logger logging.Logger, subscriptionService subscription.Service) *SubscriptionHandler {
	return &SubscriptionHandler{
		Base:                Base{Logger: logger},
		subscriptionService: subscriptionService,
	}
}

// Register registers the subscription routes
func (h *SubscriptionHandler) Register(e *echo.Echo) {
	g := e.Group("/api/v1/subscriptions")
	g.POST("", h.handleCreate)
	g.GET("", h.handleList)
	g.GET("/:id", h.handleGet)
	g.PUT("/:id/status", h.handleUpdateStatus)
	g.DELETE("/:id", h.handleDelete)
}

// handleCreate handles creating a new subscription
// @Summary Create subscription
// @Description Create a new newsletter subscription
// @Tags subscription
// @Accept json
// @Produce json
// @Param subscription body subscription.Subscription true "Subscription details"
// @Success 201 {object} subscription.Subscription
// @Failure 400 {object} echo.HTTPError
// @Router /api/v1/subscriptions [post]
func (h *SubscriptionHandler) handleCreate(c echo.Context) error {
	var sub subscription.Subscription
	if err := c.Bind(&sub); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(sub); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.subscriptionService.CreateSubscription(c.Request().Context(), &sub); err != nil {
		h.LogError("failed to create subscription", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create subscription")
	}

	return c.JSON(http.StatusCreated, sub)
}

// handleList handles listing all subscriptions
// @Summary List subscriptions
// @Description Get a list of all newsletter subscriptions
// @Tags subscription
// @Produce json
// @Success 200 {array} subscription.Subscription
// @Failure 500 {object} echo.HTTPError
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) handleList(c echo.Context) error {
	subs, err := h.subscriptionService.ListSubscriptions(c.Request().Context())
	if err != nil {
		h.LogError("failed to list subscriptions", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list subscriptions")
	}

	return c.JSON(http.StatusOK, subs)
}

// handleGet handles getting a single subscription
// @Summary Get subscription
// @Description Get a specific subscription by ID
// @Tags subscription
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} subscription.Subscription
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) handleGet(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	sub, err := h.subscriptionService.GetSubscription(c.Request().Context(), id)
	if err != nil {
		h.LogError("failed to get subscription", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get subscription")
	}

	if sub == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Subscription not found")
	}

	return c.JSON(http.StatusOK, sub)
}

// handleUpdateStatus handles updating a subscription's status
// @Summary Update subscription status
// @Description Update the status of a subscription
// @Tags subscription
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param status body subscription.Status true "New status"
// @Success 200 {object} subscription.Subscription
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Router /api/v1/subscriptions/{id}/status [put]
func (h *SubscriptionHandler) handleUpdateStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	var status subscription.Status
	if err := c.Bind(&status); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid status format")
	}

	if err := h.subscriptionService.UpdateSubscriptionStatus(c.Request().Context(), id, status); err != nil {
		h.LogError("failed to update subscription status", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update status")
	}

	return c.NoContent(http.StatusOK)
}

// handleDelete handles deleting a subscription
// @Summary Delete subscription
// @Description Delete a subscription by ID
// @Tags subscription
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 204 "No Content"
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandler) handleDelete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	if err := h.subscriptionService.DeleteSubscription(c.Request().Context(), id); err != nil {
		h.LogError("failed to delete subscription", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete subscription")
	}

	return c.NoContent(http.StatusNoContent)
}
