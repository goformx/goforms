package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/response"
)

// SubscriptionAPI handles subscription-related API endpoints
type SubscriptionAPI struct {
	service subscription.Service
	logger  logging.Logger
}

// NewSubscriptionAPI creates a new subscription API handler
//
// Dependencies:
//   - service: subscription.Service for handling subscription business logic
//   - logger: logging.Logger for structured logging
//
// The handler implements RESTful endpoints for subscription management:
//   - POST /api/v1/subscriptions - Create a new subscription
//   - GET /api/v1/subscriptions - List all subscriptions
//   - GET /api/v1/subscriptions/:id - Get a specific subscription
//   - PUT /api/v1/subscriptions/:id/status - Update subscription status
//   - DELETE /api/v1/subscriptions/:id - Delete a subscription
func NewSubscriptionAPI(service subscription.Service, logger logging.Logger) *SubscriptionAPI {
	return &SubscriptionAPI{
		service: service,
		logger:  logger,
	}
}

// Register registers the subscription API routes with the given Echo instance
func (api *SubscriptionAPI) Register(e *echo.Echo) {
	// Public routes
	v1 := e.Group("/api/v1")
	v1.POST("/subscribe", api.CreateSubscription)

	// Protected routes
	protected := v1.Group("/subscriptions", api.requireAuth())
	protected.GET("", api.ListSubscriptions)
	protected.GET("/:id", api.GetSubscription)
	protected.PUT("/:id/status", api.UpdateSubscriptionStatus)
	protected.DELETE("/:id", api.DeleteSubscription)
}

// requireAuth returns middleware that requires authentication
func (api *SubscriptionAPI) requireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}
			return next(c)
		}
	}
}

// wrapResponseError wraps errors from the response package
func (api *SubscriptionAPI) wrapResponseError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// CreateSubscription handles subscription creation
// @Summary Create a new subscription
// @Description Creates a new subscription with the provided details
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body subscription.Subscription true "Subscription details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions [post]
func (api *SubscriptionAPI) CreateSubscription(c echo.Context) error {
	var sub subscription.Subscription
	if err := c.Bind(&sub); err != nil {
		api.logger.Error("failed to bind subscription", logging.Error(err))
		return api.wrapResponseError(
			response.Error(c, http.StatusBadRequest, "invalid request"),
			"failed to bind request")
	}

	if err := sub.Validate(); err != nil {
		api.logger.Error("failed to validate subscription", logging.Error(err))
		return api.wrapResponseError(
			response.Error(c, http.StatusBadRequest, err.Error()),
			"failed to validate subscription")
	}

	if err := api.service.CreateSubscription(c.Request().Context(), &sub); err != nil {
		api.logger.Error("failed to create subscription", logging.Error(err))
		return api.wrapResponseError(
			response.Error(c, http.StatusInternalServerError, "failed to create subscription"),
			"failed to create subscription")
	}

	return api.wrapResponseError(
		response.Success(c, http.StatusCreated, sub),
		"failed to send response")
}

// ListSubscriptions handles listing all subscriptions
func (api *SubscriptionAPI) ListSubscriptions(c echo.Context) error {
	subs, err := api.service.ListSubscriptions(c.Request().Context())
	if err != nil {
		api.logger.Error("failed to list subscriptions", logging.Error(err))
		return api.wrapResponseError(
			response.Error(c, http.StatusInternalServerError, "failed to list subscriptions"),
			"failed to list subscriptions")
	}

	return api.wrapResponseError(
		response.Success(c, http.StatusOK, subs),
		"failed to send response")
}

// GetSubscription handles retrieving a single subscription
func (api *SubscriptionAPI) GetSubscription(c echo.Context) error {
	id, err := api.parseID(c)
	if err != nil {
		return api.wrapResponseError(
			response.Error(c, http.StatusBadRequest, "invalid subscription id"),
			"failed to parse id")
	}

	sub, err := api.service.GetSubscription(c.Request().Context(), id)
	if err != nil {
		api.logger.Error("failed to get subscription", logging.Error(err))
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			return api.wrapResponseError(
				response.Error(c, http.StatusNotFound, "subscription not found"),
				"subscription not found")
		}
		return api.wrapResponseError(
			response.Error(c, http.StatusInternalServerError, "failed to get subscription"),
			"failed to get subscription")
	}

	return api.wrapResponseError(
		response.Success(c, http.StatusOK, sub),
		"failed to send response")
}

// UpdateSubscriptionStatus handles updating a subscription's status
func (api *SubscriptionAPI) UpdateSubscriptionStatus(c echo.Context) error {
	id, err := api.parseID(c)
	if err != nil {
		return api.wrapResponseError(
			response.Error(c, http.StatusBadRequest, "invalid subscription id"),
			"failed to parse id")
	}

	var req struct {
		Status subscription.Status `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		api.logger.Error("failed to bind status update request", logging.Error(err))
		return api.wrapResponseError(
			response.Error(c, http.StatusBadRequest, "invalid request"),
			"failed to bind request")
	}

	if err := api.service.UpdateSubscriptionStatus(c.Request().Context(), id, req.Status); err != nil {
		api.logger.Error("failed to update subscription status", logging.Error(err))
		return api.wrapResponseError(
			response.Error(c, http.StatusInternalServerError, "failed to update subscription status"),
			"failed to update subscription status")
	}

	return api.wrapResponseError(
		response.Success(c, http.StatusOK, map[string]interface{}{
			"id":     id,
			"status": req.Status,
		}),
		"failed to send response")
}

// DeleteSubscription handles deleting a subscription
func (api *SubscriptionAPI) DeleteSubscription(c echo.Context) error {
	id, err := api.parseID(c)
	if err != nil {
		return api.wrapResponseError(response.Error(c, http.StatusBadRequest, "invalid subscription id"), "failed to parse id")
	}

	if err := api.service.DeleteSubscription(c.Request().Context(), id); err != nil {
		api.logger.Error("failed to delete subscription", logging.Error(err))
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			return api.wrapResponseError(response.Error(c, http.StatusNotFound, err.Error()), "subscription not found")
		}
		return api.wrapResponseError(response.Error(c, http.StatusInternalServerError, "failed to delete subscription"), "failed to delete subscription")
	}

	return api.wrapResponseError(response.Success(c, http.StatusOK, map[string]interface{}{
		"id":      id,
		"deleted": true,
	}), "failed to send response")
}

// parseID parses the ID parameter from the request
func (api *SubscriptionAPI) parseID(c echo.Context) (int64, error) {
	id := c.Param("id")
	if id == "" {
		return 0, errors.New("missing id parameter")
	}
	parsed, err := subscription.ParseID(id)
	if err != nil {
		return 0, fmt.Errorf("failed to parse subscription ID: %w", err)
	}
	return parsed, nil
}
