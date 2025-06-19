package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware/session"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/presentation/view"
	mockevents "github.com/goformx/goforms/test/mocks/events"
	mockform "github.com/goformx/goforms/test/mocks/form"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	mockuser "github.com/goformx/goforms/test/mocks/user"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFormAPIHandler_SubmitForm(t *testing.T) {
	// Create test form
	testForm := &model.Form{
		ID:          "test-form-1",
		Title:       "Test Form",
		Description: "A test form",
		Schema: model.JSON{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"name"},
		},
	}

	// Create test submission data
	submissionData := map[string]any{
		"name": "John Doe",
	}

	t.Run("successful submission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create fresh mocks for this test
		mockRepo := mockform.NewMockRepository(ctrl)
		mockEventBus := mockevents.NewMockEventBus(ctrl)
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockUserService := mockuser.NewMockService(ctrl)
		mockFormService := mockform.NewMockService(ctrl)
		mockSessionManager := &session.Manager{} // Create a real session manager for tests

		// Setup base handler
		baseHandler := web.NewBaseHandler(mockLogger, nil, mockUserService, mockFormService, view.NewRenderer(mockLogger), mockSessionManager)

		// Setup handler with fresh mocks
		h := &web.FormAPIHandler{
			FormBaseHandler: web.NewFormBaseHandler(baseHandler, formdomain.NewService(mockRepo, mockEventBus, mockLogger)),
		}

		// Setup mock expectations
		mockRepo.EXPECT().
			GetFormByID(gomock.Any(), testForm.ID).
			Return(testForm, nil).
			Times(2)

		mockRepo.EXPECT().
			CreateSubmission(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, submission *model.FormSubmission) error {
				submission.ID = "submission-123"
				assert.Equal(t, testForm.ID, submission.FormID)
				assert.Equal(t, submissionData["name"], submission.Data["name"])
				return nil
			})

		mockEventBus.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			Times(3)

		// Create request
		reqBody, _ := json.Marshal(submissionData)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		c.SetPath("/forms/:id/submit")
		c.SetParamNames("id")
		c.SetParamValues(testForm.ID)

		// Execute handler
		err := h.HandleFormSubmit(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify response
		var response map[string]any
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Form submitted successfully", response["message"])
		assert.NotEmpty(t, response["data"].(map[string]any)["submission_id"])
	})

	t.Run("form not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create fresh mocks for this test
		mockRepo := mockform.NewMockRepository(ctrl)
		mockEventBus := mockevents.NewMockEventBus(ctrl)
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockUserService := mockuser.NewMockService(ctrl)
		mockFormService := mockform.NewMockService(ctrl)
		mockSessionManager := &session.Manager{} // Create a real session manager for tests

		// Setup base handler
		baseHandler := web.NewBaseHandler(mockLogger, nil, mockUserService, mockFormService, view.NewRenderer(mockLogger), mockSessionManager)

		// Setup handler with fresh mocks
		h := &web.FormAPIHandler{
			FormBaseHandler: web.NewFormBaseHandler(baseHandler, formdomain.NewService(mockRepo, mockEventBus, mockLogger)),
		}

		// Setup mock expectations
		mockRepo.EXPECT().
			GetFormByID(gomock.Any(), testForm.ID).
			Return(nil, nil)

		// Create request
		reqBody, _ := json.Marshal(submissionData)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		c.SetPath("/forms/:id/submit")
		c.SetParamNames("id")
		c.SetParamValues(testForm.ID)

		// Execute handler
		err := h.HandleFormSubmit(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid submission data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create fresh mocks for this test
		mockRepo := mockform.NewMockRepository(ctrl)
		mockEventBus := mockevents.NewMockEventBus(ctrl)
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockUserService := mockuser.NewMockService(ctrl)
		mockFormService := mockform.NewMockService(ctrl)
		mockSessionManager := &session.Manager{} // Create a real session manager for tests

		// Setup base handler
		baseHandler := web.NewBaseHandler(mockLogger, nil, mockUserService, mockFormService, view.NewRenderer(mockLogger), mockSessionManager)

		// Setup handler with fresh mocks
		h := &web.FormAPIHandler{
			FormBaseHandler: web.NewFormBaseHandler(baseHandler, formdomain.NewService(mockRepo, mockEventBus, mockLogger)),
		}

		// Setup mock expectations
		mockRepo.EXPECT().
			GetFormByID(gomock.Any(), testForm.ID).
			Return(testForm, nil)

		mockLogger.EXPECT().
			Error("failed to decode submission data", "error", gomock.Any()).
			Times(1)

		// Create request with invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("invalid-json")))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		c.SetPath("/forms/:id/submit")
		c.SetParamNames("id")
		c.SetParamValues(testForm.ID)

		// Execute handler
		err := h.HandleFormSubmit(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create fresh mocks for this test
		mockRepo := mockform.NewMockRepository(ctrl)
		mockEventBus := mockevents.NewMockEventBus(ctrl)
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockUserService := mockuser.NewMockService(ctrl)
		mockFormService := mockform.NewMockService(ctrl)
		mockSessionManager := &session.Manager{} // Create a real session manager for tests

		// Setup base handler
		baseHandler := web.NewBaseHandler(mockLogger, nil, mockUserService, mockFormService, view.NewRenderer(mockLogger), mockSessionManager)

		// Setup handler with fresh mocks
		h := &web.FormAPIHandler{
			FormBaseHandler: web.NewFormBaseHandler(baseHandler, formdomain.NewService(mockRepo, mockEventBus, mockLogger)),
		}

		// Setup mock expectations
		mockRepo.EXPECT().
			GetFormByID(gomock.Any(), testForm.ID).
			Return(testForm, nil).
			Times(2)

		mockRepo.EXPECT().
			CreateSubmission(gomock.Any(), gomock.Any()).
			Return(errors.New("database error"))

		mockLogger.EXPECT().
			Error("failed to submit form", "error", gomock.Any()).
			Times(1)

		// Create request
		reqBody, _ := json.Marshal(submissionData)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		c.SetPath("/forms/:id/submit")
		c.SetParamNames("id")
		c.SetParamValues(testForm.ID)

		// Execute handler
		err := h.HandleFormSubmit(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
