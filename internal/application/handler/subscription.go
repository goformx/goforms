package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/response"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// SubscriptionHandlerOption defines a subscription handler option
type SubscriptionHandlerOption func(*SubscriptionHandler)

// WithSubscriptionService sets the subscription service
func WithSubscriptionService(svc subscription.Service) SubscriptionHandlerOption {
	return func(h *SubscriptionHandler) {
		h.subscriptionService = svc
	}
}

// SubscriptionHandler handles subscription-related requests
type SubscriptionHandler struct {
	*Base
	subscriptionService subscription.Service
}

// NewSubscriptionHandler creates a new SubscriptionHandler
func NewSubscriptionHandler(logger logging.Logger, opts ...SubscriptionHandlerOption) *SubscriptionHandler {
	h := &SubscriptionHandler{
		Base: &Base{Logger: logger},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Validate ensures all required dependencies are set
func (h *SubscriptionHandler) Validate() error {
	if err := h.Base.Validate(); err != nil {
		return err
	}
	if h.subscriptionService == nil {
		return errors.New("missing required dependency: subscription service")
	}
	return nil
}

// Register registers the subscription routes
func (h *SubscriptionHandler) Register(e *echo.Echo) {
	if err := h.Validate(); err != nil {
		h.Logger.Error("failed to validate handler", logging.Error(err))
		return
	}

	g := e.Group("/api/v1/subscriptions")
	g.POST("", h.handleCreate)
	g.GET("", h.handleList)
	g.GET("/:id", h.handleGet)
	g.PUT("/:id/status", h.handleUpdate)
	g.DELETE("/:id", h.handleDelete)
}

// handleCreate handles creating a new subscription
// @Summary Create subscription
// @Description Create a new demo form submission
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
// @Description Get a list of all demo form submissions
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

// handleUpdate handles updating a subscription's status
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
func (h *SubscriptionHandler) handleUpdate(c echo.Context) error {
	id, parseErr := strconv.ParseInt(c.Param("id"), 10, 64)
	if parseErr != nil {
		return response.BadRequest(c, "invalid subscription ID")
	}

	var status subscription.Status
	if bindErr := c.Bind(&status); bindErr != nil {
		return response.BadRequest(c, "invalid status")
	}

	if updateErr := h.subscriptionService.UpdateSubscriptionStatus(c.Request().Context(), id, status); updateErr != nil {
		return h.handleError(c, updateErr)
	}

	return response.Success(c, nil)
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
	id, parseErr := strconv.ParseInt(c.Param("id"), 10, 64)
	if parseErr != nil {
		return response.BadRequest(c, "invalid subscription ID")
	}

	if deleteErr := h.subscriptionService.DeleteSubscription(c.Request().Context(), id); deleteErr != nil {
		return h.handleError(c, deleteErr)
	}

	return response.Success(c, nil)
}

func (h *SubscriptionHandler) handleError(c echo.Context, err error) error {
	h.Logger.Error("subscription handler error", logging.Error(err))
	switch err {
	case subscription.ErrSubscriptionNotFound:
		return response.NotFound(c, "subscription not found")
	case subscription.ErrInvalidStatus:
		return response.BadRequest(c, "invalid status")
	default:
		return response.InternalError(c, "internal server error")
	}
}
