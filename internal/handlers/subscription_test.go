package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockDB is a mock implementation of the DB interface
type MockDB struct {
	mock.Mock
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	called := m.Called(ctx, query, args)
	if called.Get(0) == nil {
		return &sql.Row{}
	}
	return called.Get(0).(*sql.Row)
}

func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockDB)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "Valid subscription",
			email: "test@example.com",
			mockSetup: func(db *MockDB) {
				db.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Return(&sql.Row{})
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":0,"email":"test@example.com","name":"","created_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:           "Invalid email",
			email:          "",
			mockSetup:      func(db *MockDB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"email is required"}`,
		},
		{
			name:  "Database error",
			email: "error@example.com",
			mockSetup: func(db *MockDB) {
				db.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to create subscription"}`,
		},
		{
			name:           "Invalid email - contains spaces",
			email:          "test @ example.com",
			mockSetup:      func(db *MockDB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid email format"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/subscriptions",
				bytes.NewBufferString(`{"email":"`+test.email+`"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockDB := &MockDB{}
			test.mockSetup(mockDB)
			store := models.NewSubscriptionStore(mockDB)
			logger, _ := zap.NewDevelopment()

			handler := NewSubscriptionHandler(logger, store)

			// Test
			err := handler.CreateSubscription(c)
			if err != nil {
				he, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, test.expectedStatus, he.Code)
					assert.Equal(t, test.expectedBody, fmt.Sprintf(`{"error":"%v"}`, he.Message))
					return
				}
				t.Errorf("expected HTTPError, got %v", err)
				return
			}

			// Assert
			assert.Equal(t, test.expectedStatus, rec.Code)
			assert.JSONEq(t, test.expectedBody, rec.Body.String())
		})
	}
}
