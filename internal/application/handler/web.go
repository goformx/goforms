package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	amw "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

var (
	// ErrNoCurrentUser is returned when no user is found in the current context
	ErrNoCurrentUser = errors.New("no current user found")
)

// Signup validation constants (for linter compliance)
const (
	SignupFirstNameMinValue = 2
	SignupFirstNameMaxValue = 50
	SignupLastNameMinValue  = 2
	SignupLastNameMaxValue  = 50
	SignupEmailMinValue     = 5
	SignupEmailMaxValue     = 100
	SignupPasswordMinValue  = 8
	SignupPasswordMaxValue  = 100
)

// SignupValidation holds validation constants for signup
type SignupValidation struct {
	FirstNameMin int
	FirstNameMax int
	LastNameMin  int
	LastNameMax  int
	EmailMin     int
	EmailMax     int
	PasswordMin  int
	PasswordMax  int
}

var signupValidation = SignupValidation{
	FirstNameMin: SignupFirstNameMinValue,
	FirstNameMax: SignupFirstNameMaxValue,
	LastNameMin:  SignupLastNameMinValue,
	LastNameMax:  SignupLastNameMaxValue,
	EmailMin:     SignupEmailMinValue,
	EmailMax:     SignupEmailMaxValue,
	PasswordMin:  SignupPasswordMinValue,
	PasswordMax:  SignupPasswordMaxValue,
}

// WebHandlerOption defines a web handler option.
// This type is used to implement the functional options pattern
// for configuring the WebHandler.
type WebHandlerOption func(*WebHandler)

// WithRenderer sets the view renderer.
// This is a required option for the WebHandler as it needs
// the renderer to display web pages.
//
// Example:
//
//	handler := NewWebHandler(logger, WithRenderer(renderer))
func WithRenderer(renderer *view.Renderer) WebHandlerOption {
	return func(h *WebHandler) {
		h.renderer = renderer
	}
}

// WithMiddlewareManager sets the middleware manager for the web handler.
func WithMiddlewareManager(manager *amw.Manager) WebHandlerOption {
	return func(h *WebHandler) {
		h.middlewareManager = manager
	}
}

// WithConfig sets the config for the web handler.
func WithConfig(cfg *config.Config) WebHandlerOption {
	return func(h *WebHandler) {
		h.config = cfg
	}
}

// WebHandler handles web page requests.
// It requires a renderer, and subscription service to function properly.
// Use the functional options pattern to configure these dependencies.
//
// Dependencies:
//   - renderer: Required for rendering web pages
//   - middlewareManager: Required for security and request processing
//   - config: Required for configuration
type WebHandler struct {
	Base
	renderer          *view.Renderer
	middlewareManager *amw.Manager
	config            *config.Config
}

// NewWebHandler creates a new web handler.
// It uses the functional options pattern to configure the handler.
// The logger is required as a direct parameter, while other dependencies
// are provided through options.
//
// Example:
//
//	handler := NewWebHandler(logger,
//	    WithRenderer(renderer),
//	    WithConfig(config),
//	)
func NewWebHandler(logger logging.Logger, renderer *view.Renderer, opts ...WebHandlerOption) (*WebHandler, error) {
	h := &WebHandler{
		Base:     NewBase(WithLogger(logger)),
		renderer: renderer,
	}

	for _, opt := range opts {
		opt(h)
	}

	// Validate critical dependencies during construction
	if h.renderer == nil {
		return nil, errors.New("WebHandler initialization failed: renderer is required")
	}
	if h.middlewareManager == nil {
		return nil, errors.New("WebHandler initialization failed: middleware manager is required")
	}
	if h.config == nil {
		return nil, errors.New("WebHandler initialization failed: config is required")
	}

	return h, nil
}

// Validate validates that required dependencies are set.
// Returns an error if any required dependency is missing.
//
// Required dependencies:
//   - renderer
//   - middlewareManager
//   - config
func (h *WebHandler) Validate() error {
	if err := h.Base.Validate(); err != nil {
		return fmt.Errorf("WebHandler validation failed: %w", err)
	}
	if h.renderer == nil {
		return errors.New("WebHandler validation failed: renderer is required")
	}
	if h.middlewareManager == nil {
		return errors.New("WebHandler validation failed: middleware manager is required")
	}
	if h.config == nil {
		return errors.New("WebHandler validation failed: config is required")
	}
	return nil
}

// getCSRFToken retrieves the CSRF token from the context
func (h *WebHandler) getCSRFToken(c echo.Context) (string, error) {
	token := c.Get("csrf")
	if token == nil {
		return "", errors.New("CSRF token not found in context")
	}

	tokenStr, ok := token.(string)
	if !ok {
		return "", errors.New("invalid CSRF token type")
	}

	if tokenStr == "" {
		return "", errors.New("empty CSRF token")
	}

	return tokenStr, nil
}

