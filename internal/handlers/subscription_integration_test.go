package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type SubscriptionSuite struct {
	suite.Suite
	logger *zap.Logger
	db     *sqlx.DB
	store  models.SubscriptionStore
}

func (s *SubscriptionSuite) SetupSuite() {
	s.logger = logger.GetLogger()
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
	s.store = models.NewSubscriptionStore(s.db)

	// Clear test data
	_, err = s.db.Exec("TRUNCATE TABLE subscriptions")
	if err != nil {
		s.T().Fatalf("Failed to truncate test table: %v", err)
	}
}

func (s *SubscriptionSuite) TearDownSuite() {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.T().Errorf("Error closing database connection: %v", err)
		}
	}
}

func (s *SubscriptionSuite) TestSubscriptionIntegration() {
	// Setup test server
	e := echo.New()
	handler := NewSubscriptionHandler(s.store)
	handler.Register(e)

	// Test valid subscription
	validPayload := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(validPayload))
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	var response models.Subscription
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("test@example.com", response.Email)

	// Verify subscription was stored
	var count int
	err = s.db.Get(&count, "SELECT COUNT(*) FROM subscriptions WHERE email = ?", "test@example.com")
	s.NoError(err)
	s.Equal(1, count)

	// Test invalid email
	invalidPayload := `{"email":"invalid-email"}`
	req = httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(invalidPayload))
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func TestSubscriptionSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionSuite))
}
