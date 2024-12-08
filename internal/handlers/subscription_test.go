package handlers

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock implementation of sqlx.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Get(dest interface{}, query string, args ...interface{}) error {
	args = append([]interface{}{dest, query}, args...)
	return m.Called(args...).Error(0)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	args = append([]interface{}{query}, args...)
	return nil, m.Called(args...).Error(0)
}

func TestSubscribe(t *testing.T) {
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
				db.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				db.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"Successfully subscribed"}`,
		},
		{
			name:           "Invalid email",
			email:          "invalid-email",
			mockSetup:      func(db *MockDB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid email address"}`,
		},
		{
			name:  "Duplicate email",
			email: "existing@example.com",
			mockSetup: func(db *MockDB) {
				db.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				db.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"Email already exists"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewBufferString(`{"email":"`+test.email+`"}`))
			req.Header.Set("Content-Type", "application/json")

			db := &MockDB{}
			test.mockSetup(db)

			handler := &SubscriptionHandler{
				DB: db,
			}

			handler.Subscribe(rec, req)

			assert.Equal(t, test.expectedStatus, rec.Code)
			assert.Equal(t, test.expectedBody, rec.Body.String())
		})
	}
}