// renderPage renders a page with the given title and content
func (h *WebHandler) renderPage(c echo.Context, title string, template func(shared.PageData) templ.Component) error {
	// Get CSRF token from context
	csrfToken, err := h.getCSRFToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
	}

	// Get user from context if available
	var userData *user.User
	if u := c.Get("user"); u != nil {
		if userObj, ok := u.(*user.User); ok {
			userData = userObj
		}
	}

	// Create page data
	data := shared.PageData{
		Title:         title,
		CSRFToken:     csrfToken,
		User:          userData,
		IsDevelopment: h.config.App.IsDevelopment(),
	}

	// Render page
	if renderErr := template(data).Render(c.Request().Context(), c.Response().Writer); renderErr != nil {
		h.Logger().Error("failed to render template",
			logging.StringField("title", title),
			logging.StringField("path", c.Request().URL.Path),
			logging.ErrorField("error", renderErr))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render page")
	}

	return nil
}

// route defines a route for registration
type route struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
}

// registerRoutes registers all web routes using the route struct
func (h *WebHandler) registerRoutes(e *echo.Echo) {
	routes := []route{
		{"GET", "/", h.handleHome},
		{"GET", "/demo", h.handleDemo},
		{"GET", "/signup", h.handleSignup},
		{"GET", "/login", h.handleLogin},
		{"GET", "/api/validation/:schema", h.handleValidationSchema},
	}
	for _, r := range routes {
		e.Add(r.Method, r.Path, r.Handler)
		if h.config.App.IsDevelopment() {
			h.Logger().Debug("web handler called",
				logging.StringField("method", r.Method),
				logging.StringField("path", r.Path))
		}
	}
	// Static files
	e.Static("/"+h.config.Static.DistDir, h.config.Static.DistDir)
	e.File("/favicon.ico", "./public/favicon.ico")
}

// validateDependencies validates required dependencies for the handler
func (h *WebHandler) validateDependencies() {
	if err := h.Validate(); err != nil {
		h.Logger().Error("failed to validate web handler", logging.ErrorField("error", err))
	}
}

// Register registers the web routes (SRP: now just calls validateDependencies and registerRoutes)
func (h *WebHandler) Register(e *echo.Echo) {
	h.validateDependencies()
	if h.config.App.IsDevelopment() {
		h.Logger().Debug("registering web routes")
	}
	h.registerRoutes(e)
	if h.config.App.IsDevelopment() {
		h.Logger().Debug("web routes registration complete")
	}
}

// handleHome renders the home page
func (h *WebHandler) handleHome(c echo.Context) error {
	return h.renderPage(c, "Home", pages.Home)
}

// handleDemo renders the demo page
func (h *WebHandler) handleDemo(c echo.Context) error {
	return h.renderPage(c, "Demo", pages.Demo)
}

// handleSignup renders the signup page
func (h *WebHandler) handleSignup(c echo.Context) error {
	return h.renderPage(c, "Sign Up", pages.Signup)
}

// handleLogin renders the login page
func (h *WebHandler) handleLogin(c echo.Context) error {
	return h.renderPage(c, "Login", pages.Login)
}

// schemaBuilders maps schema names to their builder functions
var schemaBuilders = map[string]func() map[string]any{
	"signup": buildSignupSchema,
	"login":  buildLoginSchema,
}

func (h *WebHandler) handleValidationSchema(c echo.Context) error {
	schemaName := c.Param("schema")
	builder, ok := schemaBuilders[schemaName]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "validation schema not found"})
	}
	return c.JSON(http.StatusOK, builder())
}

func buildSignupSchema() map[string]any {
	return map[string]any{
		"first_name": map[string]any{
			"type":    "string",
			"min":     signupValidation.FirstNameMin,
			"max":     signupValidation.FirstNameMax,
			"message": "First name must be between 2 and 50 characters",
		},
		"last_name": map[string]any{
			"type":    "string",
			"min":     signupValidation.LastNameMin,
			"max":     signupValidation.LastNameMax,
			"message": "Last name must be between 2 and 50 characters",
		},
		"email": map[string]any{
			"type":    "email",
			"min":     signupValidation.EmailMin,
			"max":     signupValidation.EmailMax,
			"message": "Please enter a valid email address",
		},
		"password": map[string]any{
			"type":    "password",
			"min":     signupValidation.PasswordMin,
			"max":     signupValidation.PasswordMax,
			"message": "Password must be at least 8 characters and contain upper, lower, number, special",
		},
		"confirm_password": map[string]any{
			"type":       "match",
			"matchField": "password",
			"message":    "Passwords don't match",
		},
	}
}

func buildLoginSchema() map[string]any {
	return map[string]any{
		"email": map[string]any{
			"type":    "email",
			"message": "Please enter a valid email address",
		},
		"password": map[string]any{
			"type":    "string",
			"message": "Password is required",
		},
	}
}
