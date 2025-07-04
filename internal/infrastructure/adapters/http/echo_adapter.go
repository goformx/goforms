package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/dto"
	"github.com/labstack/echo/v4"
)

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
	if userID == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return nil, fmt.Errorf("user_id is not a string")
	}

	return &dto.LogoutRequest{
		UserID: userIDStr,
	}, nil
}

// ParseCreateFormRequest parses create form request from Echo context
func (a *EchoRequestAdapter) ParseCreateFormRequest(ctx Context) (*dto.CreateFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	contentType := echoCtx.Request().Header.Get("Content-Type")

	var request dto.CreateFormRequest

	if strings.Contains(contentType, "application/json") {
		if err := echoCtx.Bind(&request); err != nil {
			return nil, fmt.Errorf("failed to bind create form request: %w", err)
		}
	} else {
		request.Title = echoCtx.FormValue("title")
		request.Description = echoCtx.FormValue("description")
		// For form data, schema would need to be parsed from JSON string
		// For now, we'll require JSON for schema
		return nil, fmt.Errorf("schema must be provided as JSON")
	}

	// Get user ID from context
	userID := echoCtx.Get("user_id")
	if userID == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return nil, fmt.Errorf("user_id is not a string")
	}

	request.UserID = userIDStr

	return &request, nil
}

// ParseUpdateFormRequest parses update form request from Echo context
func (a *EchoRequestAdapter) ParseUpdateFormRequest(ctx Context) (*dto.UpdateFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	contentType := echoCtx.Request().Header.Get("Content-Type")

	var request dto.UpdateFormRequest

	if strings.Contains(contentType, "application/json") {
		if err := echoCtx.Bind(&request); err != nil {
			return nil, fmt.Errorf("failed to bind update form request: %w", err)
		}
	} else {
		request.Title = echoCtx.FormValue("title")
		request.Description = echoCtx.FormValue("description")
		// For form data, schema would need to be parsed from JSON string
		// For now, we'll require JSON for schema
		return nil, fmt.Errorf("schema must be provided as JSON")
	}

	// Get form ID from path
	request.ID = echoCtx.Param("id")

	// Get user ID from context
	userID := echoCtx.Get("user_id")
	if userID == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return nil, fmt.Errorf("user_id is not a string")
	}

	request.UserID = userIDStr

	return &request, nil
}

// ParseDeleteFormRequest parses delete form request from Echo context
func (a *EchoRequestAdapter) ParseDeleteFormRequest(ctx Context) (*dto.DeleteFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	// Get form ID from path
	formID := echoCtx.Param("id")

	// Get user ID from context
	userID := echoCtx.Get("user_id")
	if userID == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return nil, fmt.Errorf("user_id is not a string")
	}

	return &dto.DeleteFormRequest{
		ID:     formID,
		UserID: userIDStr,
	}, nil
}

// ParseSubmitFormRequest parses submit form request from Echo context
func (a *EchoRequestAdapter) ParseSubmitFormRequest(ctx Context) (*dto.SubmitFormRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	contentType := echoCtx.Request().Header.Get("Content-Type")

	var request dto.SubmitFormRequest

	if strings.Contains(contentType, "application/json") {
		if err := echoCtx.Bind(&request); err != nil {
			return nil, fmt.Errorf("failed to bind submit form request: %w", err)
		}
	} else {
		// For form data, we need to parse the form values
		if err := echoCtx.Request().ParseForm(); err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}

		// Convert form values to map[string]interface{}
		data := make(map[string]interface{})

		for key, values := range echoCtx.Request().Form {
			if len(values) == 1 {
				data[key] = values[0]
			} else {
				data[key] = values
			}
		}

		request.Data = data
	}

	// Get form ID from path
	request.FormID = echoCtx.Param("id")

	return &request, nil
}

// ParsePaginationRequest parses pagination request from Echo context
func (a *EchoRequestAdapter) ParsePaginationRequest(ctx Context) (*dto.PaginationRequest, error) {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	pageStr := echoCtx.QueryParam("page")
	limitStr := echoCtx.QueryParam("limit")

	page := 1
	limit := constants.DefaultPageSize

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= constants.MaxPageSize {
			limit = l
		}
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
		return "", fmt.Errorf("form ID not found in path")
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
		return "", fmt.Errorf("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("user_id is not a string")
	}

	return userIDStr, nil
}

// EchoResponseAdapter implements ResponseAdapter for Echo
type EchoResponseAdapter struct{}

// NewEchoResponseAdapter creates a new Echo response adapter
func NewEchoResponseAdapter() *EchoResponseAdapter {
	return &EchoResponseAdapter{}
}

// BuildLoginResponse builds login response for Echo context
func (a *EchoResponseAdapter) BuildLoginResponse(ctx Context, response *dto.LoginResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusOK, response)
	}

	// For web requests, redirect to dashboard
	return echoCtx.Redirect(http.StatusFound, "/dashboard")
}

// BuildSignupResponse builds signup response for Echo context
func (a *EchoResponseAdapter) BuildSignupResponse(ctx Context, response *dto.SignupResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusCreated, response)
	}

	// For web requests, redirect to login
	return echoCtx.Redirect(http.StatusFound, "/login")
}

