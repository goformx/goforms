package session_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

type noopLogger struct{}

func (l *noopLogger) Debug(string, ...any)                     {}
func (l *noopLogger) Info(string, ...any)                      {}
func (l *noopLogger) Warn(string, ...any)                      {}
func (l *noopLogger) Error(string, ...any)                     {}
func (l *noopLogger) Fatal(string, ...any)                     {}
func (l *noopLogger) With(...any) logging.Logger               { return l }
func (l *noopLogger) WithComponent(string) logging.Logger      { return l }
func (l *noopLogger) WithOperation(string) logging.Logger      { return l }
func (l *noopLogger) WithRequestID(string) logging.Logger      { return l }
func (l *noopLogger) WithUserID(string) logging.Logger         { return l }
func (l *noopLogger) WithError(error) logging.Logger           { return l }
func (l *noopLogger) WithFields(map[string]any) logging.Logger { return l }
func (l *noopLogger) WithFieldsStructured(...logging.Field) logging.Logger {
	return l
}
func (l *noopLogger) DebugWithFields(string, ...logging.Field) {}
func (l *noopLogger) InfoWithFields(string, ...logging.Field)  {}
func (l *noopLogger) WarnWithFields(string, ...logging.Field)  {}
func (l *noopLogger) ErrorWithFields(string, ...logging.Field) {}
func (l *noopLogger) FatalWithFields(string, ...logging.Field) {}
func (l *noopLogger) SanitizeField(string, any) string         { return "" }

func TestManagerStartStopPersistsSessions(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "sessions.json")

	sessionCfg := config.SessionConfig{
		StoreFile:  storeFile,
		MaxAge:     2 * time.Hour,
		CookieName: "goforms-session",
	}
	appCfg := &config.Config{Session: sessionCfg}

	manager := session.NewManager(
		&noopLogger{},
		&session.Config{
			SessionConfig: &sessionCfg,
			Config:        appCfg,
		},
		access.NewManager(&access.Config{DefaultAccess: access.Public}, nil),
	)

	require.NoError(t, manager.Start(context.Background()))

	_, err := manager.CreateSession("user-1", "user@example.com", "user")
	require.NoError(t, err)

	require.NoError(t, manager.Stop(context.Background()))

	_, readErr := os.ReadFile(storeFile)
	require.NoError(t, readErr)
}
