package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/dto"
	"github.com/labstack/echo/v4"
)

// EchoContextAdapter adapts Echo context to our framework-agnostic Context interface
type EchoContextAdapter struct {
	echo.Context
}

// NewEchoContextAdapter creates a new Echo context adapter
func NewEchoContextAdapter(ctx echo.Context) *EchoContextAdapter {
	return &EchoContextAdapter{Context: ctx}
}

// Method returns the HTTP method
func (e *EchoContextAdapter) Method() string {
	return e.Request().Method
}

// Path returns the request path
func (e *EchoContextAdapter) Path() string {
	return e.Request().URL.Path
}

// Param returns a path parameter by name
func (e *EchoContextAdapter) Param(name string) string {
	return e.Context.Param(name)
}

// QueryParam returns a query parameter by name
func (e *EchoContextAdapter) QueryParam(name string) string {
	return e.Context.QueryParam(name)
}

// FormValue returns a form value by name
func (e *EchoContextAdapter) FormValue(name string) string {
	return e.Context.FormValue(name)
}

// Headers returns request headers as a map
func (e *EchoContextAdapter) Headers() map[string]string {
	headers := make(map[string]string)

	for key, values := range e.Request().Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return headers
}

// Body returns the request body as bytes
func (e *EchoContextAdapter) Body() []byte {
	// This is a simplified implementation
	// In a real implementation, you'd need to handle body reading properly
	return nil
}

// JSON sends a JSON response
func (e *EchoContextAdapter) JSON(statusCode int, data interface{}) error {
	return e.Context.JSON(statusCode, data)
}

// Redirect sends a redirect response
func (e *EchoContextAdapter) Redirect(statusCode int, url string) error {
	return e.Context.Redirect(statusCode, url)
}

// NoContent sends a no content response
func (e *EchoContextAdapter) NoContent(statusCode int) error {
	return e.Context.NoContent(statusCode)
}

// Get retrieves a value from the context
func (e *EchoContextAdapter) Get(key string) interface{} {
	return e.Context.Get(key)
}

// Set stores a value in the context
func (e *EchoContextAdapter) Set(key string, value interface{}) {
	e.Context.Set(key, value)
}

// EchoRequestAdapter implements RequestAdapter for Echo
type EchoRequestAdapter struct{}

// NewEchoRequestAdapter creates a new Echo request adapter
func NewEchoRequestAdapter() *EchoRequestAdapter {
	return &EchoRequestAdapter{}
}

// ParseLoginRequest parses login request from Echo context
func (a *EchoRequestAdapter) ParseLoginRequest(ctx Context) (*dto.LoginRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	contentType := echoCtx.Request().Header.Get("Content-Type")

	var request dto.LoginRequest

	if strings.Contains(contentType, "application/json") {
		if err := echoCtx.Bind(&request); err != nil {
			return nil, fmt.Errorf("failed to bind login request: %w", err)
		}
	} else {
		request.Email = echoCtx.FormValue("email")
		request.Password = echoCtx.FormValue("password")
	}

	// Sanitize inputs
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))

	return &request, nil
}

// ParseSignupRequest parses signup request from Echo context
func (a *EchoRequestAdapter) ParseSignupRequest(ctx Context) (*dto.SignupRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	contentType := echoCtx.Request().Header.Get("Content-Type")

	var request dto.SignupRequest

	if strings.Contains(contentType, "application/json") {
		if err := echoCtx.Bind(&request); err != nil {
			return nil, fmt.Errorf("failed to bind signup request: %w", err)
		}
	} else {
		request.Email = echoCtx.FormValue("email")
		request.Password = echoCtx.FormValue("password")
		request.ConfirmPassword = echoCtx.FormValue("confirm_password")
	}

	// Sanitize inputs
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))

	return &request, nil
}

// ParseLogoutRequest parses logout request from Echo context
func (a *EchoRequestAdapter) ParseLogoutRequest(ctx Context) (*dto.LogoutRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	userID := echoCtx.Get("user_id")
	sessionID := echoCtx.Get("session_id")

	if userID == nil || sessionID == nil {
		return nil, fmt.Errorf("missing user_id or session_id")
	}

	return &dto.LogoutRequest{
		UserID:    userID.(string),
		SessionID: sessionID.(string),
	}, nil
}

