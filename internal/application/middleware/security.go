package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// SecurityConfig holds configuration for security middleware
type SecurityConfig struct {
	Logger           logging.Logger
	StaticPaths      []string          // Configurable static paths
	StaticExtensions []string          // Configurable static extensions
	CSPConfig        CSPConfig         // Default CSP configuration
	HeadersConfig    map[string]string // Configurable security headers
	DangerousHeaders []string          // Headers to remove
}

// CSPConfig holds Content Security Policy configuration
type CSPConfig struct {
	DefaultSrc     []string
	ScriptSrc      []string
	StyleSrc       []string
	ImgSrc         []string
	FontSrc        []string
	ConnectSrc     []string
	MediaSrc       []string
	ObjectSrc      []string
	ChildSrc       []string
	FrameAncestors []string
	FormAction     []string
	BaseURI        []string
	ManifestSrc    []string
	Upgrades       bool // For upgrade-insecure-requests
	BlockMixed     bool // For block-all-mixed-content
}

// SecurityManager handles security-related middleware
type SecurityManager struct {
	logger logging.Logger
	config SecurityConfig
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config SecurityConfig) *SecurityManager {
	if config.Logger == nil {
		panic("logger is required for SecurityManager")
	}

	// Set default static paths if none provided
	if len(config.StaticPaths) == 0 {
		config.StaticPaths = []string{"/static/", "/favicon.ico"}
	}

	// Set default static extensions if none provided
	if len(config.StaticExtensions) == 0 {
		config.StaticExtensions = []string{
			".js", ".css", ".png", ".jpg", ".jpeg",
			".gif", ".svg", ".ico", ".woff", ".woff2",
		}
	}

	return &SecurityManager{
		logger: config.Logger,
		config: config,
	}
}

// generateNonce creates a cryptographically secure random nonce
func (sm *SecurityManager) generateNonce() (string, error) {
	nonceBytes := make([]byte, NonceSize)
	if _, err := rand.Read(nonceBytes); err != nil {
		sm.logger.Error("failed to generate nonce", logging.Error(err))
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

// isStaticAsset checks if the request is for a static asset
func (sm *SecurityManager) isStaticAsset(path string) bool {
	// Check configured static paths
	for _, staticPath := range sm.config.StaticPaths {
		if strings.HasPrefix(path, staticPath) {
			return true
		}
	}

	// Check configured static extensions
	for _, ext := range sm.config.StaticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// buildCSPValue builds a CSP value string from config
func buildCSPValue(config CSPConfig, nonce string) string {
	directives := make([]string, 0)

	// Helper to build directive
	addDirective := func(name string, sources []string) {
		if len(sources) > 0 {
			directives = append(directives, fmt.Sprintf("%s %s", name, strings.Join(sources, " ")))
		}
	}

	addDirective("default-src", config.DefaultSrc)

	// Add nonce to script-src if provided
	scriptSrc := make([]string, 0, len(config.ScriptSrc)+1) // Pre-allocate capacity for nonce
	scriptSrc = append(scriptSrc, config.ScriptSrc...)
	if nonce != "" {
		scriptSrc = append(scriptSrc, fmt.Sprintf("'nonce-%s'", nonce))
	}
	addDirective("script-src", scriptSrc)

	addDirective("style-src", config.StyleSrc)
	addDirective("img-src", config.ImgSrc)
	addDirective("font-src", config.FontSrc)
	addDirective("connect-src", config.ConnectSrc)
	addDirective("media-src", config.MediaSrc)
	addDirective("object-src", config.ObjectSrc)
	addDirective("child-src", config.ChildSrc)
	addDirective("frame-ancestors", config.FrameAncestors)
	addDirective("form-action", config.FormAction)
	addDirective("base-uri", config.BaseURI)
	addDirective("manifest-src", config.ManifestSrc)

	if config.Upgrades {
		directives = append(directives, "upgrade-insecure-requests")
	}
	if config.BlockMixed {
		directives = append(directives, "block-all-mixed-content")
	}

	return strings.Join(directives, "; ")
}

// SecurityMiddleware returns the security middleware handler
func (sm *SecurityManager) SecurityMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			var csp string

			if sm.isStaticAsset(path) {
				sm.logger.Debug("using static CSP for asset",
					logging.String("path", path))
				csp = buildCSPValue(sm.config.CSPConfig, "")
			} else {
				nonce, err := sm.generateNonce()
				if err != nil {
					sm.logger.Error("failed to generate nonce", logging.Error(err))
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate security nonce")
				}

				sm.logger.Debug("generated nonce for dynamic content",
					logging.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					logging.String("path", path))

				c.Set("nonce", nonce)
				csp = buildCSPValue(sm.config.CSPConfig, nonce)
			}

			sm.setSecurityHeaders(c, csp)
			return next(c)
		}
	}
}

// setSecurityHeaders sets security headers based on configuration
func (sm *SecurityManager) setSecurityHeaders(c echo.Context, csp string) {
	// Set CSP header first
	c.Response().Header().Set("Content-Security-Policy", csp)

	// Set configured security headers
	for key, value := range sm.config.HeadersConfig {
		sm.logger.Debug("set security header",
			logging.String("header", key),
			logging.String("value", value))
		c.Response().Header().Set(key, value)
	}

	// Remove dangerous headers
	for _, header := range sm.config.DangerousHeaders {
		c.Response().Header().Del(header)
		sm.logger.Debug("removed dangerous header",
			logging.String("header", header))
	}
}
