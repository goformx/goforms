package auth

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/adapters/http"
	"github.com/goformx/goforms/internal/application/services"
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
	authService     *services.AuthUseCaseService
	requestAdapter  http.RequestAdapter
	responseAdapter http.ResponseAdapter
	renderer        view.Renderer
	config          *config.Config
	assetManager    web.AssetManagerInterface
	logger          logging.Logger
}

// NewAuthHandler creates a new AuthHandler and registers all auth routes
func NewAuthHandler(
	authService *services.AuthUseCaseService,
	requestAdapter http.RequestAdapter,
	responseAdapter http.ResponseAdapter,
	renderer view.Renderer,
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	logger logging.Logger,
) *AuthHandler {
	h := &AuthHandler{
		BaseHandler:     *handlers.NewBaseHandler("auth"),
		authService:     authService,
		requestAdapter:  requestAdapter,
		responseAdapter: responseAdapter,
		renderer:        renderer,
		config:          cfg,
		assetManager:    assetManager,
		logger:          logger,
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
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse request using adapter
	loginReq, err := h.requestAdapter.ParseLoginRequest(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse login request", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid request format"))
	}

	// Call application service
	loginResp, err := h.authService.Login(echoCtx.Request().Context(), loginReq)
	if err != nil {
		h.logger.Warn("login failed", "email", loginReq.Email, "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid email or password"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildLoginResponse(adapterCtx, loginResp)
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
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse request using adapter
	signupReq, err := h.requestAdapter.ParseSignupRequest(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse signup request", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid request format"))
	}

	// Call application service
	signupResp, err := h.authService.Signup(echoCtx.Request().Context(), signupReq)
	if err != nil {
		h.logger.Warn("signup failed", "email", signupReq.Email, "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Failed to create account. Please try again."))
	}

	// Build response using adapter
	return h.responseAdapter.BuildSignupResponse(adapterCtx, signupResp)
}

// Logout handles POST /logout
func (h *AuthHandler) Logout(ctx httpiface.Context) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Parse logout request
	logoutReq, err := h.requestAdapter.ParseLogoutRequest(adapterCtx)
	if err != nil {
		h.logger.Error("failed to parse logout request", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Invalid request format"))
	}

	// Call application service
	logoutResp, err := h.authService.Logout(echoCtx.Request().Context(), logoutReq)
	if err != nil {
		h.logger.Error("logout failed", "error", err)

		return h.responseAdapter.BuildErrorResponse(adapterCtx, fmt.Errorf("Failed to logout"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildLogoutResponse(adapterCtx, logoutResp)
}

// TestEndpoint handles GET /api/v1/test
func (h *AuthHandler) TestEndpoint(ctx httpiface.Context) error {
	return fmt.Errorf("test endpoint working (placeholder)")
}
