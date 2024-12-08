package handlers

import (
	"net/http"
	"net/mail"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/models"
)

type SubscriptionHandler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSubscriptionHandler(db *sqlx.DB, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		db:     db,
		logger: logger,
	}
}

func (h *SubscriptionHandler) Register(e *echo.Echo) {
	e.POST("/api/subscribe", h.Subscribe)
}

func (h *SubscriptionHandler) Subscribe(c echo.Context) error {
	var req models.SubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// Validate email
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid email address")
	}

	// Check if email already exists
	var exists bool
	err := h.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subscriptions WHERE email = $1)", req.Email)
	if err != nil {
		h.logger.Error("database error", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if exists {
		return echo.NewHTTPError(http.StatusConflict, "Email already subscribed")
	}

	// Insert new subscription
	_, err = h.db.Exec(`
		INSERT INTO subscriptions (email, status)
		VALUES ($1, 'active')
	`, req.Email)

	if err != nil {
		h.logger.Error("failed to insert subscription", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process subscription")
	}

	return c.JSON(http.StatusCreated, models.SubscriptionResponse{
		Message: "Successfully subscribed",
	})
}
