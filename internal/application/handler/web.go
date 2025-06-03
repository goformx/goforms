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
	"github.com/goformx/goforms/internal/presentation/handlers"
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
//   - userService: Required for user operations
type WebHandler struct {
	*handlers.BaseHandler
	renderer          *view.Renderer
	middlewareManager *amw.Manager
	config            *config.Config
	userService       user.Service
}

// validate validates that required dependencies are set.
// Returns an error if any required dependency is missing.
//
// Required dependencies:
//   - renderer
//   - middlewareManager
//   - config
//   - userService
func (h *WebHandler) validate() error {
	if h.renderer == nil {
		return errors.New("WebHandler initialization failed: renderer is required")
	}
	if h.middlewareManager == nil {
		return errors.New("WebHandler initialization failed: middleware manager is required")
	}
	if h.config == nil {
		return errors.New("WebHandler initialization failed: config is required")
	}
	if h.userService == nil {
		return errors.New("WebHandler initialization failed: user service is required")
	}
	return nil
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
//	    WithUserService(userService),
//	)
func NewWebHandler(logger logging.Logger, renderer *view.Renderer, opts ...WebHandlerOption) (*WebHandler, error) {
	h := &WebHandler{
		BaseHandler: handlers.NewBaseHandler(nil, nil, logger),
		renderer:    renderer,
	}

	for _, opt := range opts {
		opt(h)
	}

	if err := h.validate(); err != nil {
		return nil, err
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
	if err := h.BaseHandler.Validate(); err != nil {
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
	if token, ok := c.Get("csrf").(string); ok && token != "" {
		return token, nil
	}
	return "", errors.New("CSRF token not found or invalid")
}

// buildPageData constructs the shared page data for rendering
func (h *WebHandler) buildPageData(c echo.Context, title string) (shared.PageData, error) {
	csrfToken, err := h.getCSRFToken(c)
	if err != nil {
		return shared.PageData{}, echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
	}

	var data shared.PageData
	if userData, ok := c.Get("user").(*user.User); ok {
		data.User = userData
	}

	data.Title = title
	data.CSRFToken = csrfToken
	data.IsDevelopment = h.config.App.IsDevelopment()

	return data, nil
}

// renderPage renders a page with the given title and content
func (h *WebHandler) renderPage(c echo.Context, title string, template func(shared.PageData) templ.Component) error {
	data, err := h.buildPageData(c, title)
	if err != nil {
		return err
	}

	if renderErr := template(data).Render(c.Request().Context(), c.Response().Writer); renderErr != nil {
		return fmt.Errorf("failed to render template: %w", renderErr)
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
		{"POST", "/signup", h.handleSignupPost},
		{"GET", "/login", h.handleLogin},
		{"GET", "/api/validation/:schema", h.handleValidationSchema},
	}

	for _, r := range routes {
		e.Add(r.Method, r.Path, r.Handler)
		if h.config.App.IsDevelopment() {
			h.LogDebug("web handler registered",
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
		h.LogError("failed to validate web handler", err)
	}
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	h.validateDependencies()
	if h.config.App.IsDevelopment() {
		h.LogDebug("registering web routes")
	}

	h.middlewareManager.Setup(e) // Ensure middleware is loaded properly
	h.registerRoutes(e)

	if h.config.App.IsDevelopment() {
		h.LogDebug("web routes registration complete")
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

// handleSignupPost handles the signup form submission
func (h *WebHandler) handleSignupPost(c echo.Context) error {
	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")

	// Validate password confirmation
	if password != passwordConfirm {
		return h.renderer.Render(c, pages.SignupWithError(shared.PageData{}, "Passwords do not match"))
	}

	// Create signup request
	signup := &user.Signup{
		Email:     email,
		Password:  password,
		FirstName: c.FormValue("first_name"),
		LastName:  c.FormValue("last_name"),
	}

	// Create user
	_, err := h.userService.SignUp(c.Request().Context(), signup)
	if err != nil {
		if errors.Is(err, user.ErrUserExists) {
			return h.renderer.Render(c, pages.SignupWithError(shared.PageData{}, "Email already exists"))
		}
		return h.renderer.Render(c, pages.SignupWithError(shared.PageData{}, "Failed to create account"))
	}

	// Generate tokens
	login := &user.Login{
		Email:    email,
		Password: password,
	}
	tokenPair, err := h.userService.Login(c.Request().Context(), login)
	if err != nil {
		return h.renderer.Render(c, pages.SignupWithError(shared.PageData{}, "Failed to generate authentication tokens"))
	}

	// Set refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
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
