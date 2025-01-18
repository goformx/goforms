package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/response"
)

type ContactSuite struct {
	suite.Suite
	logger logger.Logger
	db     *sqlx.DB
	store  contact.Store
}

func (s *ContactSuite) SetupSuite() {
	var err error

	// Initialize logger
	s.logger = logger.GetLogger()

	// Setup test database connection
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "goforms_test:goforms_test@tcp(localhost:3306)/goforms_test"
	}

	redactedDSN := strings.Replace(dsn, "goforms_test:", "goforms_test:[REDACTED]@", 1)
	s.logger.Info("Attempting to connect to database", logger.String("dsn", redactedDSN))

	s.db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		s.T().Fatalf("Failed to connect to test database: %v", err)
	}

	// Initialize store with test database
	s.store = contact.NewMockStore()

	// Clear test data
	_, err = s.db.Exec("TRUNCATE TABLE contact_submissions")
	if err != nil {
		s.T().Fatalf("Failed to truncate test table: %v", err)
	}
}

func (s *ContactSuite) TearDownSuite() {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.T().Errorf("Error closing database connection: %v", err)
		}
	}
}

func (s *ContactSuite) TestContactIntegration() {
	// Setup test server
	e := echo.New()
	handler := NewContactHandler(s.logger, s.store)

	// Register routes
	e.POST("/api/contact", handler.CreateContact)
	e.GET("/api/contact", handler.GetContacts)

	// Test valid contact submission
	validPayload := `{
		"name": "Test User",
		"email": "test@example.com",
		"message": "This is a test message",
		"status": "pending"
	}`

	// Setup mock expectations
	mockStore := s.store.(*contact.MockStore)
	mockStore.On("Create", mock.Anything, mock.MatchedBy(func(submission *contact.Submission) bool {
		return submission.Name == "Test User" &&
			submission.Email == "test@example.com" &&
			submission.Message == "This is a test message" &&
			submission.Status == contact.StatusPending
	})).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(validPayload))
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	// Parse the wrapped response
	var resp response.Response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	s.NoError(err)
	s.Equal("success", resp.Status)

	// Convert the data field to contact submission
	contactData, err := json.Marshal(resp.Data)
	s.NoError(err)
	var submission contact.Submission
	err = json.Unmarshal(contactData, &submission)
	s.NoError(err)

	s.Equal("Test User", submission.Name)
	s.Equal("test@example.com", submission.Email)
	s.Equal("This is a test message", submission.Message)
	s.Equal(contact.StatusPending, submission.Status)

	// Test invalid submission
	invalidPayload := `{
		"name": "",
		"email": "invalid-email",
		"message": ""
	}`
	req = httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(invalidPayload))
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)

	// Verify error response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	s.NoError(err)
	s.Equal("error", resp.Status)
	s.NotEmpty(resp.Message)

	// Verify mock expectations
	mockStore.AssertExpectations(s.T())
}

func TestContactSuite(t *testing.T) {
	suite.Run(t, new(ContactSuite))
}
