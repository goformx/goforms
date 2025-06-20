package context

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Key represents a key in the context
type Key string

const (
	// Request context keys
	RequestIDHeader = "X-Trace-Id"
	RequestTimeout  = 30 * time.Second

	// Context keys
	RequestIDKey     Key = "request_id"
	CorrelationIDKey Key = "correlation_id"
	LoggerKey        Key = "logger"
	UserIDKey        Key = "user_id"
	EmailKey         Key = "email"
	RoleKey          Key = "role"
	SessionKey       Key = "session"
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

// ClearUserContext clears all user-related data from context
func ClearUserContext(c echo.Context) {
	c.Set(string(UserIDKey), "")
	c.Set(string(EmailKey), "")
	c.Set(string(RoleKey), "")
}
