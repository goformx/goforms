// Package constants provides application-wide constants for HTTP status codes,
// paths, timeouts, headers, content types, and other configuration values.
// These constants are used throughout the application to ensure consistency
// and maintainability.
package constants

import (
	"net/http"
)

// HTTP Status Codes
const (
	// Success responses
	StatusOK        = http.StatusOK        // 200
	StatusCreated   = http.StatusCreated   // 201
	StatusNoContent = http.StatusNoContent // 204

	// Client error responses
	StatusBadRequest      = http.StatusBadRequest      // 400
	StatusUnauthorized    = http.StatusUnauthorized    // 401
	StatusForbidden       = http.StatusForbidden       // 403
	StatusNotFound        = http.StatusNotFound        // 404
	StatusConflict        = http.StatusConflict        // 409
	StatusTooManyRequests = http.StatusTooManyRequests // 429

	// Server error responses
	StatusInternalServerError = http.StatusInternalServerError // 500
	StatusBadGateway          = http.StatusBadGateway          // 502
	StatusServiceUnavailable  = http.StatusServiceUnavailable  // 503
	StatusGatewayTimeout      = http.StatusGatewayTimeout      // 504

	// Redirect responses
	StatusFound    = http.StatusFound    // 302
	StatusSeeOther = http.StatusSeeOther // 303
)

// Application Paths
const (
	// Public paths
	PathLogin          = "/login"
	PathSignup         = "/signup"
	PathHealth         = "/health"
	PathMetrics        = "/metrics"
	PathForgotPassword = "/forgot-password"
	PathResetPassword  = "/reset-password"
	PathVerifyEmail    = "/verify-email"

	// Authenticated paths
	PathDashboard = "/dashboard"
	PathForms     = "/forms"
	PathProfile   = "/profile"
	PathSettings  = "/settings"

	// Admin paths
	PathAdmin      = "/admin"
	PathAdminUsers = "/admin/users"
	PathAdminForms = "/admin/forms"

	// API paths
	PathAPIv1               = "/api/v1"
	PathAPIValidation       = "/api/v1/validation"
	PathAPIValidationLogin  = "/api/v1/validation/login"
	PathAPIValidationSignup = "/api/v1/validation/signup"
	PathAPIPublic           = "/api/v1/public"
	PathAPIHealth           = "/api/v1/health"
	PathAPIMetrics          = "/api/v1/metrics"
	PathAPIForms            = "/api/v1/forms"
	PathAPIFormsLaravel     = "/api/forms"
	PathFormsPublic         = "/forms" // Public embed routes: /forms/:id/embed, schema, submit
	PathAPIAdmin            = "/api/v1/admin"
	PathAPIAdminUsers       = "/api/v1/admin/users"
	PathAPIAdminForms       = "/api/v1/admin/forms"

	// Static asset paths
	PathStatic    = "/static"
	PathAssets    = "/assets"
	PathImages    = "/images"
	PathCSS       = "/css"
	PathJS        = "/js"
	PathFonts     = "/fonts"
	PathFavicon   = "/favicon.ico"
	PathRobotsTxt = "/robots.txt"

	PathLoginPost  = "/login"
	PathSignupPost = "/signup"
	PathLogout     = "/logout"
	PathAPIV1      = "/api/v1"
	PathValidation = "/validation"
)

// Timeouts and Intervals
const (
	// Session timeouts
	SessionExpiryHours     = 24
	SessionIDLength        = 32
	SessionTimeout         = 5 // seconds
	SessionCleanupInterval = 1 // hour

	// Rate limiting
	RateLimitBurst   = 5
	DefaultRateLimit = 20
	RateLimitWindow  = 60 // seconds
)

// Headers
const (
	HeaderContentType    = "Content-Type"
	HeaderXRequestedWith = "X-Requested-With"
	HeaderXMLHttpRequest = "XMLHttpRequest"
	HeaderAuthorization  = "Authorization"
	HeaderUserAgent      = "User-Agent"
	HeaderXForwardedFor  = "X-Forwarded-For"
	HeaderXRealIP        = "X-Real-IP"
)

// Content Types
const (
	ContentTypeJSON = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeHTML = "text/html"
	ContentTypeText = "text/plain"
	ContentTypeIcon = "image/x-icon"
)

// Cookie Names
const (
	CookieSession = "session"
	CookieCSRF    = "csrf_token"
)

// Error Messages
const (
	ErrMsgInvalidRequest     = "Invalid request format"
	ErrMsgInvalidCredentials = "Invalid email or password"
	ErrMsgAccountCreated     = "Account created successfully!"
	ErrMsgServiceUnhealthy   = "Service is not healthy"
	ErrMsgDatabaseError      = "Database connection failed"
	ErrMsgInternalError      = "Internal server error"
	ErrMsgNotFound           = "Resource not found"
	ErrMsgUnauthorized       = "Unauthorized access"
	ErrMsgForbidden          = "Access forbidden"
	ErrMsgTooManyRequests    = "Too many requests"
	ErrMsgValidationFailed   = "Validation failed"
)

