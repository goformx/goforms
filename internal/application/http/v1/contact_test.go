package v1_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "github.com/jonesrussell/goforms/internal/application/http/v1"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	contactmock "github.com/jonesrussell/goforms/test/mocks/contact"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	"github.com/jonesrussell/goforms/test/utils"
)

func TestCreateContact(t *testing.T) {
	tests := []struct {
		name           string
		submission     contact.Submission
		setupMocks     func(*contactmock.MockService, *mocklogging.MockLogger)
		expectedStatus int
	}{
		{
			name: "valid submission",
			submission: contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				sub := &contact.Submission{
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
				}
				ms.ExpectSubmit(context.Background(), sub, nil)
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
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				sub := &contact.Submission{
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
				}
				ms.ExpectSubmit(context.Background(), sub, assert.AnError)
				logger.ExpectError("failed to create contact submission",
					logging.Error(assert.AnError),
				)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid json",
			submission: contact.Submission{},
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				logger.ExpectError("failed to bind contact submission",
					logging.Error(errors.New("json binding error")),
				)
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
			mockLogger := setup.Logger.(*mocklogging.MockLogger)
			tt.setupMocks(mockService, mockLogger)

			api := v1.NewContactAPI(mockService, setup.Logger)

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
			if err := mockService.Verify(); err != nil {
				t.Errorf("service expectations not met: %v", err)
			}
			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

func TestListContacts(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*contactmock.MockService, *mocklogging.MockLogger)
		expectedStatus int
	}{
		{
			name: "success",
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				ms.ExpectListSubmissions(context.Background(), []contact.Submission{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				ms.ExpectListSubmissions(context.Background(), nil, assert.AnError)
				logger.ExpectError("failed to list contact submissions",
					logging.Error(assert.AnError),
				)
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
			mockLogger := setup.Logger.(*mocklogging.MockLogger)
			tt.setupMocks(mockService, mockLogger)

			api := v1.NewContactAPI(mockService, setup.Logger)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
			c, rec := utils.NewTestContext(setup.Echo, req)

			// Execute request
			err := api.ListContacts(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			if err := mockService.Verify(); err != nil {
				t.Errorf("service expectations not met: %v", err)
			}
			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

func TestGetContact(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupMocks     func(*contactmock.MockService, *mocklogging.MockLogger)
		expectedStatus int
	}{
		{
			name: "success",
			id:   "123",
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				ms.ExpectGetSubmission(context.Background(), int64(123), &contact.Submission{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			id:   "123",
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				ms.ExpectGetSubmission(context.Background(), int64(123), nil, assert.AnError)
				logger.ExpectError("failed to get contact submission",
					logging.Error(assert.AnError),
				)
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
			mockLogger := setup.Logger.(*mocklogging.MockLogger)
			tt.setupMocks(mockService, mockLogger)

			api := v1.NewContactAPI(mockService, setup.Logger)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/"+tt.id, nil)
			c, rec := utils.NewTestContext(setup.Echo, req)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			// Execute request
			err := api.GetContact(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			if err := mockService.Verify(); err != nil {
				t.Errorf("service expectations not met: %v", err)
			}
			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

func TestUpdateContactStatus(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		status         string
		setupMocks     func(*contactmock.MockService, *mocklogging.MockLogger)
		expectedStatus int
	}{
		{
			name:   "success",
			id:     "123",
			status: "pending",
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				ms.ExpectUpdateSubmissionStatus(context.Background(), int64(123), contact.StatusPending, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "service error",
			id:     "123",
			status: "pending",
			setupMocks: func(ms *contactmock.MockService, logger *mocklogging.MockLogger) {
				ms.ExpectUpdateSubmissionStatus(context.Background(), int64(123), contact.StatusPending, assert.AnError)
				logger.ExpectError("failed to update contact submission status",
					logging.Error(assert.AnError),
				)
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
			mockLogger := setup.Logger.(*mocklogging.MockLogger)
			tt.setupMocks(mockService, mockLogger)

			api := v1.NewContactAPI(mockService, setup.Logger)

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/api/v1/contacts/"+tt.id+"/status", nil)
			c, rec := utils.NewTestContext(setup.Echo, req)
			c.SetParamNames("id", "status")
			c.SetParamValues(tt.id, tt.status)

			// Execute request
			err := api.UpdateContactStatus(c)
			assert.NoError(t, err)

			// Assert response
			if tt.expectedStatus == http.StatusOK {
				utils.AssertSuccessResponse(t, rec, tt.expectedStatus)
			} else {
				utils.AssertErrorResponse(t, rec, tt.expectedStatus, "")
			}
			if err := mockService.Verify(); err != nil {
				t.Errorf("service expectations not met: %v", err)
			}
			if err := mockLogger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
		})
	}
}

func TestContactRegister(t *testing.T) {
	// Setup
	setup := utils.NewTestSetup()
	defer setup.Close()

	mockService := contactmock.NewMockService()

	// Set up mock expectations for any potential service calls
	mockService.ExpectSubmit(context.Background(), &contact.Submission{}, nil)
	mockService.ExpectListSubmissions(context.Background(), []contact.Submission{}, nil)
	mockService.ExpectGetSubmission(context.Background(), int64(1), &contact.Submission{}, nil)
	mockService.ExpectUpdateSubmissionStatus(context.Background(), int64(1), contact.StatusPending, nil)

	api := v1.NewContactAPI(mockService, setup.Logger)

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
