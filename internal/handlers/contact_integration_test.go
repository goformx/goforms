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
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type ContactSuite struct {
	suite.Suite
	logger *zap.Logger
	db     *sqlx.DB
	store  models.ContactStore
}

func (s *ContactSuite) SetupSuite() {
	var err error

	// Initialize logger first
	s.logger, err = zap.NewDevelopment()
	if err != nil {
		s.T().Fatalf("Failed to create logger: %v", err)
	}

	// Setup test database connection
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "goforms_test:goforms_test@tcp(localhost:3306)/goforms_test"
	}

	redactedDSN := strings.Replace(dsn, "goforms_test:", "goforms_test:[REDACTED]@", 1)
	s.logger.Info("Attempting to connect to database", zap.String("dsn", redactedDSN))

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

	var response models.ContactSubmission
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Test User", response.Name)
	s.Equal("test@example.com", response.Email)
	s.Equal("This is a test message", response.Message)

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
}

func TestContactSuite(t *testing.T) {
	suite.Run(t, new(ContactSuite))
}
