package http

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/dto"
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
		if err := a.parseFormData(echoCtx, &request); err != nil {
			return nil, err
		}
	}

	// Get form ID from path
	request.FormID = echoCtx.Param("id")

	return &request, nil
}

// parseFormData parses form data and populates the request
func (a *EchoRequestAdapter) parseFormData(echoCtx *EchoContextAdapter, request *dto.SubmitFormRequest) error {
	// For form data, we need to parse the form values
	if err := echoCtx.Request().ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	// Convert form values to map[string]any
	data := make(map[string]any)

	for key, values := range echoCtx.Request().Form {
		if len(values) == 1 {
			data[key] = values[0]
		} else {
			data[key] = values
		}
	}

	request.Data = data

	return nil
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
