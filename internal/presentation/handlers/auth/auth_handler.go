package auth

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related routes
// Implements httpiface.Handler
type AuthHandler struct {
	handlers.BaseHandler
	userService     user.Service
	sessionManager  *session.Manager
	renderer        view.Renderer
	config          *config.Config
	assetManager    web.AssetManagerInterface
	logger          logging.Logger
	requestParser   *AuthRequestParser
	responseBuilder *AuthResponseBuilder
}

// NewAuthHandler creates a new AuthHandler and registers all auth routes
func NewAuthHandler(
	userService user.Service,
	sessionManager *session.Manager,
	renderer view.Renderer,
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	logger logging.Logger,
) *AuthHandler {
	h := &AuthHandler{
		BaseHandler:     *handlers.NewBaseHandler("auth"),
		userService:     userService,
		sessionManager:  sessionManager,
		renderer:        renderer,
		config:          cfg,
		assetManager:    assetManager,
		logger:          logger,
		requestParser:   NewAuthRequestParser(),
		responseBuilder: NewAuthResponseBuilder(cfg, assetManager, renderer, logger),
	}

	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/login",
		Handler: h.Login,
	})
	h.AddRoute(httpiface.Route{
		Method:  "POST",
		Path:    "/login",
		Handler: h.LoginPost,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/signup",
		Handler: h.Signup,
	})
	h.AddRoute(httpiface.Route{
		Method:  "POST",
		Path:    "/signup",
		Handler: h.SignupPost,
	})
	h.AddRoute(httpiface.Route{
		Method:  "POST",
		Path:    "/logout",
		Handler: h.Logout,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/api/v1/test",
		Handler: h.TestEndpoint,
	})

	return h
}

// Login handles GET /login
func (h *AuthHandler) Login(ctx httpiface.Context) error {
	// Extract the underlying Echo context for rendering
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Create page data for the login template
	pageData := view.NewPageData(h.config, h.assetManager, echoCtx, "Login")

	// Render the login page using the generated template
	loginComponent := pages.Login(*pageData)

	if err := h.renderer.Render(echoCtx, loginComponent); err != nil {
		return fmt.Errorf("failed to render login page: %w", err)
	}

	return nil
}

// LoginPost handles POST /login
func (h *AuthHandler) LoginPost(ctx httpiface.Context) error {
	// Extract the underlying Echo context for form parsing and session management
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Parse login credentials using the request parser
	email, password, err := h.requestParser.ParseLogin(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse login request", "error", err)

		return h.responseBuilder.BuildLoginErrorResponse(echoCtx, "Invalid request format")
	}

	// Validate login credentials
	if validateErr := h.requestParser.ValidateLogin(email, password); validateErr != nil {
		return h.responseBuilder.BuildValidationErrorResponse(echoCtx, "credentials", validateErr.Error())
	}

	// Attempt login using user service
	loginRequest := &user.Login{
		Email:    email,
		Password: password,
	}

	loginResponse, err := h.userService.Login(echoCtx.Request().Context(), loginRequest)
	if err != nil {
		h.logger.Warn("login failed", "email", email, "error", err)

		return h.responseBuilder.BuildLoginErrorResponse(echoCtx, "Invalid email or password")
	}

	// Create session
	sessionID, err := h.sessionManager.CreateSession(
		loginResponse.User.ID,
		loginResponse.User.Email,
		loginResponse.User.Role,
	)
	if err != nil {
		h.logger.Error("failed to create session", "user_id", loginResponse.User.ID, "error", err)

		return h.responseBuilder.BuildLoginErrorResponse(echoCtx, "Failed to create session. Please try again.")
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(echoCtx, sessionID)

	// Build success response
	return h.responseBuilder.BuildLoginSuccessResponse(echoCtx, loginResponse.User)
}

// Signup handles GET /signup
func (h *AuthHandler) Signup(ctx httpiface.Context) error {
	// Extract the underlying Echo context for rendering
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Create page data for the signup template
	pageData := view.NewPageData(h.config, h.assetManager, echoCtx, "Sign Up")

	// Render the signup page using the generated template
	signupComponent := pages.Signup(*pageData)

	if err := h.renderer.Render(echoCtx, signupComponent); err != nil {
		return fmt.Errorf("failed to render signup page: %w", err)
	}

	return nil
}

// SignupPost handles POST /signup
func (h *AuthHandler) SignupPost(ctx httpiface.Context) error {
	// Extract the underlying Echo context for form parsing and session management
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Parse signup data using the request parser
	signupRequest, err := h.requestParser.ParseSignup(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse signup request", "error", err)

		return h.responseBuilder.BuildSignupErrorResponse(echoCtx, "Invalid request format")
	}

	// Validate signup data
	if validateErr := h.requestParser.ValidateSignup(signupRequest); validateErr != nil {
		return h.responseBuilder.BuildValidationErrorResponse(echoCtx, "signup", validateErr.Error())
	}

	// Additional password strength validation
	if len(signupRequest.Password) < constants.MinPasswordLength {
		return h.responseBuilder.BuildValidationErrorResponse(echoCtx, "password",
			fmt.Sprintf("Password must be at least %d characters long", constants.MinPasswordLength))
	}

	// Attempt signup using user service
	newUser, err := h.userService.SignUp(echoCtx.Request().Context(), &signupRequest)
	if err != nil {
		h.logger.Warn("signup failed", "email", signupRequest.Email, "error", err)

		// Handle specific errors
		if err.Error() == "user already exists" {
			return h.responseBuilder.BuildSignupErrorResponse(echoCtx, "An account with this email already exists")
		}

		return h.responseBuilder.BuildSignupErrorResponse(echoCtx, "Failed to create account. Please try again.")
	}

	// Create session for the new user
	sessionID, err := h.sessionManager.CreateSession(
		newUser.ID,
		newUser.Email,
		newUser.Role,
	)
	if err != nil {
		h.logger.Error("failed to create session for new user", "user_id", newUser.ID, "error", err)

		return h.responseBuilder.BuildSignupErrorResponse(
			echoCtx,
			"Account created but failed to log you in. Please try logging in.",
		)
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(echoCtx, sessionID)

	// Build success response
	return h.responseBuilder.BuildSignupSuccessResponse(echoCtx, newUser)
}

// Logout handles POST /logout
func (h *AuthHandler) Logout(ctx httpiface.Context) error {
	// Extract the underlying Echo context for session management
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Get session cookie
	cookie, err := echoCtx.Cookie(h.sessionManager.GetCookieName())
	if err == nil && cookie.Value != "" {
		// Delete the session
		h.sessionManager.DeleteSession(cookie.Value)
		h.logger.Info("user logged out", "session_id", cookie.Value)
	}

	// Clear session cookie
	h.sessionManager.ClearSessionCookie(echoCtx)

	// Build success response
	return h.responseBuilder.BuildLogoutSuccessResponse(echoCtx)
}

// TestEndpoint handles GET /api/v1/test
func (h *AuthHandler) TestEndpoint(ctx httpiface.Context) error {
	return fmt.Errorf("test endpoint working (placeholder)")
}
