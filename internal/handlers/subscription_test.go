package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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
	args = append([]interface{}{ctx, query}, args...)
	m.Called(args...)

	// Create a new sql.DB connection just for creating a row
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// For the error case, return a row that will fail to scan
	if args[2] == "error@example.com" {
		return db.QueryRow("SELECT NULL WHERE 1=0")
	}

	// Return a row that will scan the ID successfully
	return db.QueryRow("SELECT 1")
}

func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, string)
	}{
		{
			name:  "Valid subscription",
			email: "test@example.com",
			mockSetup: func(db *MockDB) {
				db.On("QueryRowContext",
					mock.Anything,
					mock.Anything,
					"test@example.com",
					mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(1), response["id"])
				assert.Equal(t, "test@example.com", response["email"])
				assert.NotEmpty(t, response["created_at"])
			},
		},
		{
			name:           "Invalid email",
			email:          "",
			mockSetup:      func(db *MockDB) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				assert.JSONEq(t, `{"error":"code=400, message=email is required"}`, body)
			},
		},
		{
			name:  "Database error",
			email: "error@example.com",
			mockSetup: func(db *MockDB) {
				db.On("QueryRowContext",
					mock.Anything,
					mock.Anything,
					"error@example.com",
					mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body string) {
				assert.JSONEq(t, `{"error":"failed to create subscription"}`, body)
			},
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
					test.checkResponse(t, fmt.Sprintf(`{"error":"%v"}`, he.Message))
					return
				}
				t.Errorf("expected HTTPError, got %v", err)
				return
			}

			// Assert
			assert.Equal(t, test.expectedStatus, rec.Code)
			test.checkResponse(t, rec.Body.String())
		})
	}
}
