package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"strings"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

const (
	// NonceSize is the size of the nonce in bytes (32 bytes = 256 bits)
	NonceSize = 32
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger      logging.Logger
	config      *ManagerConfig
	securityMgr *SecurityManager
}

// ManagerConfig holds middleware configuration
type ManagerConfig struct {
	Logger      logging.Logger
	JWTSecret   string
	UserService any
	EnableCSRF  bool
	CSRF        CSRFMiddlewareConfig
	Security    SecurityConfig // New security configuration
}

// New creates a new middleware manager
func New(config *ManagerConfig) *Manager {
	if config.Logger == nil {
		panic("logger is required for Manager")
	}

	// Initialize security manager with configuration
	securityMgr := NewSecurityManager(SecurityConfig{
		Logger:           config.Logger,
		CSPConfig:        getDefaultCSPConfig(),
		HeadersConfig:    getDefaultSecurityHeaders(),
		DangerousHeaders: getDefaultDangerousHeaders(),
	})

	return &Manager{
		logger:      config.Logger,
		config:      config,
		securityMgr: securityMgr,
	}
}

// Setup configures all middleware for an Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Debug("setting up middleware")

	// Basic middleware
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.Secure())
	e.Use(echomw.BodyLimit("2M"))

	// CORS
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-CSRF-Token",
		},
	}))

	// Security headers and CSP
	e.Use(m.securityMgr.SecurityMiddleware())

	// CSRF if enabled
	if m.config != nil && m.config.EnableCSRF {
		csrfConfig := m.config.CSRF
		if csrfConfig.Logger == nil {
			csrfConfig.Logger = m.logger
		}
		e.Use(CSRF(csrfConfig))
		e.Use(CSRFToken())
	}

	m.logger.Debug("middleware setup complete")
}

// CSRF returns CSRF middleware with the given configuration
func (m *Manager) CSRF(config CSRFConfig) echo.MiddlewareFunc {
	m.logger.Debug("creating CSRF middleware",
		logging.Bool("secure", config.Secure),
		logging.String("cookie_name", "csrf_token"),
	)

	csrfMiddleware := csrf.Protect(
		[]byte(config.SecretKey),
		csrf.Secure(config.Secure),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.logger.Error("CSRF validation failed",
				logging.String("path", r.URL.Path),
				logging.String("method", r.Method),
			)
			http.Error(w, "CSRF validation failed", http.StatusForbidden)
		})),
	)

	return echo.WrapMiddleware(func(next http.Handler) http.Handler {
		return csrfMiddleware(next)
	})
}

// CSRFToken returns middleware to add CSRF token to templates
func (m *Manager) CSRFToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := csrf.Token(c.Request())
			c.Set("csrf", token)
			return next(c)
		}
	}
}

// CSRFConfig holds configuration for CSRF middleware
type CSRFConfig struct {
	SecretKey string
	Secure    bool
}

// Helper functions to provide default configurations

func getDefaultCSPConfig() CSPConfig {
	return CSPConfig{
		DefaultSrc:     []string{"'self'"},
		ScriptSrc:      []string{"'self'", "https://cdn.jsdelivr.net"},
		StyleSrc:       []string{"'self'", "'unsafe-inline'"},
		ImgSrc:         []string{"'self'", "data:"},
		FontSrc:        []string{"'self'"},
		ConnectSrc:     []string{"'self'"},
		MediaSrc:       []string{"'self'"},
		ObjectSrc:      []string{"'none'"},
		ChildSrc:       []string{"'none'"},
		FrameAncestors: []string{"'none'"},
		FormAction:     []string{"'self'"},
		BaseURI:        []string{"'self'"},
		ManifestSrc:    []string{"'self'"},
		Upgrades:       true,
		BlockMixed:     true,
	}
}

func getDefaultSecurityHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Permissions-Policy": strings.Join([]string{
			"accelerometer=()",
			"camera=()",
			"geolocation=()",
			"gyroscope=()",
			"magnetometer=()",
			"microphone=()",
			"payment=()",
			"usb=()",
		}, ", "),
		"Cross-Origin-Opener-Policy":   "same-origin",
		"Cross-Origin-Embedder-Policy": "require-corp",
		"Cross-Origin-Resource-Policy": "same-origin",
		"Strict-Transport-Security":    "max-age=31536000; includeSubDomains",
		"Cache-Control":                "no-store, max-age=0",
		"Clear-Site-Data":              "\"cache\",\"cookies\",\"storage\"",
	}
}

func getDefaultDangerousHeaders() []string {
	return []string{
		"Server",
		"X-Powered-By",
		"X-AspNet-Version",
		"X-AspNetMvc-Version",
	}
}