// BuildLogoutResponse builds logout response for Echo context
func (a *EchoResponseAdapter) BuildLogoutResponse(ctx Context, response *dto.LogoutResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	if a.isAPIRequest(echoCtx) {
		return echoCtx.JSON(http.StatusOK, response)
	}

	// For web requests, redirect to login
	return echoCtx.Redirect(http.StatusFound, "/login")
}

// BuildFormResponse builds form response for Echo context
func (a *EchoResponseAdapter) BuildFormResponse(ctx Context, response *dto.FormResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, response)
}

// BuildFormListResponse builds form list response for Echo context
func (a *EchoResponseAdapter) BuildFormListResponse(ctx Context, response *dto.FormListResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, response)
}

// BuildFormSchemaResponse builds form schema response for Echo context
func (a *EchoResponseAdapter) BuildFormSchemaResponse(ctx Context, response *dto.FormSchemaResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, response)
}

// BuildSubmitFormResponse builds submit form response for Echo context
func (a *EchoResponseAdapter) BuildSubmitFormResponse(ctx Context, response *dto.SubmitFormResponse) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusCreated, response)
}

// BuildErrorResponse builds error response for Echo context
func (a *EchoResponseAdapter) BuildErrorResponse(ctx Context, err error) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error": err.Error(),
	})
}

// BuildValidationErrorResponse builds validation error response for Echo context
func (a *EchoResponseAdapter) BuildValidationErrorResponse(ctx Context, errors []dto.ValidationError) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusBadRequest, map[string]interface{}{
		"errors": errors,
	})
}

// BuildNotFoundResponse builds not found response for Echo context
func (a *EchoResponseAdapter) BuildNotFoundResponse(ctx Context, resource string) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusNotFound, map[string]interface{}{
		"error": fmt.Sprintf("%s not found", resource),
	})
}

// BuildUnauthorizedResponse builds unauthorized response for Echo context
func (a *EchoResponseAdapter) BuildUnauthorizedResponse(ctx Context) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusUnauthorized, map[string]interface{}{
		"error": "unauthorized",
	})
}

// BuildForbiddenResponse builds forbidden response for Echo context
func (a *EchoResponseAdapter) BuildForbiddenResponse(ctx Context) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusForbidden, map[string]interface{}{
		"error": "forbidden",
	})
}

// BuildSuccessResponse builds success response for Echo context
func (a *EchoResponseAdapter) BuildSuccessResponse(ctx Context, message string, data interface{}) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": message,
		"data":    data,
	})
}

// BuildJSONResponse builds generic JSON response for Echo context
func (a *EchoResponseAdapter) BuildJSONResponse(ctx Context, statusCode int, data interface{}) error {
	echoCtx, ok := ctx.(*EchoContextAdapter)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	return echoCtx.JSON(statusCode, data)
}

// isAPIRequest checks if the request is an API request
func (a *EchoResponseAdapter) isAPIRequest(ctx echo.Context) bool {
	accept := ctx.Request().Header.Get("Accept")

	return strings.Contains(accept, "application/json")
}

// EchoContextAdapter implements Context for Echo
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

// Headers returns all request headers
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
	// This would need to be implemented based on how you want to handle the body
	// For now, return empty slice
	return []byte{}
}

// JSON sends a JSON response
func (e *EchoContextAdapter) JSON(statusCode int, data interface{}) error {
	return e.Context.JSON(statusCode, data)
}

// JSONBlob sends a JSON blob response
func (e *EchoContextAdapter) JSONBlob(statusCode int, data []byte) error {
	return e.Context.JSONBlob(statusCode, data)
}

// String sends a string response
func (e *EchoContextAdapter) String(statusCode int, data string) error {
	return e.Context.String(statusCode, data)
}

// Redirect redirects the request
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

// RequestContext returns the underlying context.Context
func (e *EchoContextAdapter) RequestContext() context.Context {
	return e.Request().Context()
}

// GetUnderlyingContext returns the underlying Echo context for bridge methods
func (e *EchoContextAdapter) GetUnderlyingContext() interface{} {
	return e.Context
}

// RenderComponent renders a component
func (e *EchoContextAdapter) RenderComponent(component interface{}) error {
	// This would delegate to the renderer service
	// For now, we'll use Echo's built-in rendering
	return e.Render(http.StatusOK, "component", component)
}

// EchoAdapter registers handlers with an echo.Echo instance.
type EchoAdapter struct {
	e *echo.Echo
}

// NewEchoAdapter creates a new EchoAdapter for the given echo.Echo instance.
func NewEchoAdapter(e *echo.Echo) *EchoAdapter {
	return &EchoAdapter{e: e}
}

// RegisterHandler registers all routes from the given handler with Echo.
func (a *EchoAdapter) RegisterHandler(handler interface{}) error {
	// This is a simple adapter that just passes through to Echo
	// The actual handler registration logic should be in the presentation layer
	// This is just a placeholder to satisfy the dependency injection
	return nil
}
