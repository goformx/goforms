package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Rate limit configuration
const (
	authRateLimit  = 5  // requests per second
	authBurstLimit = 10 // maximum burst size
	authWindowSize = 1 * time.Minute
)

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	userService user.Service
	secret      string
	logger      logging.Logger
	config      *config.Config
	limiter     *rate.Limiter
}

// TokenClaims represents the JWT claims
type TokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Type   string `json:"type"` // "access" or "refresh"
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(
	userService user.Service,
	secret string,
	logger logging.Logger,
	cfg *config.Config,
) echo.MiddlewareFunc {
	if userService == nil {
		panic("JWTMiddleware initialization failed: userService is required")
	}
	if logger == nil {
		panic("JWTMiddleware initialization failed: logger is required")
	}
	if cfg == nil {
		panic("JWTMiddleware initialization failed: config is required")
	}
	if secret == "" {
		panic("JWTMiddleware initialization failed: secret is required")
	}

	m := &JWTMiddleware{
		userService: userService,
		secret:      secret,
		logger:      logger,
		config:      cfg,
		limiter:     rate.NewLimiter(rate.Limit(authRateLimit), authBurstLimit),
	}

	// Configure Echo's JWT middleware
	jwtConfig := echojwt.Config{
		SigningKey:  []byte(secret),
		TokenLookup: "header:Authorization",
		ContextKey:  "jwt",
		ErrorHandler: func(c echo.Context, err error) error {
			return m.handleAuthError(c, err)
		},
		SuccessHandler: func(c echo.Context) {
			// Extract claims and set user in context
			token := c.Get("jwt").(*jwt.Token)
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				c.Error(echo.NewHTTPError(http.StatusForbidden, "invalid token claims format"))
				return
			}

			userID, err := extractUserID(claims)
			if err != nil {
				c.Error(echo.NewHTTPError(http.StatusForbidden, err.Error()))
				return
			}

			// Get user from service
			userData, err := m.userService.GetByID(c.Request().Context(), userID)
			if err != nil {
				c.Error(echo.NewHTTPError(http.StatusForbidden, "user not found or inactive"))
				return
			}

			// Set user in context
			c.Set("user", userData)
		},
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			return m.isAuthExempt(path) || m.isPublicAPI(path)
		},
	}

	return echojwt.WithConfig(jwtConfig)
}

// isAuthExempt checks if the path is exempt from authentication
func (m *JWTMiddleware) isAuthExempt(path string) bool {
	return strings.HasPrefix(path, "/"+m.config.Static.DistDir+"/") ||
		path == "/favicon.ico" || path == "/robots.txt" ||
		strings.HasPrefix(path, "/api/validation/") ||
		strings.HasPrefix(path, "/login") || strings.HasPrefix(path, "/signup") ||
		strings.HasPrefix(path, "/forgot-password") || strings.HasPrefix(path, "/contact") ||
		strings.HasPrefix(path, "/demo")
}

// isPublicAPI checks if the path is for a public API endpoint
func (m *JWTMiddleware) isPublicAPI(path string) bool {
	return strings.HasPrefix(path, "/api/v1/forms/") &&
		(strings.HasSuffix(path, "/schema") || strings.HasSuffix(path, "/submit"))
}

// handleAuthError handles authentication errors
func (m *JWTMiddleware) handleAuthError(c echo.Context, err error) error {
	// Extract status code from HTTPError if available
	status := http.StatusUnauthorized
	var he *echo.HTTPError
	if errors.As(err, &he) {
		status = he.Code
	}

	if c != nil {
		m.logger.Error("auth check failed",
			logging.StringField("path", c.Path()),
			logging.StringField("method", c.Request().Method),
			logging.IntField("status", status),
			logging.ErrorField("error", err))
	}
	return echo.NewHTTPError(status, err.Error())
}

// extractUserID extracts and validates the user ID from JWT claims
func extractUserID(claims jwt.MapClaims) (string, error) {
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", errors.New("invalid or missing user_id claim")
	}
	return userID, nil
}

// GenerateTokenPair generates a new pair of access and refresh tokens
func (m *JWTMiddleware) GenerateTokenPair(userID string) (*TokenPair, error) {
	// Generate access token
	accessClaims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
		Type:   "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(m.secret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshClaims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
		Type:   "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(m.secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (m *JWTMiddleware) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	// Parse and validate refresh token
	token, err := jwt.ParseWithClaims(refreshTokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid || claims.Type != "refresh" {
		return nil, errors.New("invalid refresh token")
	}

	// Generate new token pair
	return m.GenerateTokenPair(claims.UserID)
}

// SecurityHeaders adds security-related headers to responses
func (m *JWTMiddleware) SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			return next(c)
		}
	}
}
