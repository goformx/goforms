package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/validation"
	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock dependencies
type mockFormService struct {
	forms map[string]*model.Form
}

func (m *mockFormService) GetForm(ctx context.Context, formID string) (*model.Form, error) {
	if form, exists := m.forms[formID]; exists {
		return form, nil
	}
	return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "Form not found", nil)
}

func (m *mockFormService) CreateForm(ctx context.Context, form *model.Form) error {
	m.forms[form.ID] = form
	return nil
}

func (m *mockFormService) UpdateForm(ctx context.Context, form *model.Form) error {
	if _, exists := m.forms[form.ID]; !exists {
		return domainerrors.New(domainerrors.ErrCodeNotFound, "Form not found", nil)
	}
	m.forms[form.ID] = form
	return nil
}

func (m *mockFormService) DeleteForm(ctx context.Context, formID string) error {
	if _, exists := m.forms[formID]; !exists {
		return domainerrors.New(domainerrors.ErrCodeNotFound, "Form not found", nil)
	}
	delete(m.forms, formID)
	return nil
}

func (m *mockFormService) ListForms(ctx context.Context, userID string) ([]*model.Form, error) {
	var userForms []*model.Form
	for _, form := range m.forms {
		if form.UserID == userID {
			userForms = append(userForms, form)
		}
	}
	return userForms, nil
}

func (m *mockFormService) SubmitForm(ctx context.Context, submission *model.FormSubmission) error {
	return nil
}

func (m *mockFormService) GetFormSubmission(ctx context.Context, submissionID string) (*model.FormSubmission, error) {
	return nil, domainerrors.New(domainerrors.ErrCodeNotFound, "Submission not found", nil)
}

func (m *mockFormService) ListFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	return nil, nil
}

func (m *mockFormService) UpdateFormState(ctx context.Context, formID, state string) error {
	return nil
}

func (m *mockFormService) TrackFormAnalytics(ctx context.Context, formID, eventType string) error {
	return nil
}

func (m *mockFormService) LogFormAccess(form *model.Form) {
	// Mock implementation
}

func setupTestFormAPIHandler(t *testing.T) (*web.FormAPIHandler, *mockFormService, *echo.Echo) {
	// Create mock form service
	mockFormSvc := &mockFormService{
		forms: make(map[string]*model.Form),
	}

	// Create mock form with test data
	testForm := &model.Form{
		ID:          "test-form-123",
		UserID:      "test-user-123",
		Title:       "Test Form",
		Description: "Test Description",
		Schema: model.JSON{
			"type": "object",
			"components": []any{
				map[string]any{
					"key":  "name",
					"type": "textfield",
					"validate": map[string]any{
						"required": true,
					},
				},
			},
		},
		Active:    true,
		Status:    "published",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set up mock behavior
	mockFormSvc.forms[testForm.ID] = testForm

	// Create sanitizer
	sanitizer := sanitization.NewService()

	// Create logger factory
	loggerFactory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test",
		Version:     "1.0.0",
		Environment: "test",
	}, sanitizer)
	logger, err := loggerFactory.CreateLogger()
	require.NoError(t, err)

	// Create Echo instance
	e := echo.New()

	// Create error handler
	errorHandler := response.NewErrorHandler(logger, sanitizer)

	// Create base handler
	baseHandler := &web.BaseHandler{
		Logger:       logger,
		ErrorHandler: errorHandler,
	}

	// Create form base handler
	formBaseHandler := &web.FormBaseHandler{
		BaseHandler: baseHandler,
		FormService: mockFormSvc,
	}

	// Create handler with proper dependencies
	handler := &web.FormAPIHandler{
		FormBaseHandler:        formBaseHandler,
		ComprehensiveValidator: validation.NewComprehensiveValidator(),
		RequestProcessor:       web.NewFormRequestProcessor(sanitizer, validation.NewFormValidator(logger)),
		ResponseBuilder:        web.NewFormResponseBuilder(),
		ErrorHandler:           web.NewFormErrorHandler(web.NewFormResponseBuilder()),
		AccessManager:          access.NewAccessManager(access.DefaultConfig(), access.DefaultRules()),
	}

	return handler, mockFormSvc, e
}

func TestFormAPIHandler_RegisterRoutes(t *testing.T) {
	handler, _, e := setupTestFormAPIHandler(t)

	// Register routes
	handler.RegisterRoutes(e)

	// Test that routes are registered
	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	// Check for expected routes
	expectedRoutes := []string{
		"/api/v1/forms/:id/schema",
		"/api/v1/forms/:id/validation",
		"/api/v1/forms/:id/submit",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, routePaths[expectedRoute], "Expected route %s not found", expectedRoute)
	}
}