// ParseCreateFormRequest parses create form request from Echo context
func (a *EchoRequestAdapter) ParseCreateFormRequest(ctx Context) (*dto.CreateFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	var request dto.CreateFormRequest
	if err := echoCtx.Bind(&request); err != nil {
		return nil, fmt.Errorf("failed to bind create form request: %w", err)
	}

	// Set user ID from context
	if userID := echoCtx.Get("user_id"); userID != nil {
		request.UserID = userID.(string)
	}

	return &request, nil
}

// ParseUpdateFormRequest parses update form request from Echo context
func (a *EchoRequestAdapter) ParseUpdateFormRequest(ctx Context) (*dto.UpdateFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	var request dto.UpdateFormRequest
	if err := echoCtx.Bind(&request); err != nil {
		return nil, fmt.Errorf("failed to bind update form request: %w", err)
	}

	// Set form ID from path parameter
	request.ID = echoCtx.Param("id")

	// Set user ID from context
	if userID := echoCtx.Get("user_id"); userID != nil {
		request.UserID = userID.(string)
	}

	return &request, nil
}

// ParseDeleteFormRequest parses delete form request from Echo context
func (a *EchoRequestAdapter) ParseDeleteFormRequest(ctx Context) (*dto.DeleteFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	request := &dto.DeleteFormRequest{
		ID: echoCtx.Param("id"),
	}

	// Set user ID from context
	if userID := echoCtx.Get("user_id"); userID != nil {
		request.UserID = userID.(string)
	}

	return request, nil
}

// ParseSubmitFormRequest parses submit form request from Echo context
func (a *EchoRequestAdapter) ParseSubmitFormRequest(ctx Context) (*dto.SubmitFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	var request dto.SubmitFormRequest
	if err := echoCtx.Bind(&request); err != nil {
		return nil, fmt.Errorf("failed to bind submit form request: %w", err)
	}

	// Set form ID from path parameter
	request.FormID = echoCtx.Param("id")

	// Set user ID from context if available
	if userID := echoCtx.Get("user_id"); userID != nil {
		request.UserID = userID.(string)
	}

	return &request, nil
}

// ParsePaginationRequest parses pagination request from Echo context
func (a *EchoRequestAdapter) ParsePaginationRequest(ctx Context) (*dto.PaginationRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	page, _ := strconv.Atoi(echoCtx.QueryParam("page"))
	limit, _ := strconv.Atoi(echoCtx.QueryParam("limit"))

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	return &dto.PaginationRequest{
		Page:  page,
		Limit: limit,
	}, nil
}

// ParseFormID parses form ID from Echo context
func (a *EchoRequestAdapter) ParseFormID(ctx Context) (string, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return "", fmt.Errorf("invalid context type")
	}

	formID := echoCtx.Param("id")
	if formID == "" {
		return "", fmt.Errorf("form ID is required")
	}

	return formID, nil
}

// ParseUserID parses user ID from Echo context
func (a *EchoRequestAdapter) ParseUserID(ctx Context) (string, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return "", fmt.Errorf("invalid context type")
	}

	userID := echoCtx.Get("user_id")
	if userID == nil {
		return "", fmt.Errorf("user ID not found in context")
	}

	return userID.(string), nil
}

// EchoResponseAdapter implements ResponseAdapter for Echo
type EchoResponseAdapter struct{}

// NewEchoResponseAdapter creates a new Echo response adapter
func NewEchoResponseAdapter() *EchoResponseAdapter {
	return &EchoResponseAdapter{}
}

// BuildLoginResponse builds login response for Echo
func (a *EchoResponseAdapter) BuildLoginResponse(ctx Context, response *dto.LoginResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusOK, dto.NewSuccessResponse("Login successful", response))
	}

	// For web requests, redirect to dashboard
	return echoCtx.Redirect(http.StatusSeeOther, constants.PathDashboard)
}

