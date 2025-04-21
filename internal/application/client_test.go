package application_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/application"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		wantErr bool
	}{
		{
			name:    "valid form ID",
			formID:  "test-form",
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			result, err := c.GetForm(t.Context(), tt.formID)
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
	forms, err := c.ListForms(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, forms)
}

func TestClient_DeleteForm(t *testing.T) {
	tests := []struct {
		name    string
		formID  string
		wantErr bool
	}{
		{
			name:    "valid form ID",
			formID:  "test-form",
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			err := c.DeleteForm(t.Context(), tt.formID)
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
		wantErr bool
	}{
		{
			name:   "valid form update",
			formID: "test-form",
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			err := c.UpdateForm(t.Context(), tt.formID, tt.form)
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
		wantErr  bool
	}{
		{
			name:   "valid response",
			formID: "test-form",
			response: form.Response{
				FormID:      "test-form",
				Values:      map[string]any{"field1": "value1"},
				SubmittedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:     "invalid form ID",
			formID:   "",
			response: form.Response{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			err := c.SubmitResponse(t.Context(), tt.formID, tt.response)
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
		wantErr    bool
	}{
		{
			name:       "valid response ID",
			responseID: "test-response",
			wantErr:    false,
		},
		{
			name:       "invalid response ID",
			responseID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			result, err := c.GetResponse(t.Context(), tt.responseID)
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
		wantErr bool
	}{
		{
			name:    "valid form ID",
			formID:  "test-form",
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			responses, err := c.ListResponses(t.Context(), tt.formID)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, responses)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, responses)
			}
		})
	}
}

func TestClient_DeleteResponse(t *testing.T) {
	tests := []struct {
		name       string
		responseID string
		wantErr    bool
	}{
		{
			name:       "valid response ID",
			responseID: "test-response",
			wantErr:    false,
		},
		{
			name:       "invalid response ID",
			responseID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := application.NewClient()
			err := c.DeleteResponse(t.Context(), tt.responseID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
