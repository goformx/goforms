package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/jonesrussell/goforms/internal/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockContactStore struct {
	mock.Mock
}

func (m *MockContactStore) CreateContact(ctx context.Context, contact *models.ContactSubmission) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactStore) GetContacts(ctx context.Context) ([]models.ContactSubmission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ContactSubmission), args.Error(1)
}

func TestContactHandler_CreateContact(t *testing.T) {
	// Setup
	e := echo.New()
	mockStore := new(MockContactStore)
	mockLogger := logger.NewMockLogger()
	handler := NewContactHandler(mockLogger, mockStore)

	tests := []struct {
		name           string
		contact        *models.ContactSubmission
		setupMock      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful submission",
			contact: &models.ContactSubmission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupMock: func() {
				mockStore.On("CreateContact", mock.Anything, &models.ContactSubmission{
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
				}).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:           "invalid request",
			contact:        nil,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "store error",
			contact: &models.ContactSubmission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupMock: func() {
				mockStore.On("CreateContact", mock.Anything, &models.ContactSubmission{
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
				}).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock
			mockStore.ExpectedCalls = nil
			mockStore.Calls = nil

			// Setup mock expectations
			tc.setupMock()

			// Create request
			var jsonData []byte
			var err error
			if tc.contact != nil {
				jsonData, err = json.Marshal(tc.contact)
				assert.NoError(t, err)
			} else {
				jsonData = []byte(`{"email": "invalid"}`)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/contact", bytes.NewReader(jsonData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Test
			err = handler.CreateContact(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse response
			var resp response.Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tc.expectedError {
				assert.Equal(t, "error", resp.Status)
				assert.NotEmpty(t, resp.Message)
			} else {
				assert.Equal(t, "success", resp.Status)

				// Verify contact data
				contactData, err := json.Marshal(resp.Data)
				assert.NoError(t, err)
				var contact models.ContactSubmission
				err = json.Unmarshal(contactData, &contact)
				assert.NoError(t, err)
				assert.Equal(t, tc.contact.Name, contact.Name)
				assert.Equal(t, tc.contact.Email, contact.Email)
				assert.Equal(t, tc.contact.Message, contact.Message)
			}

			// Verify mock
			mockStore.AssertExpectations(t)
		})
	}
}

func TestContactHandler_GetContacts(t *testing.T) {
	// Setup
	e := echo.New()
	mockStore := new(MockContactStore)
	mockLogger := logger.NewMockLogger()
	handler := NewContactHandler(mockLogger, mockStore)

	tests := []struct {
		name           string
		setupMock      func()
		expectedStatus int
		expectedError  bool
		expectedData   []models.ContactSubmission
	}{
		{
			name: "successful retrieval",
			setupMock: func() {
				mockStore.On("GetContacts", mock.Anything).Return([]models.ContactSubmission{
					{
						ID:      1,
						Name:    "Test User",
						Email:   "test@example.com",
						Message: "Test message",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedData: []models.ContactSubmission{
				{
					ID:      1,
					Name:    "Test User",
					Email:   "test@example.com",
					Message: "Test message",
				},
			},
		},
		{
			name: "store error",
			setupMock: func() {
				mockStore.On("GetContacts", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
			expectedData:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock
			mockStore.ExpectedCalls = nil
			mockStore.Calls = nil

			// Setup mock expectations
			tc.setupMock()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/contact", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Test
			err := handler.GetContacts(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse response
			var resp response.Response
			err = json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tc.expectedError {
				assert.Equal(t, "error", resp.Status)
				assert.NotEmpty(t, resp.Message)
			} else {
				assert.Equal(t, "success", resp.Status)

				// Verify contacts data
				contactsData, err := json.Marshal(resp.Data)
				assert.NoError(t, err)
				var contacts []models.ContactSubmission
				err = json.Unmarshal(contactsData, &contacts)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData, contacts)
			}

			// Verify mock
			mockStore.AssertExpectations(t)
		})
	}
}
