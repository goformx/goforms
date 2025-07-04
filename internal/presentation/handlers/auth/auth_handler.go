package auth

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/view"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
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
	// Create page data for the login template using the view package
	pageData := &view.PageData{
		Title:       "Login",
		Description: "Login to your account",
		Version:     h.config.App.Version,
		Environment: h.config.App.Environment,
		User:        nil, // Will be set by middleware if authenticated
		Forms:       make([]*model.Form, 0),
		Submissions: make([]*model.FormSubmission, 0),
		Config:      h.config,
	}

	// Render the login page using the framework-agnostic interface
	loginComponent := pages.Login(*pageData)

	return ctx.RenderComponent(loginComponent)
}

// handleAuthRequest is a helper method to reduce code duplication in auth handlers
func (h *AuthHandler) handleAuthRequest(ctx httpiface.Context, operation string, handler func(http.Context, echo.Context) error) error {
	// Extract the underlying Echo context
	echoCtx, ok := ctx.GetUnderlyingContext().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")

		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Wrap echo context with our adapter
	adapterCtx := http.NewEchoContextAdapter(echoCtx)

	// Call the specific handler
	return handler(adapterCtx, echoCtx)
}

// getInfraContext is a simple bridge to convert presentation Context to infrastructure Context
func (h *AuthHandler) getInfraContext(ctx httpiface.Context) (http.Context, error) {
	// Simple type assertion to the infrastructure adapter
	if infraCtx, ok := ctx.(*http.EchoContextAdapter); ok {
		return infraCtx, nil
	}
	return nil, fmt.Errorf("invalid context type")
}

// LoginPost handles POST /login
func (h *AuthHandler) LoginPost(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)
		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Parse request using adapter
	loginReq, err := h.requestAdapter.ParseLoginRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse login request", "error", err)
		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid request format"))
	}

	// Call application service
	loginResp, err := h.authService.Login(ctx.RequestContext(), loginReq)
	if err != nil {
		h.logger.Warn("login failed", "email", loginReq.Email, "error", err)
		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid email or password"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildLoginResponse(infraCtx, loginResp)
}

// Signup handles GET /signup
func (h *AuthHandler) Signup(ctx httpiface.Context) error {
	// Create page data for the signup template using the view package
	pageData := &view.PageData{
		Title:       "Sign Up",
		Description: "Create a new account",
		Version:     h.config.App.Version,
		Environment: h.config.App.Environment,
		User:        nil, // Will be set by middleware if authenticated
		Forms:       make([]*model.Form, 0),
		Submissions: make([]*model.FormSubmission, 0),
		Config:      h.config,
	}

	// Render the signup page using the framework-agnostic interface
	signupComponent := pages.Signup(*pageData)

	return ctx.RenderComponent(signupComponent)
}

// SignupPost handles POST /signup
func (h *AuthHandler) SignupPost(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)
		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Parse request using adapter
	signupReq, err := h.requestAdapter.ParseSignupRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse signup request", "error", err)
		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid request format"))
	}

	// Call application service
	signupResp, err := h.authService.Signup(ctx.RequestContext(), signupReq)
	if err != nil {
		h.logger.Warn("signup failed", "email", signupReq.Email, "error", err)
		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to create account, please try again"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildSignupResponse(infraCtx, signupResp)
}

// Logout handles POST /logout
func (h *AuthHandler) Logout(ctx httpiface.Context) error {
	// Get infrastructure context using bridge
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)
		return fmt.Errorf("internal server error: context conversion failed")
	}

	// Parse logout request
	logoutReq, err := h.requestAdapter.ParseLogoutRequest(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse logout request", "error", err)
		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("invalid request format"))
	}

	// Call application service
	logoutResp, err := h.authService.Logout(ctx.RequestContext(), logoutReq)
	if err != nil {
		h.logger.Error("logout failed", "error", err)
		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("failed to logout"))
	}

	// Build response using adapter
	return h.responseAdapter.BuildLogoutResponse(infraCtx, logoutResp)
}

// TestEndpoint handles GET /api/v1/test
func (h *AuthHandler) TestEndpoint(ctx httpiface.Context) error {
	return fmt.Errorf("test endpoint working (placeholder)")
}
