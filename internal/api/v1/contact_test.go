package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateContact(t *testing.T) {
	tests := []struct {
		name           string
		submission     contact.Submission
		setupFn        func(*contact.MockService)
		expectedStatus int
	}{
		{
			name: "valid submission",
			submission: contact.Submission{
				Name:    "Test User",
				Email:   "test@example.com",
				Message: "Test message",
			},
			setupFn: func(ms *contact.MockService) {
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
			setupFn: func(ms *contact.MockService) {
				ms.On("CreateSubmission", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := contact.NewMockService()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, mockLogger)
			e := echo.New()

			body, _ := json.Marshal(tt.submission)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = api.CreateContact(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestListContacts(t *testing.T) {
	tests := []struct {
		name           string
		setupFn        func(*contact.MockService)
		expectedStatus int
	}{
		{
			name: "successful list",
			setupFn: func(ms *contact.MockService) {
				ms.On("ListSubmissions", mock.Anything).Return([]contact.Submission{
					{ID: 1, Name: "Test User", Email: "test@example.com"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupFn: func(ms *contact.MockService) {
				ms.On("ListSubmissions", mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := contact.NewMockService()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, mockLogger)
			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = api.ListContacts(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetContact(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupFn        func(*contact.MockService)
		expectedStatus int
	}{
		{
			name: "existing submission",
			id:   "1",
			setupFn: func(ms *contact.MockService) {
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
			setupFn: func(ms *contact.MockService) {
				ms.On("GetSubmission", mock.Anything, int64(999)).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid id",
			id:   "invalid",
			setupFn: func(ms *contact.MockService) {
				// No mock setup needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := contact.NewMockService()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, mockLogger)
			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			_ = api.GetContact(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateContactStatus(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		status         contact.Status
		setupFn        func(*contact.MockService)
		expectedStatus int
	}{
		{
			name:   "valid update",
			id:     "1",
			status: contact.StatusApproved,
			setupFn: func(ms *contact.MockService) {
				ms.On("UpdateSubmissionStatus", mock.Anything, int64(1), contact.StatusApproved).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "service error",
			id:     "1",
			status: contact.StatusApproved,
			setupFn: func(ms *contact.MockService) {
				ms.On("UpdateSubmissionStatus", mock.Anything, int64(1), contact.StatusApproved).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "invalid id",
			id:     "invalid",
			status: contact.StatusApproved,
			setupFn: func(ms *contact.MockService) {
				// No mock setup needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := contact.NewMockService()
			mockLogger := logger.NewMockLogger()
			tt.setupFn(mockService)

			api := NewContactAPI(mockService, mockLogger)
			e := echo.New()

			body, _ := json.Marshal(map[string]string{"status": string(tt.status)})
			req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			_ = api.UpdateContactStatus(c)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}
