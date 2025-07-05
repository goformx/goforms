package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"context"

	"github.com/goformx/goforms/internal/application/dto"

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

// generateSimpleCSRFToken generates a simple CSRF token
func generateSimpleCSRFToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate CSRF token: " + err.Error())
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

// Login handles GET /login
func (h *AuthHandler) Login(ctx httpiface.Context) error {
	// Generate CSRF token directly
	csrfToken := generateSimpleCSRFToken()
	ctx.Set("csrf", csrfToken)

	// Create page data with proper asset path function
	pageData := &view.PageData{
		Title:         "Login",
		Description:   "Login to your account",
		Version:       h.config.App.Version,
		Environment:   h.config.App.Environment,
		AssetPath:     h.assetManager.AssetPath,
		User:          nil, // Will be set by middleware if authenticated
		Forms:         make([]*model.Form, 0),
		Submissions:   make([]*model.FormSubmission, 0),
		CSRFToken:     csrfToken, // Set the generated token
		IsDevelopment: h.config.App.IsDevelopment(),
		Config:        h.config,
	}

	// Render the login page using the framework-agnostic interface
	loginComponent := pages.Login(*pageData)

	if err := ctx.RenderComponent(loginComponent); err != nil {
		return fmt.Errorf("failed to render login component: %w", err)
	}

	return nil
}

// getInfraContext is a simple bridge to convert presentation Context to infrastructure Context
func (h *AuthHandler) getInfraContext(ctx httpiface.Context) (http.Context, error) {
	// Simple type assertion to the infrastructure adapter
	if infraCtx, ok := ctx.(*http.EchoContextAdapter); ok {
		return infraCtx, nil
	}

	return nil, fmt.Errorf("invalid context type")
}

// handleAuthPost is a generic handler for login/signup POST requests
func handleAuthPost[
	Req any, Resp any](
	h *AuthHandler,
	ctx httpiface.Context,
	parseReq func(http.Context) (Req, error),
	serviceFunc func(context.Context, Req) (Resp, error),
	buildResp func(http.Context, Resp) error,
	invalidInputMsg, serviceFailMsg string,
) error {
	infraCtx, err := h.getInfraContext(ctx)
	if err != nil {
		h.logger.Error("failed to get infrastructure context", "error", err)

		return fmt.Errorf("internal server error: context conversion failed")
	}

	req, err := parseReq(infraCtx)
	if err != nil {
		h.logger.Error("failed to parse request", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("%s", invalidInputMsg))
	}

	resp, err := serviceFunc(ctx.RequestContext(), req)
	if err != nil {
		h.logger.Warn("service failed", "error", err)

		return h.responseAdapter.BuildErrorResponse(infraCtx, fmt.Errorf("%s", serviceFailMsg))
	}

	if buildErr := buildResp(infraCtx, resp); buildErr != nil {
		return fmt.Errorf("failed to build response: %w", buildErr)
	}

	return nil
}

// LoginPost handles POST /login
func (h *AuthHandler) LoginPost(ctx httpiface.Context) error {
	return handleAuthPost[
		*dto.LoginRequest, *dto.LoginResponse](
		h,
		ctx,
		h.requestAdapter.ParseLoginRequest,
		h.authService.Login,
		h.responseAdapter.BuildLoginResponse,
		"invalid request format",
		"invalid email or password",
	)
}

// Signup handles GET /signup
func (h *AuthHandler) Signup(ctx httpiface.Context) error {
	// Generate CSRF token directly
	csrfToken := generateSimpleCSRFToken()
	ctx.Set("csrf", csrfToken)

	// Create page data with proper asset path function
	pageData := &view.PageData{
		Title:         "Sign Up",
		Description:   "Create a new account",
		Version:       h.config.App.Version,
		Environment:   h.config.App.Environment,
		AssetPath:     h.assetManager.AssetPath,
		User:          nil, // Will be set by middleware if authenticated
		Forms:         make([]*model.Form, 0),
		Submissions:   make([]*model.FormSubmission, 0),
		CSRFToken:     csrfToken, // Set the generated token
		IsDevelopment: h.config.App.IsDevelopment(),
		Config:        h.config,
	}

	// Render the signup page using the framework-agnostic interface
	signupComponent := pages.Signup(*pageData)

	if err := ctx.RenderComponent(signupComponent); err != nil {
		return fmt.Errorf("failed to render signup component: %w", err)
	}

	return nil
}

// SignupPost handles POST /signup
func (h *AuthHandler) SignupPost(ctx httpiface.Context) error {
	return handleAuthPost[
		*dto.SignupRequest, *dto.SignupResponse](
		h,
		ctx,
		h.requestAdapter.ParseSignupRequest,
		h.authService.Signup,
		h.responseAdapter.BuildSignupResponse,
		"invalid request format",
		"failed to create account, please try again",
	)
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
	if logoutBuildErr := h.responseAdapter.BuildLogoutResponse(infraCtx, logoutResp); logoutBuildErr != nil {
		return fmt.Errorf("failed to build logout response: %w", logoutBuildErr)
	}

	return nil
}

// TestEndpoint handles GET /api/v1/test
func (h *AuthHandler) TestEndpoint(ctx httpiface.Context) error {
	return fmt.Errorf("test endpoint working (placeholder)")
}
