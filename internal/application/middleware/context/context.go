// Package context provides middleware utilities for managing request context and user authentication state.
package context

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Key represents a key in the context
type Key string

const (
	// RequestIDHeader is the HTTP header name for request ID
	RequestIDHeader = "X-Trace-Id"
	// RequestTimeout is the default timeout for request context
	RequestTimeout = 30 * time.Second

	// RequestIDKey is the context key for request ID
	RequestIDKey Key = "request_id"
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey Key = "correlation_id"
	// LoggerKey is the context key for logger
	LoggerKey Key = "logger"
	// UserIDKey is the context key for user ID
	UserIDKey Key = "user_id"
	// EmailKey is the context key for user email
	EmailKey Key = "email"
	// RoleKey is the context key for user role
	RoleKey Key = "role"
	// FirstNameKey is the context key for user first name
	FirstNameKey Key = "first_name"
	// LastNameKey is the context key for user last name
	LastNameKey Key = "last_name"
	// SessionKey is the context key for session
	SessionKey Key = "session"
	// FormIDKey is the context key for form ID
	FormIDKey Key = "form_id"
)

// Middleware provides context handling for HTTP requests
type Middleware struct {
	logger         logging.Logger
	requestTimeout time.Duration
}

// NewMiddleware creates a new context middleware
func NewMiddleware(logger logging.Logger, requestTimeout time.Duration) *Middleware {
	return &Middleware{
		logger:         logger,
		requestTimeout: requestTimeout,
	}
}

// WithContext adds context to the request
func (m *Middleware) WithContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get or generate request ID
			requestID := c.Request().Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
				c.Request().Header.Set(RequestIDHeader, requestID)
			}

			// Create request context with timeout
			ctx, cancel := context.WithTimeout(c.Request().Context(), m.requestTimeout)
			defer cancel()

			// Add request ID and logger to context
			ctx = context.WithValue(ctx, RequestIDKey, requestID)
			ctx = context.WithValue(ctx, LoggerKey, m.logger)

			// Update request context
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// Request Context Helpers

// GetLogger retrieves the logger from context
func GetLogger(ctx context.Context) logging.Logger {
	if logger, ok := ctx.Value(LoggerKey).(logging.Logger); ok {
		return logger
	}

	return nil
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}

	return ""
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}

	return ""
}

// Echo Context Helpers

// GetUserID retrieves the user ID from context
func GetUserID(c echo.Context) (string, bool) {
	if c == nil {
		return "", false
	}

	userID, ok := c.Get(string(UserIDKey)).(string)

	return userID, ok && userID != ""
}

// GetEmail retrieves the user email from context
func GetEmail(c echo.Context) (string, bool) {
	if c == nil {
		return "", false
	}

	email, ok := c.Get(string(EmailKey)).(string)

	return email, ok && email != ""
}

// GetRole retrieves the user role from context
func GetRole(c echo.Context) (string, bool) {
	if c == nil {
		return "", false
	}

	role, ok := c.Get(string(RoleKey)).(string)

	return role, ok && role != ""
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c echo.Context) bool {
	userID, ok := GetUserID(c)

	return ok && userID != ""
}

// IsAdmin checks if the user is an admin
func IsAdmin(c echo.Context) bool {
	role, ok := GetRole(c)

	return ok && role == "admin"
}

// GetFirstName retrieves the user first name from context
func GetFirstName(c echo.Context) (string, bool) {
	if c == nil {
		return "", false
	}

	firstName, ok := c.Get(string(FirstNameKey)).(string)

	return firstName, ok && firstName != ""
}

// GetLastName retrieves the user last name from context
func GetLastName(c echo.Context) (string, bool) {
	if c == nil {
		return "", false
	}

	lastName, ok := c.Get(string(LastNameKey)).(string)

	return lastName, ok && lastName != ""
}

// SetUserID sets the user ID in context
func SetUserID(c echo.Context, userID string) {
	c.Set(string(UserIDKey), userID)
}

// SetEmail sets the user email in context
func SetEmail(c echo.Context, email string) {
	c.Set(string(EmailKey), email)
}

// SetRole sets the user role in context
func SetRole(c echo.Context, role string) {
	c.Set(string(RoleKey), role)
}

// SetFirstName sets the user first name in context
func SetFirstName(c echo.Context, firstName string) {
	c.Set(string(FirstNameKey), firstName)
}

// SetLastName sets the user last name in context
func SetLastName(c echo.Context, lastName string) {
	c.Set(string(LastNameKey), lastName)
}

// ClearUserContext clears all user-related data from context
func ClearUserContext(c echo.Context) {
	c.Set(string(UserIDKey), "")
	c.Set(string(EmailKey), "")
	c.Set(string(RoleKey), "")
	c.Set(string(FirstNameKey), "")
	c.Set(string(LastNameKey), "")
}

// GetFormID retrieves the form ID from context (Go context)
func GetFormID(ctx context.Context) string {
	if id, ok := ctx.Value(FormIDKey).(string); ok {
		return id
	}

	return ""
}

// GetFormIDFromEcho retrieves the form ID from Echo context
func GetFormIDFromEcho(c echo.Context) (string, bool) {
	if c == nil {
		return "", false
	}

	formID, ok := c.Get(string(FormIDKey)).(string)

	return formID, ok && formID != ""
}

// SetFormID sets the form ID in Echo context
func SetFormID(c echo.Context, formID string) {
	c.Set(string(FormIDKey), formID)
}

// SetFormIDInContext sets the form ID in Go context
func SetFormIDInContext(ctx context.Context, formID string) context.Context {
	return context.WithValue(ctx, FormIDKey, formID)
}
