package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/presentation/handlers/auth"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRequestParser_ParseLogin(t *testing.T) {
	parser := auth.NewAuthRequestParser()

	tests := []struct {
		name          string
		contentType   string
		body          string
		expectedEmail string
		expectedPass  string
		expectError   bool
	}{
		{
			name:          "form data login",
			contentType:   "application/x-www-form-urlencoded",
			body:          "email=test@example.com&password=password123",
			expectedEmail: "test@example.com",
			expectedPass:  "password123",
			expectError:   false,
		},
		{
			name:          "JSON login",
			contentType:   "application/json",
			body:          `{"email":"test@example.com","password":"password123"}`,
			expectedEmail: "test@example.com",
			expectedPass:  "password123",
			expectError:   false,
		},
		{
			name:        "invalid JSON",
			contentType: "application/json",
			body:        `{"email":"test@example.com"`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Create Echo context
			e := echo.New()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Parse login
			email, password, err := parser.ParseLogin(c)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedEmail, email)
				assert.Equal(t, tt.expectedPass, password)
			}
		})
	}
}

func TestAuthRequestParser_ParseSignup(t *testing.T) {
	parser := auth.NewAuthRequestParser()

	tests := []struct {
		name        string
		contentType string
		body        string
		expectError bool
	}{
		{
			name:        "form data signup",
			contentType: "application/x-www-form-urlencoded",
			body:        "email=test@example.com&password=password123&confirm_password=password123",
			expectError: false,
		},
		{
			name:        "JSON signup",
			contentType: "application/json",
			body:        `{"email":"test@example.com","password":"password123","confirm_password":"password123"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Create Echo context
			e := echo.New()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Parse signup
			signup, err := parser.ParseSignup(c)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "test@example.com", signup.Email)
				assert.Equal(t, "password123", signup.Password)
				assert.Equal(t, "password123", signup.ConfirmPassword)
			}
		})
	}
}

func TestAuthRequestParser_ValidateLogin(t *testing.T) {
	parser := auth.NewAuthRequestParser()

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
	}{
		{
			name:        "valid credentials",
			email:       "test@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "missing email",
			email:       "",
			password:    "password123",
			expectError: true,
		},
		{
			name:        "missing password",
			email:       "test@example.com",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateLogin(tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthRequestParser_ValidateSignup(t *testing.T) {
	parser := auth.NewAuthRequestParser()

	tests := []struct {
		name        string
		signup      user.Signup
		expectError bool
	}{
		{
			name: "valid signup",
			signup: user.Signup{
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			expectError: false,
		},
		{
			name: "passwords don't match",
			signup: user.Signup{
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "different",
			},
			expectError: true,
		},
		{
			name: "missing email",
			signup: user.Signup{
				Email:           "",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateSignup(tt.signup)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
