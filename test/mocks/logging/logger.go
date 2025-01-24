package logging

import (
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// MockLogger is a mock implementation of logging.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...logging.Field)  { m.Called(msg, fields) }
func (m *MockLogger) Error(msg string, fields ...logging.Field) { m.Called(msg, fields) }
func (m *MockLogger) Debug(msg string, fields ...logging.Field) { m.Called(msg, fields) }
func (m *MockLogger) Warn(msg string, fields ...logging.Field)  { m.Called(msg, fields) }
