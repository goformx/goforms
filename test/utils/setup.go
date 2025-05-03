package utils

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// TestSetup contains common test setup utilities
type TestSetup struct {
	T      *testing.T
	Echo   *echo.Echo
	Logger logging.Logger
}

// NewTestSetup creates a new test setup with common configurations
func NewTestSetup(t *testing.T) *TestSetup {
	require.NotNil(t, t, "testing.T must not be nil")

	e := echo.New()
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	))

	return &TestSetup{
		T:      t,
		Echo:   e,
		Logger: logging.NewZapLogger(logger),
	}
}

// AssertNoError asserts that the given error is nil
func (ts *TestSetup) AssertNoError(err error, msgAndArgs ...any) {
	assert.NoError(ts.T, err, msgAndArgs...)
}

// RequireNoError requires that the given error is nil
func (ts *TestSetup) RequireNoError(err error, msgAndArgs ...any) {
	require.NoError(ts.T, err, msgAndArgs...)
}

// Close performs any necessary cleanup
func (ts *TestSetup) Close() error {
	return nil
}
