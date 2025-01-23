package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/core/contact"
	contactmock "github.com/jonesrussell/goforms/test/mocks/contact"
	"github.com/jonesrussell/goforms/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateContact(t *testing.T) {
	tests := []struct {
		name           string
		submission     contact.Submission
		setupFn        func(*contactmock.MockService)
		expectedStatus int
	}{
		{
			name: "valid submission",
			submission: contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupFn: func(ms *contactmock.MockService) {
				ms.On("CreateSubmission", mock.Anything, mock.MatchedBy(func(s *contact.Submission) bool {
					return s.Name == "Test User" && s.Email == "test@example.com"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "service error",
			submission: contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupFn: func(ms *contactmock.MockService) {
				ms.On("CreateSubmission", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid json",
			submission: contact.Submission{},
			setupFn: func(ms *contactmock.MockService) {
				// No mock setup needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setup := utils.NewTestSetup()
			defer setup.Close()

			mockService := contactmock.NewMockService()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, setup.Logger)

			// Create request
			var body interface{}
			if tt.name == "invalid json" {
				body = "{invalid json"
			} else {
				body = tt.submission
			}
			req, err := utils.NewJSONRequest(http.MethodPost, "/api/v1/contacts", body)
			assert.NoError(t, err)

			// Execute request
			c, rec := utils.NewTestContext(setup.Echo, req)
			err = api.CreateContact(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusCreated {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestListContacts(t *testing.T) {
	tests := []struct {
		name           string
		setupFn        func(*contactmock.MockService)
		expectedStatus int
	}{
		{
			name: "successful list",
			setupFn: func(ms *contactmock.MockService) {
				ms.On("ListSubmissions", mock.Anything).Return([]contact.Submission{
					{ID: 1, Name: "Test User", Email: "test@example.com"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupFn: func(ms *contactmock.MockService) {
				ms.On("ListSubmissions", mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setup := utils.NewTestSetup()
			defer setup.Close()

			mockService := contactmock.NewMockService()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, setup.Logger)

			// Create request
			req, err := utils.NewJSONRequest(http.MethodGet, "/api/v1/contacts", nil)
			assert.NoError(t, err)

			// Execute request
			c, rec := utils.NewTestContext(setup.Echo, req)
			err = api.ListContacts(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetContact(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupFn        func(*contactmock.MockService)
		expectedStatus int
	}{
		{
			name: "existing submission",
			id:   "1",
			setupFn: func(ms *contactmock.MockService) {
				ms.On("GetSubmission", mock.Anything, int64(1)).Return(&contact.Submission{
					ID:    1,
					Name:  "Test User",
					Email: "test@example.com",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "non-existent submission",
			id:   "999",
			setupFn: func(ms *contactmock.MockService) {
				ms.On("GetSubmission", mock.Anything, int64(999)).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid id",
			id:   "invalid",
			setupFn: func(ms *contactmock.MockService) {
				// No mock setup needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setup := utils.NewTestSetup()
			defer setup.Close()

			mockService := contactmock.NewMockService()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, setup.Logger)

			// Create request
			req, err := utils.NewJSONRequest(http.MethodGet, "/", nil)
			assert.NoError(t, err)

			// Execute request
			c, rec := utils.NewTestContext(setup.Echo, req)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err = api.GetContact(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateContactStatus(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		status         contact.Status
		setupFn        func(*contactmock.MockService)
		expectedStatus int
	}{
		{
			name:   "valid update",
			id:     "1",
			status: contact.StatusApproved,
			setupFn: func(ms *contactmock.MockService) {
				ms.On("UpdateSubmissionStatus", mock.Anything, int64(1), contact.StatusApproved).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "service error",
			id:     "1",
			status: contact.StatusApproved,
			setupFn: func(ms *contactmock.MockService) {
				ms.On("UpdateSubmissionStatus", mock.Anything, int64(1), contact.StatusApproved).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "invalid id",
			id:     "invalid",
			status: contact.StatusApproved,
			setupFn: func(ms *contactmock.MockService) {
				// No mock setup needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid json",
			id:     "1",
			status: contact.StatusApproved,
			setupFn: func(ms *contactmock.MockService) {
				// No mock setup needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setup := utils.NewTestSetup()
			defer setup.Close()

			mockService := contactmock.NewMockService()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, setup.Logger)

			// Create request
			var body interface{}
			if tt.name == "invalid json" {
				body = "{invalid json"
			} else {
				body = map[string]string{"status": string(tt.status)}
			}
			req, err := utils.NewJSONRequest(http.MethodPut, "/", body)
			assert.NoError(t, err)

			// Execute request
			c, rec := utils.NewTestContext(setup.Echo, req)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err = api.UpdateContactStatus(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestContactRegister(t *testing.T) {
	// Setup
	setup := utils.NewTestSetup()
	defer setup.Close()

	mockService := contactmock.NewMockService()

	// Set up mock expectations for any potential service calls
	mockService.On("CreateSubmission", mock.Anything, mock.Anything).Return(nil)
	mockService.On("ListSubmissions", mock.Anything).Return([]contact.Submission{}, nil)
	mockService.On("GetSubmission", mock.Anything, mock.Anything).Return(&contact.Submission{}, nil)
	mockService.On("UpdateSubmissionStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	api := NewContactAPI(mockService, setup.Logger)

	// Test registration
	api.Register(setup.Echo)

	// Verify routes are registered by making test requests
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/contacts"},
		{http.MethodGet, "/api/v1/contacts"},
		{http.MethodGet, "/api/v1/contacts/1"},
		{http.MethodPut, "/api/v1/contacts/1/status"},
	}

	for _, route := range routes {
		req, err := utils.NewJSONRequest(route.method, route.path, nil)
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		setup.Echo.ServeHTTP(rec, req)
		assert.NotEqual(t, http.StatusNotFound, rec.Code, "Route %s %s should exist", route.method, route.path)
	}
}
