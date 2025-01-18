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
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/jonesrussell/goforms/internal/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ContactSuite struct {
	suite.Suite
	logger logger.Logger
	db     *sqlx.DB
	store  models.ContactStore
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
	s.store = models.NewContactStore(s.db)

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
	handler.Register(e)

	// Test valid contact submission
	validPayload := `{
		"name": "Test User",
		"email": "test@example.com",
		"message": "This is a test message"
	}`

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
	var contact models.ContactSubmission
	err = json.Unmarshal(contactData, &contact)
	s.NoError(err)

	s.Equal("Test User", contact.Name)
	s.Equal("test@example.com", contact.Email)
	s.Equal("This is a test message", contact.Message)

	// Verify submission was stored
	var count int
	err = s.db.Get(&count, "SELECT COUNT(*) FROM contact_submissions WHERE email = ?", "test@example.com")
	s.NoError(err)
	s.Equal(1, count)

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
}

func TestContactSuite(t *testing.T) {
	suite.Run(t, new(ContactSuite))
}
