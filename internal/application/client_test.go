package application_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/application"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestForm(t *testing.T, c *application.Client) string {
	testForm := form.Form{
		Name:    "Test Form",
		Fields:  []form.Field{{Name: "field1", Type: "text"}},
		Options: form.FormOptions{},
	}
	err := c.SubmitForm(t.Context(), testForm)
	require.NoError(t, err)
	forms, err := c.ListForms(t.Context())
	require.NoError(t, err)
	require.Len(t, forms, 1)
	return forms[0].ID
}

func setupTestResponse(t *testing.T, c *application.Client, formID string) string {
	testResponse := form.Response{
		FormID:      formID,
		Values:      map[string]any{"field1": "value1"},
		SubmittedAt: time.Now(),
	}
	err := c.SubmitResponse(t.Context(), formID, testResponse)
	require.NoError(t, err)
	responses, err := c.ListResponses(t.Context(), formID)
	require.NoError(t, err)
	require.Len(t, responses, 1)
	return responses[0].ID
}

func TestClient_SubmitForm(t *testing.T) {
	tests := []struct {
		name    string
		form    form.Form
		wantErr bool
	}{
		{
			name: "valid form",
			form: form.Form{
				Name:    "Test Form",
				Fields:  []form.Field{{Name: "field1", Type: "text"}},
				Options: form.FormOptions{},
			},
			wantErr: false,
		},
		{
			name: "invalid form",
			form: form.Form{
				Name:    "",
				Fields:  []form.Field{},
				Options: form.FormOptions{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			err := c.SubmitForm(t.Context(), tt.form)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetForm(t *testing.T) {
	tests := []struct {
		name    string
		formID  string
		setup   bool
		wantErr bool
	}{
		{
			name:    "valid form ID",
			setup:   true,
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			setup:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			var formID string
			if tt.setup {
				formID = setupTestForm(t, c)
			} else {
				formID = tt.formID
			}
			result, err := c.GetForm(t.Context(), formID)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestClient_ListForms(t *testing.T) {
	c := application.NewClient()
	setupTestForm(t, c)
	forms, err := c.ListForms(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, forms)
}

func TestClient_DeleteForm(t *testing.T) {
	tests := []struct {
		name    string
		formID  string
		setup   bool
		wantErr bool
	}{
		{
			name:    "valid form ID",
			setup:   true,
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			setup:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			var formID string
			if tt.setup {
				formID = setupTestForm(t, c)
			} else {
				formID = tt.formID
			}
			err := c.DeleteForm(t.Context(), formID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_UpdateForm(t *testing.T) {
	tests := []struct {
		name    string
		formID  string
		form    form.Form
		setup   bool
		wantErr bool
	}{
		{
			name:  "valid form update",
			setup: true,
			form: form.Form{
				Name:    "Updated Form",
				Fields:  []form.Field{{Name: "field1", Type: "text"}},
				Options: form.FormOptions{},
			},
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			form:    form.Form{},
			setup:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			var formID string
			if tt.setup {
				formID = setupTestForm(t, c)
			} else {
				formID = tt.formID
			}
			err := c.UpdateForm(t.Context(), formID, tt.form)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_SubmitResponse(t *testing.T) {
	tests := []struct {
		name     string
		formID   string
		response form.Response
		setup    bool
		wantErr  bool
	}{
		{
			name:  "valid response",
			setup: true,
			response: form.Response{
				Values:      map[string]any{"field1": "value1"},
				SubmittedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:     "invalid form ID",
			formID:   "",
			response: form.Response{},
			setup:    false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			var formID string
			if tt.setup {
				formID = setupTestForm(t, c)
				tt.response.FormID = formID
			} else {
				formID = tt.formID
			}
			err := c.SubmitResponse(t.Context(), formID, tt.response)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetResponse(t *testing.T) {
	tests := []struct {
		name       string
		responseID string
		setup      bool
		wantErr    bool
	}{
		{
			name:    "valid response ID",
			setup:   true,
			wantErr: false,
		},
		{
			name:       "invalid response ID",
			responseID: "",
			setup:      false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			var responseID string
			if tt.setup {
				formID := setupTestForm(t, c)
				responseID = setupTestResponse(t, c, formID)
			} else {
				responseID = tt.responseID
			}
			result, err := c.GetResponse(t.Context(), responseID)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestClient_ListResponses(t *testing.T) {
	tests := []struct {
		name    string
		formID  string
		setup   bool
		wantErr bool
	}{
		{
			name:    "valid form ID",
			setup:   true,
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			setup:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			var formID string
			if tt.setup {
				formID = setupTestForm(t, c)
				_ = setupTestResponse(t, c, formID)
			} else {
				formID = tt.formID
			}
			responses, err := c.ListResponses(t.Context(), formID)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, responses)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, responses)
			}
		})
	}
}

func TestClient_DeleteResponse(t *testing.T) {
	tests := []struct {
		name       string
		responseID string
		setup      bool
		wantErr    bool
	}{
		{
			name:    "valid response ID",
			setup:   true,
			wantErr: false,
		},
		{
			name:       "invalid response ID",
			responseID: "",
			setup:      false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			var responseID string
			if tt.setup {
				formID := setupTestForm(t, c)
				responseID = setupTestResponse(t, c, formID)
			} else {
				responseID = tt.responseID
			}
			err := c.DeleteResponse(t.Context(), responseID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
