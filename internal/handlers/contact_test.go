package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/logger"
)

// MockContactStore is a mock implementation of contact.Store
type MockContactStore struct {
	mock.Mock
}

func (m *MockContactStore) Create(ctx context.Context, submission *contact.Submission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

func (m *MockContactStore) List(ctx context.Context) ([]contact.Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]contact.Submission), args.Error(1)
}

func (m *MockContactStore) GetByID(ctx context.Context, id int64) (*contact.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Submission), args.Error(1)
}

func (m *MockContactStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type createContactTestCase struct {
	name          string
	contact       *contact.Submission
	expectedCode  int
	expectedError string
}

func TestCreateContact(t *testing.T) {
	testCases := []createContactTestCase{
		{
			name: "valid contact",
			contact: &contact.Submission{
				Name:    "John Doe",
				Email:   "john@example.com",
				Message: "Test message",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "invalid contact",
			contact: &contact.Submission{
				Name:    "",
				Email:   "invalid-email",
				Message: "",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			mockStore := new(MockContactStore)
			mockLogger := logger.NewMockLogger()
			handler := NewContactHandler(mockLogger, mockStore)

			// Create request
			jsonBytes, _ := json.Marshal(tc.contact)
			req := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(string(jsonBytes)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set expectations
			if tc.expectedCode == http.StatusCreated {
				mockStore.On("Create", mock.Anything, tc.contact).Return(nil)
			}

			// Test
			err := handler.CreateContact(c)

			// Assert
			if tc.expectedError != "" {
				assert.Error(t, err)
				// Add more specific assertions about the error if needed
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rec.Code)

				var submission contact.Submission
				err := json.Unmarshal(rec.Body.Bytes(), &submission)
				assert.NoError(t, err)
				assert.Equal(t, tc.contact.Name, submission.Name)
				assert.Equal(t, tc.contact.Email, submission.Email)
				assert.Equal(t, tc.contact.Message, submission.Message)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

type listContactsTestCase struct {
	name          string
	expectedCode  int
	expectedError string
	expectedData  []contact.Submission
	setupMock     func(*MockContactStore)
}

func TestListContacts(t *testing.T) {
	sampleSubmission := contact.Submission{
		ID:      1,
		Name:    "John Doe",
		Email:   "john@example.com",
		Message: "Test message",
		Status:  contact.StatusPending,
	}

	testCases := []listContactsTestCase{
		{
			name:         "success",
			expectedCode: http.StatusOK,
			setupMock: func(ms *MockContactStore) {
				ms.On("List", mock.Anything).Return([]contact.Submission{sampleSubmission}, nil)
			},
			expectedData: []contact.Submission{sampleSubmission},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			mockStore := new(MockContactStore)
			mockLogger := logger.NewMockLogger()
			handler := NewContactHandler(mockLogger, mockStore)

			// Setup mock expectations
			if tc.setupMock != nil {
				tc.setupMock(mockStore)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/contact", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Test
			err := handler.GetContacts(c)

			// Assert
			if tc.expectedError != "" {
				assert.Error(t, err)
				// Add more specific assertions about the error if needed
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rec.Code)

				var submissions []contact.Submission
				err := json.Unmarshal(rec.Body.Bytes(), &submissions)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData, submissions)
			}

			mockStore.AssertExpectations(t)
		})
	}
}
