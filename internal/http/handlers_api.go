package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

func (h *Handlers) registerContact(g *echo.Group) {
	g.POST("/contact", h.handleContactSubmit)
	g.GET("/contact", h.handleContactList)
	g.GET("/contact/:id", h.handleContactGet)
}

func (h *Handlers) registerSubscription(g *echo.Group) {
	g.POST("/subscription", h.handleSubscriptionCreate)
	g.GET("/subscription", h.handleSubscriptionList)
	g.GET("/subscription/:id", h.handleSubscriptionGet)
}

func (h *Handlers) handleContactList(c echo.Context) error {
	submissions, err := h.contactService.ListSubmissions(c.Request().Context())
	if err != nil {
		h.logger.Error("failed to list contacts", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list contacts")
	}
	return c.JSON(http.StatusOK, submissions)
}

func (h *Handlers) handleContactGet(c echo.Context) error {
	id := c.Param("id")
	// Convert id to int64...
	submission, err := h.contactService.GetSubmission(c.Request().Context(), 0) // TODO: proper ID conversion
	if err != nil {
		h.logger.Error("failed to get contact", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get contact")
	}
	return c.JSON(http.StatusOK, submission)
}

func (h *Handlers) handleContactSubmit(c echo.Context) error {
	var submission contact.Submission
	if err := c.Bind(&submission); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.contactService.Submit(c.Request().Context(), &submission); err != nil {
		h.logger.Error("failed to submit contact form", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit contact form")
	}

	return c.JSON(http.StatusCreated, submission)
}

func (h *Handlers) handleSubscriptionCreate(c echo.Context) error {
	var sub subscription.Subscription
	if err := c.Bind(&sub); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.subscriptionService.Create(c.Request().Context(), &sub); err != nil {
		h.logger.Error("failed to create subscription", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create subscription")
	}

	return c.JSON(http.StatusCreated, sub)
}

func (h *Handlers) handleSubscriptionList(c echo.Context) error {
	subs, err := h.subscriptionService.List(c.Request().Context())
	if err != nil {
		h.logger.Error("failed to list subscriptions", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list subscriptions")
	}
	return c.JSON(http.StatusOK, subs)
}

func (h *Handlers) handleSubscriptionGet(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	sub, err := h.subscriptionService.Get(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get subscription", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get subscription")
	}
	return c.JSON(http.StatusOK, sub)
}

// ... implement other contact handlers