// BuildSignupResponse builds signup response for Echo
func (a *EchoResponseAdapter) BuildSignupResponse(ctx Context, response *dto.SignupResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusCreated, dto.NewSuccessResponse("Signup successful", response))
	}

	// For web requests, redirect to dashboard
	return echoCtx.Redirect(http.StatusSeeOther, constants.PathDashboard)
}

// BuildLogoutResponse builds logout response for Echo
func (a *EchoResponseAdapter) BuildLogoutResponse(ctx Context, response *dto.LogoutResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusOK, dto.NewSuccessResponse("Logout successful", response))
	}

	// For web requests, redirect to login
	return echoCtx.Redirect(http.StatusSeeOther, constants.PathLogin)
}

// BuildFormResponse builds form response for Echo
func (a *EchoResponseAdapter) BuildFormResponse(ctx Context, response *dto.FormResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, dto.NewSuccessResponse("Form retrieved successfully", response))
}

// BuildFormListResponse builds form list response for Echo
func (a *EchoResponseAdapter) BuildFormListResponse(ctx Context, response *dto.FormListResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, dto.NewSuccessResponse("Forms retrieved successfully", response))
}

// BuildFormSchemaResponse builds form schema response for Echo
func (a *EchoResponseAdapter) BuildFormSchemaResponse(ctx Context, response *dto.FormSchemaResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, dto.NewSuccessResponse("Form schema retrieved successfully", response))
}

// BuildSubmitFormResponse builds submit form response for Echo
func (a *EchoResponseAdapter) BuildSubmitFormResponse(ctx Context, response *dto.SubmitFormResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusCreated, dto.NewSuccessResponse("Form submitted successfully", response))
}

// BuildErrorResponse builds error response for Echo
func (a *EchoResponseAdapter) BuildErrorResponse(ctx Context, err error) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	errorResponse := dto.NewErrorResponse("ERROR", err.Error())

	return echoCtx.JSON(http.StatusInternalServerError, errorResponse)
}

// BuildValidationErrorResponse builds validation error response for Echo
func (a *EchoResponseAdapter) BuildValidationErrorResponse(ctx Context, errors []dto.ValidationError) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	errorResponse := dto.NewValidationErrorResponse(errors)

	return echoCtx.JSON(http.StatusBadRequest, errorResponse)
}

// BuildNotFoundResponse builds not found response for Echo
func (a *EchoResponseAdapter) BuildNotFoundResponse(ctx Context, resource string) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	errorResponse := dto.NewErrorResponse("NOT_FOUND", fmt.Sprintf("%s not found", resource))

	return echoCtx.JSON(http.StatusNotFound, errorResponse)
}

// BuildUnauthorizedResponse builds unauthorized response for Echo
func (a *EchoResponseAdapter) BuildUnauthorizedResponse(ctx Context) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	errorResponse := dto.NewErrorResponse("UNAUTHORIZED", "Authentication required")

	return echoCtx.JSON(http.StatusUnauthorized, errorResponse)
}

// BuildForbiddenResponse builds forbidden response for Echo
func (a *EchoResponseAdapter) BuildForbiddenResponse(ctx Context) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	errorResponse := dto.NewErrorResponse("FORBIDDEN", "Access denied")

	return echoCtx.JSON(http.StatusForbidden, errorResponse)
}

// BuildSuccessResponse builds success response for Echo
func (a *EchoResponseAdapter) BuildSuccessResponse(ctx Context, message string, data interface{}) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	response := dto.NewSuccessResponse(message, data)

	return echoCtx.JSON(http.StatusOK, response)
}

// BuildJSONResponse builds JSON response for Echo
func (a *EchoResponseAdapter) BuildJSONResponse(ctx Context, statusCode int, data interface{}) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(statusCode, data)
}

// isAPIRequest checks if the request is an API request
func (a *EchoResponseAdapter) isAPIRequest(ctx *EchoContextAdapter) bool {
	accept := ctx.Request().Header.Get("Accept")
	contentType := ctx.Request().Header.Get("Content-Type")

	return strings.Contains(accept, "application/json") ||
		strings.Contains(contentType, "application/json") ||
		strings.HasPrefix(ctx.Request().URL.Path, "/api/")
}
