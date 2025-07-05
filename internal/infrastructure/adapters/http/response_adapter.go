package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goformx/goforms/internal/application/dto"
	"github.com/goformx/goforms/internal/infrastructure/session"
	"github.com/labstack/echo/v4"
)

// EchoResponseAdapter implements ResponseAdapter for Echo
type EchoResponseAdapter struct {
	sessionManager *session.Manager
}

// NewEchoResponseAdapter creates a new Echo response adapter
func NewEchoResponseAdapter(sessionManager *session.Manager) *EchoResponseAdapter {
	return &EchoResponseAdapter{
		sessionManager: sessionManager,
	}
}

// BuildLoginResponse builds login response for Echo context
func (a *EchoResponseAdapter) BuildLoginResponse(ctx Context, response *dto.LoginResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	// Check if this is a form submission (POST request to /login)
	if echoCtx.Method() == "POST" && echoCtx.Path() == "/login" {
		// Set session cookie before redirecting
		a.setSessionCookie(echoCtx, response.SessionID, response.ExpiresAt)

		// For login form submissions, always redirect to dashboard
		// This ensures proper login flow regardless of Accept header
		return echoCtx.Redirect(http.StatusFound, "/dashboard")
	}

	// For actual API requests (non-form submissions), return JSON
	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusOK, response)
	}

	// For web requests, redirect to dashboard
	return echoCtx.Redirect(http.StatusFound, "/dashboard")
}

// setSessionCookie sets the session cookie in the response
func (a *EchoResponseAdapter) setSessionCookie(ctx *EchoContextAdapter, sessionID string, expiresAt time.Time) {
	// Get the configured cookie name from session manager
	cookieName := a.sessionManager.GetCookieName()

	cookie := &http.Cookie{
		Name:     cookieName, // Use configured cookie name from session manager
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	}

	ctx.SetCookie(cookie)
}

// BuildSignupResponse builds signup response for Echo context
func (a *EchoResponseAdapter) BuildSignupResponse(ctx Context, response *dto.SignupResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	// Check if this is a form submission (POST request to /signup)
	if echoCtx.Method() == "POST" && echoCtx.Path() == "/signup" {
		// Set session cookie before redirecting
		a.setSessionCookie(echoCtx, response.SessionID, response.ExpiresAt)

		// For signup form submissions, always redirect to dashboard
		// This ensures proper signup flow regardless of Accept header
		return echoCtx.Redirect(http.StatusFound, "/dashboard")
	}

	// For actual API requests (non-form submissions), return JSON
	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusCreated, map[string]any{
			"user":       response.User,
			"session_id": response.SessionID,
			"expires_at": response.ExpiresAt,
			"data": map[string]any{
				"redirect": "/dashboard",
				"message":  "Account created successfully! Redirecting to dashboard...",
			},
		})
	}

	// For web requests, redirect to dashboard
	return echoCtx.Redirect(http.StatusFound, "/dashboard")
}

// BuildLogoutResponse builds logout response for Echo context
func (a *EchoResponseAdapter) BuildLogoutResponse(ctx Context, response *dto.LogoutResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusOK, response)
	}

	// For web requests, redirect to login
	return echoCtx.Redirect(http.StatusFound, "/login")
}

// BuildFormResponse builds form response for Echo context
func (a *EchoResponseAdapter) BuildFormResponse(ctx Context, response *dto.FormResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, response)
}

// BuildFormListResponse builds form list response for Echo context
func (a *EchoResponseAdapter) BuildFormListResponse(ctx Context, response *dto.FormListResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, response)
}

// BuildFormSchemaResponse builds form schema response for Echo context
func (a *EchoResponseAdapter) BuildFormSchemaResponse(ctx Context, response *dto.FormSchemaResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, response)
}

// BuildSubmitFormResponse builds submit form response for Echo context
func (a *EchoResponseAdapter) BuildSubmitFormResponse(ctx Context, response *dto.SubmitFormResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusCreated, response)
}

// BuildErrorResponse builds error response for Echo context
func (a *EchoResponseAdapter) BuildErrorResponse(ctx Context, err error) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusInternalServerError, map[string]any{
		"error": err.Error(),
	})
}

// BuildValidationErrorResponse builds validation error response for Echo context
func (a *EchoResponseAdapter) BuildValidationErrorResponse(ctx Context, errors []dto.ValidationError) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusBadRequest, map[string]any{
		"errors": errors,
	})
}

// BuildNotFoundResponse builds not found response for Echo context
func (a *EchoResponseAdapter) BuildNotFoundResponse(ctx Context, resource string) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusNotFound, map[string]any{
		"error": fmt.Sprintf("%s not found", resource),
	})
}

// BuildUnauthorizedResponse builds unauthorized response for Echo context
func (a *EchoResponseAdapter) BuildUnauthorizedResponse(ctx Context) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusUnauthorized, map[string]any{
		"error": "unauthorized",
	})
}

// BuildForbiddenResponse builds forbidden response for Echo context
func (a *EchoResponseAdapter) BuildForbiddenResponse(ctx Context) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusForbidden, map[string]any{
		"error": "forbidden",
	})
}

// BuildSuccessResponse builds success response for Echo context
func (a *EchoResponseAdapter) BuildSuccessResponse(ctx Context, message string, data any) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, map[string]any{
		"message": message,
		"data":    data,
	})
}

// BuildJSONResponse builds generic JSON response for Echo context
func (a *EchoResponseAdapter) BuildJSONResponse(ctx Context, statusCode int, data any) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(statusCode, data)
}

// isAPIRequest checks if the request is an API request
func (a *EchoResponseAdapter) isAPIRequest(ctx echo.Context) bool {
	accept := ctx.Request().Header.Get("Accept")

	return strings.Contains(accept, "application/json")
}
