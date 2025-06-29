package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/application/handlers/web"
	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form/model"
)

func setupTestFormErrorHandler() (web.FormErrorHandler, *echo.Echo) {
	// Create response builder
	responseBuilder := web.NewFormResponseBuilder()

	// Create error handler
	errorHandler := web.NewFormErrorHandler(responseBuilder)

	// Create Echo instance for testing
	e := echo.New()

	return errorHandler, e
}

// runErrorHandlerTest is a helper function to reduce duplication in error handler tests
func runErrorHandlerTest(t *testing.T, e *echo.Echo, tests []struct {
	name           string
	err            error
	expectedStatus int
	expectedBody   string
	description    string
}, handlerFunc func(echo.Context, error) error) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handle error
			err := handlerFunc(c, tt.err)
			require.NoError(t, err)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

func TestFormErrorHandler_HandleSchemaError(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "schema required error",
			err:            model.ErrFormSchemaRequired,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Form schema is required",
			description:    "Should return 400 for missing schema",
		},
		{
			name:           "invalid schema error",
			err:            model.ErrFormInvalid,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid form schema format",
			description:    "Should return 400 for invalid schema",
		},
		{
			name:           "unknown schema error",
			err:            domainerrors.New(domainerrors.ErrCodeServerError, "unknown schema error", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to process form schema",
			description:    "Should return 500 for unknown schema errors",
		},
	}

	runErrorHandlerTest(t, e, tests, handler.HandleSchemaError)
}

func TestFormErrorHandler_HandleSubmissionError(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "form not found error",
			err:            model.ErrFormNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Form not found",
			description:    "Should return 404 for missing form",
		},
		{
			name:           "invalid submission error",
			err:            model.ErrFormInvalid,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid submission data",
			description:    "Should return 400 for invalid submission",
		},
		{
			name:           "submission not found error",
			err:            model.ErrSubmissionNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Submission not found",
			description:    "Should return 404 for missing submission",
		},
		{
			name:           "unknown submission error",
			err:            domainerrors.New(domainerrors.ErrCodeServerError, "unknown submission error", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to process form submission",
			description:    "Should return 500 for unknown submission errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handle error
			err := handler.HandleSubmissionError(c, tt.err)
			require.NoError(t, err)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

func TestFormErrorHandler_HandleError(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "title required error",
			err:            model.ErrFormTitleRequired,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Form title is required",
			description:    "Should return 400 for missing title",
		},
		{
			name:           "form invalid error",
			err:            model.ErrFormInvalid,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Form validation failed",
			description:    "Should return 400 for invalid form",
		},
		{
			name:           "unknown validation error",
			err:            domainerrors.New(domainerrors.ErrCodeValidation, "unknown validation error", nil),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
			description:    "Should return 400 for unknown validation errors",
		},
	}

	runErrorHandlerTest(t, e, tests, handler.HandleError)
}

func TestFormErrorHandler_HandleOwnershipError(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "forbidden error",
			err:            domainerrors.New(domainerrors.ErrCodeForbidden, "Forbidden", nil),
			expectedStatus: http.StatusForbidden,
			expectedBody:   "You don't have permission to access this resource",
			description:    "Should return 403 for forbidden errors",
		},
		{
			name:           "authentication error",
			err:            domainerrors.New(domainerrors.ErrCodeUnauthorized, "Unauthorized", nil),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication required",
			description:    "Should return 401 for authentication errors",
		},
		{
			name:           "unknown ownership error",
			err:            domainerrors.New(domainerrors.ErrCodeForbidden, "unknown ownership error", nil),
			expectedStatus: http.StatusForbidden,
			expectedBody:   "You don't have permission to access this resource",
			description:    "Should return 403 for unknown ownership errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handle error
			err := handler.HandleOwnershipError(c, tt.err)
			require.NoError(t, err)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

func TestFormErrorHandler_HandleFormNotFoundError(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	tests := []struct {
		name           string
		formID         string
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "form not found with ID",
			formID:         "test-form-123",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Form not found",
			description:    "Should return 404 for missing form",
		},
		{
			name:           "form not found without ID",
			formID:         "",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Form not found",
			description:    "Should return 404 for missing form without ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handle error
			err := handler.HandleFormNotFoundError(c, tt.formID)
			require.NoError(t, err)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

func TestFormErrorHandler_HandleFormAccessError(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "not found error",
			err:            domainerrors.New(domainerrors.ErrCodeNotFound, "Not found", nil),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Form not found",
			description:    "Should return 404 for not found errors",
		},
		{
			name:           "forbidden error",
			err:            domainerrors.New(domainerrors.ErrCodeForbidden, "Forbidden", nil),
			expectedStatus: http.StatusForbidden,
			expectedBody:   "You don't have permission to access this resource",
			description:    "Should return 403 for forbidden errors",
		},
		{
			name:           "authentication error",
			err:            domainerrors.New(domainerrors.ErrCodeUnauthorized, "Unauthorized", nil),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication required",
			description:    "Should return 401 for authentication errors",
		},
		{
			name:           "unknown access error",
			err:            domainerrors.New(domainerrors.ErrCodeServerError, "unknown access error", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to access form",
			description:    "Should return 500 for unknown access errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handle error
			err := handler.HandleFormAccessError(c, tt.err)
			require.NoError(t, err)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

func TestFormErrorHandler_ErrorResponseFormat(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handle error
	err := handler.HandleFormNotFoundError(c, "test-form-123")
	require.NoError(t, err)

	// Check response format
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Form not found")
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

func TestFormErrorHandler_ConsistentErrorHandling(t *testing.T) {
	handler, e := setupTestFormErrorHandler()

	// Test that all error handlers return consistent response format
	errorHandlers := []struct {
		name string
		fn   func(echo.Context, error) error
		err  error
	}{
		{
			name: "HandleSchemaError",
			fn:   handler.HandleSchemaError,
			err:  model.ErrFormSchemaRequired,
		},
		{
			name: "HandleSubmissionError",
			fn:   handler.HandleSubmissionError,
			err:  model.ErrFormNotFound,
		},
		{
			name: "HandleError",
			fn:   handler.HandleError,
			err:  model.ErrFormTitleRequired,
		},
		{
			name: "HandleOwnershipError",
			fn:   handler.HandleOwnershipError,
			err:  domainerrors.New(domainerrors.ErrCodeForbidden, "Forbidden", nil),
		},
		{
			name: "HandleFormAccessError",
			fn:   handler.HandleFormAccessError,
			err:  domainerrors.New(domainerrors.ErrCodeNotFound, "Not found", nil),
		},
	}

	for _, eh := range errorHandlers {
		t.Run(eh.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Handle error
			err := eh.fn(c, eh.err)
			require.NoError(t, err)

			// Check consistent format
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			assert.NotEmpty(t, rec.Body.String())
		})
	}
}
