package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockform "github.com/goformx/goforms/test/mocks/form"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	mockuser "github.com/goformx/goforms/test/mocks/user"
	mockview "github.com/goformx/goforms/test/mocks/view"

	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWebHandler_handleHome(t *testing.T) {
	// Setup
	e := echo.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockLogger := mocklogging.NewMockLogger(ctrl)
	mockUserService := mockuser.NewMockService(ctrl)
	mockFormService := mockform.NewMockService(ctrl)
	mockRenderer := mockview.NewMockRenderer(ctrl)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
	}

	tests := []struct {
		name           string
		setupSession   func(c echo.Context)
		expectedStatus int
		expectedPath   string
	}{
		{
			name: "unauthenticated user should see homepage",
			setupSession: func(c echo.Context) {
				// No session setup - user is unauthenticated
			},
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
		{
			name: "authenticated user should be redirected to dashboard (302)",
			setupSession: func(c echo.Context) {
				// Setup authenticated session
				sess := &session.Session{
					UserID:    "test-user-id",
					Email:     "test@example.com",
					Role:      "user",
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				c.Set(session.SessionKey, sess)
			},
			expectedStatus: http.StatusFound,
			expectedPath:   "/dashboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup handler dependencies
			deps := HandlerDeps{
				Logger:      mockLogger,
				Config:      cfg,
				UserService: mockUserService,
				FormService: mockFormService,
				Renderer:    mockRenderer,
			}

			// Create handler
			handler, err := NewWebHandler(deps)
			assert.NoError(t, err)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Setup session based on test case
			tt.setupSession(c)

			// Setup mock expectations
			mockLogger.EXPECT().
				Debug("handleHome: data.User", "user", gomock.Any()).
				Times(1)

			if tt.expectedStatus == http.StatusOK {
				mockRenderer.EXPECT().
					Render(c, gomock.Any()).
					Return(nil).
					Times(1)
			}

			// Handle request
			err = handler.handleHome(c)

			// Assertions
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Equal(t, tt.expectedPath, rec.Header().Get("Location"))
			}
		})
	}
}