func TestFormAPIHandler_StartStop(t *testing.T) {
	handler, _, _ := setupTestFormAPIHandler(t)

	// Test start
	err := handler.Start(t.Context())
	require.NoError(t, err)

	// Test stop
	err = handler.Stop(t.Context())
	require.NoError(t, err)
}

func TestFormAPIHandler_GetFormByID(t *testing.T) {
	handler, _, e := setupTestFormAPIHandler(t)

	// Test valid form ID
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-form-123")

	form, err := handler.GetFormByID(c)
	require.NoError(t, err)
	assert.Equal(t, "test-form-123", form.ID)
	assert.Equal(t, "Test Form", form.Title)

	// Test non-existent form ID
	c.SetParamValues("non-existent-form")
	form, err = handler.GetFormByID(c)
	require.Error(t, err)
	assert.Nil(t, form)
	assert.Contains(t, err.Error(), "Form not found")

	// Test empty form ID
	c.SetParamValues("")
	form, err = handler.GetFormByID(c)
	require.Error(t, err)
	assert.Nil(t, form)
}

func TestFormAPIHandler_RequireFormOwnership(t *testing.T) {
	handler, _, e := setupTestFormAPIHandler(t)

	// Create a test form with ownership
	testForm := &model.Form{
		ID:     "owned-form-123",
		UserID: "test-user-123",
	}

	// Test with correct user ID
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "test-user-123")

	err := handler.RequireFormOwnership(c, testForm)
	require.NoError(t, err)

	// Test with incorrect user ID
	c.Set("user_id", "different-user")
	err = handler.RequireFormOwnership(c, testForm)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "You don't have permission")

	// Test with no user ID
	c.Set("user_id", nil)
	err = handler.RequireFormOwnership(c, testForm)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "User not authenticated")
}

func TestFormAPIHandler_GetFormWithOwnership(t *testing.T) {
	handler, mockFormSvc, e := setupTestFormAPIHandler(t)

	// Create a form with ownership
	testForm := &model.Form{
		ID:     "owned-form-123",
		UserID: "test-user-123",
	}
	mockFormSvc.forms[testForm.ID] = testForm

	// Test with correct user ID
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("owned-form-123")
	c.Set("user_id", "test-user-123")

	form, err := handler.GetFormWithOwnership(c)
	require.NoError(t, err)
	assert.Equal(t, "owned-form-123", form.ID)

	// Test with incorrect user ID
	c.Set("user_id", "different-user")
	form, err = handler.GetFormWithOwnership(c)
	require.Error(t, err)
	assert.Nil(t, form)
	assert.Contains(t, err.Error(), "You don't have permission")
}

func TestFormAPIHandler_RegisterAuthenticatedRoutes(t *testing.T) {
	handler, _, e := setupTestFormAPIHandler(t)

	// Create API group
	api := e.Group(constants.PathAPIv1)
	formsAPI := api.Group(constants.PathForms)

	// Register authenticated routes
	handler.RegisterAuthenticatedRoutes(formsAPI)

	// Test that authenticated routes are registered
	routes := e.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	// Check for expected authenticated routes
	expectedRoutes := []string{
		"/api/v1/forms/:id/schema",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, routePaths[expectedRoute], "Expected authenticated route %s not found", expectedRoute)
	}
}

func TestFormAPIHandler_ValidationEndpoint(t *testing.T) {
	handler, _, e := setupTestFormAPIHandler(t)

	// Register routes
	handler.RegisterRoutes(e)

	// Test validation endpoint with valid form ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/forms/test-form-123/validation", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 200 OK for valid form
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test validation endpoint with non-existent form ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/forms/non-existent-form/validation", http.NoBody)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 404 for non-existent form
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestFormAPIHandler_SchemaEndpoint(t *testing.T) {
	handler, _, e := setupTestFormAPIHandler(t)

	// Register routes
	handler.RegisterRoutes(e)

	// Test schema endpoint with valid form ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/forms/test-form-123/schema", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 200 OK for valid form
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test schema endpoint with non-existent form ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/forms/non-existent-form/schema", http.NoBody)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 404 for non-existent form
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestFormAPIHandler_SubmitEndpoint(t *testing.T) {
	handler, _, e := setupTestFormAPIHandler(t)

	// Register routes
	handler.RegisterRoutes(e)

	// Test submit endpoint with valid form ID and valid JSON
	validJSON := `{"name":"John Doe","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forms/test-form-123/submit", strings.NewReader(validJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 200 OK for valid submission
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test submit endpoint with invalid JSON
	invalidJSON := `{"name":"John Doe",}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/forms/test-form-123/submit", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return 400 for invalid JSON
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