// Success Messages
const (
	MsgLoginSuccess       = "Login successful"
	MsgLogoutSuccess      = "Logout successful"
	MsgSignupSuccess      = "Account created successfully!"
	MsgFormCreated        = "Form created successfully"
	MsgFormUpdated        = "Form updated successfully"
	MsgFormDeleted        = "Form deleted successfully"
	MsgSubmissionReceived = "Form submission received"
)

// Default Values
const (
	DefaultPageSize        = 20
	MaxPageSize            = 100
	DefaultFormTitle       = "Untitled Form"
	DefaultFormDescription = "No description provided"
	MaxFormSchemaSize      = 1024 * 1024 // 1MB
)

// Validation Rules
const (
	MinPasswordLength        = 8
	MaxPasswordLength        = 128
	MinEmailLength           = 3
	MaxEmailLength           = 254
	MinNameLength            = 1
	MaxNameLength            = 100
	MinFormTitleLength       = 1
	MaxFormTitleLength       = 255
	MinFormDescriptionLength = 0
	MaxFormDescriptionLength = 1000
)

// File Upload Limits
const (
	MaxFileSize          = 10 * 1024 * 1024 // 10MB
	MaxFilesPerRequest   = 10
	AllowedImageTypes    = "image/jpeg,image/png,image/gif,image/webp"
	AllowedDocumentTypes = "application/pdf,application/msword," +
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

// Security Constants
const (
	CSRFTokenLength         = 32
	SessionTokenLength      = 64
	PasswordHashCost        = 12
	MaxLoginAttempts        = 5
	LoginLockoutDuration    = 15 // minutes
	PasswordResetExpiry     = 24 // hours
	EmailVerificationExpiry = 24 // hours
)

// Environment Constants
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// Database Constants
const (
	MaxDBConnections      = 100
	MaxIdleConnections    = 10
	ConnectionMaxLifetime = 300 // seconds
	ConnectionMaxIdleTime = 60  // seconds
)

// Logging Constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"

	MaxLogMessageLength = 1000
	MaxLogFieldLength   = 500
)

// Cache Constants
const (
	CacheTTLDefault = 300  // seconds
	CacheTTLShort   = 60   // seconds
	CacheTTLLong    = 3600 // seconds
	CacheMaxSize    = 1000
)

// Pagination Constants
const (
	DefaultPage  = 1
	MinPage      = 1
	MaxPage      = 10000
	DefaultLimit = 20
	MinLimit     = 1
	MaxLimit     = 100
)

// Form Constants
const (
	FormStatusDraft     = "draft"
	FormStatusPublished = "published"
	FormStatusArchived  = "archived"

	SubmissionStatusPending    = "pending"
	SubmissionStatusCompleted  = "completed"
	SubmissionStatusFailed     = "failed"
	SubmissionStatusProcessing = "processing"
)

// User Constants
const (
	UserRoleUser      = "user"
	UserRoleAdmin     = "admin"
	UserRoleModerator = "moderator"

	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusPending   = "pending"
)

// Middleware Constants
const (
	// NonceSize is the size of the nonce in bytes (32 bytes = 256 bits)
	NonceSize = 32
	// HSTSOneYear is the number of seconds in one year
	HSTSOneYear = 31536000
	// DefaultTokenLength is the default length for generated tokens
	DefaultTokenLength = 32
	// CookieMaxAge is the maximum age of cookies in seconds (24 hours)
	CookieMaxAge = 86400
	// FieldPairSize represents the number of elements in a key-value pair
	FieldPairSize = 2
	// DefaultUnknown is the default value for unknown identifiers
	DefaultUnknown = "unknown"
	// DefaultMiddlewarePriority is the default priority for middleware
	DefaultMiddlewarePriority = 50
)

// Middleware Priority Constants
const (
	PriorityRecovery        = 10
	PriorityCORS            = 20
	PriorityRequestID       = 30
	PriorityTimeout         = 40
	PrioritySecurityHeaders = 50
	PriorityCSRF            = 60
	PriorityRateLimit       = 70
	PriorityInputValidation = 80
	PriorityLogging         = 90
	PrioritySession         = 100
	PriorityAuthentication  = 110
	PriorityAuthorization   = 120
)

// Configuration Time Constants (in seconds)
const (
	TimeoutShort      = 5
	TimeoutDefault    = 30
	TimeoutMedium     = 60
	TimeoutAuth       = 15
	TimeoutPublic     = 10
	SessionExpiry     = 3600  // 1 hour
	RefreshTimeout    = 300   // 5 minutes
	TokenExpiry       = 3600  // 1 hour
	RefreshExpiry     = 86400 // 24 hours
	CacheTTLShortSec  = 300   // 5 minutes
	CacheDurationDay  = 86400 // 24 hours
	RequestsPerMinute = 60
	BurstSizeDefault  = 10
	WindowSizeDefault = 60
	GracePeriod       = 5
)

// StaticFileExtensions contains all supported static file extensions
var StaticFileExtensions = []string{
	".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".ico",
	".svg", ".woff", ".woff2", ".ttf", ".eot",
}
