package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// AuthResponseBuilder builds authentication responses
type AuthResponseBuilder struct {
	config       *config.Config
	assetManager web.AssetManagerInterface
	renderer     view.Renderer
	logger       logging.Logger
}

// NewAuthResponseBuilder creates a new auth response builder
func NewAuthResponseBuilder(
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	renderer view.Renderer,
	logger logging.Logger,
) *AuthResponseBuilder {
	return &AuthResponseBuilder{
		config:       cfg,
		assetManager: assetManager,
		renderer:     renderer,
		logger:       logger,
	}
}

// BuildLoginSuccessResponse builds a successful login response
func (b *AuthResponseBuilder) BuildLoginSuccessResponse(c echo.Context, user *entities.User) error {
	b.logger.Info("user logged in successfully", "user_id", user.ID, "email", user.Email)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": constants.MsgLoginSuccess,
			"user": map[string]interface{}{
				"id":    user.ID,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}

	// For web requests, redirect to dashboard
	return c.Redirect(http.StatusSeeOther, constants.PathDashboard)
}

// BuildSignupSuccessResponse builds a successful signup response
func (b *AuthResponseBuilder) BuildSignupSuccessResponse(c echo.Context, user *entities.User) error {
	b.logger.Info("user signed up successfully", "user_id", user.ID, "email", user.Email)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"status":  "success",
			"message": constants.MsgSignupSuccess,
			"user": map[string]interface{}{
				"id":    user.ID,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}

	// For web requests, redirect to dashboard
	return c.Redirect(http.StatusSeeOther, constants.PathDashboard)
}

// BuildLogoutSuccessResponse builds a successful logout response
func (b *AuthResponseBuilder) BuildLogoutSuccessResponse(c echo.Context) error {
	b.logger.Info("user logged out successfully")

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": constants.MsgLogoutSuccess,
		})
	}

	// For web requests, redirect to login
	return c.Redirect(http.StatusSeeOther, constants.PathLogin)
}

// BuildLoginErrorResponse builds an error response for login failures
func (b *AuthResponseBuilder) BuildLoginErrorResponse(c echo.Context, errorMessage string) error {
	b.logger.Warn("login failed", "error", errorMessage)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status":  "error",
			"message": errorMessage,
		})
	}

	// For web requests, render login page with error
	return b.renderLoginWithError(c, errorMessage)
}

// BuildSignupErrorResponse builds an error response for signup failures
func (b *AuthResponseBuilder) BuildSignupErrorResponse(c echo.Context, errorMessage string) error {
	b.logger.Warn("signup failed", "error", errorMessage)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": errorMessage,
		})
	}

	// For web requests, render signup page with error
	return b.renderSignupWithError(c, errorMessage)
}

// BuildValidationErrorResponse builds a validation error response
func (b *AuthResponseBuilder) BuildValidationErrorResponse(c echo.Context, field, message string) error {
	b.logger.Warn("validation error", "field", field, "message", message)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors": map[string]string{
				field: message,
			},
		})
	}

	// For web requests, render appropriate page with error
	// This is a simplified version - you might want to handle different pages
	return b.renderLoginWithError(c, fmt.Sprintf("%s: %s", field, message))
}

// Helper methods

// isAPIRequest checks if the request is an API request
func (b *AuthResponseBuilder) isAPIRequest(c echo.Context) bool {
	accept := c.Request().Header.Get("Accept")
	contentType := c.Request().Header.Get("Content-Type")

	return strings.Contains(accept, "application/json") ||
		strings.Contains(contentType, "application/json") ||
		strings.HasPrefix(c.Request().URL.Path, "/api/")
}

// renderLoginWithError renders the login page with an error message
func (b *AuthResponseBuilder) renderLoginWithError(c echo.Context, errorMessage string) error {
	pageData := view.NewPageData(b.config, b.assetManager, c, "Login")
	pageData = pageData.WithMessage("error", errorMessage)

	loginComponent := pages.Login(*pageData)

	return b.renderer.Render(c, loginComponent)
}

// renderSignupWithError renders the signup page with an error message
func (b *AuthResponseBuilder) renderSignupWithError(c echo.Context, errorMessage string) error {
	pageData := view.NewPageData(b.config, b.assetManager, c, "Sign Up")
	pageData = pageData.WithMessage("error", errorMessage)

	signupComponent := pages.Signup(*pageData)

	return b.renderer.Render(c, signupComponent)
}
