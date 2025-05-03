package application_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/application"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestForm(t *testing.T, c *application.Client) string {
	testForm := form.Form{
		Title:       "Test Form",
		Description: "Test Description",
		Schema:      form.JSON{"fields": []map[string]any{{"name": "field1", "type": "text"}}},
	}
	err := c.SubmitForm(t.Context(), testForm)
	require.NoError(t, err)
	forms, err := c.ListForms(t.Context())
	require.NoError(t, err)
	require.Len(t, forms, 1)
	return strconv.FormatUint(uint64(forms[0].ID), 10)
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
				Title:       "Test Form",
				Description: "Test Description",
				Schema:      form.JSON{"fields": []map[string]any{{"name": "field1", "type": "text"}}},
			},
			wantErr: false,
		},
		{
			name: "invalid form - missing title",
			form: form.Form{
				Title:       "",
				Description: "Test Description",
				Schema:      form.JSON{"fields": []map[string]any{{"name": "field1", "type": "text"}}},
			},
			wantErr: true,
		},
		{
			name: "invalid form - missing schema",
			form: form.Form{
				Title:       "Test Form",
				Description: "Test Description",
				Schema:      nil,
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
	c := application.NewClient()
	formID := setupTestForm(t, c)

	tests := []struct {
		name    string
		formID  string
		wantErr bool
	}{
		{
			name:    "valid form ID",
			formID:  formID,
			wantErr: false,
		},
		{
			name:    "invalid form ID",
			formID:  "",
			wantErr: true,
		},
		{
			name:    "non-existent form ID",
			formID:  "999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := c.GetForm(t.Context(), tt.formID)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, f)
			} else {
				require.NoError(t, err)
				require.NotNil(t, f)
				assert.Equal(t, formID, strconv.FormatUint(uint64(f.ID), 10))
			}
		})
	}
}

func TestClient_ListForms(t *testing.T) {
	c := application.NewClient()
	formID := setupTestForm(t, c)

	forms, err := c.ListForms(t.Context())
	require.NoError(t, err)
	require.Len(t, forms, 1)
	assert.Equal(t, formID, strconv.FormatUint(uint64(forms[0].ID), 10))
}

func TestClient_DeleteForm(t *testing.T) {
	c := application.NewClient()
	formID := setupTestForm(t, c)

	err := c.DeleteForm(t.Context(), formID)
	require.NoError(t, err)

	forms, err := c.ListForms(t.Context())
	require.NoError(t, err)
	require.Empty(t, forms)
}

func TestClient_UpdateForm(t *testing.T) {
	c := application.NewClient()
	formID := setupTestForm(t, c)

	updatedForm := form.Form{
		Title:       "Updated Form",
		Description: "Updated Description",
		Schema:      form.JSON{"fields": []map[string]any{{"name": "field2", "type": "number"}}},
	}

	err := c.UpdateForm(t.Context(), formID, updatedForm)
	require.NoError(t, err)

	f, err := c.GetForm(t.Context(), formID)
	require.NoError(t, err)
	require.NotNil(t, f)
	assert.Equal(t, "Updated Form", f.Title)
	assert.Equal(t, "Updated Description", f.Description)
}

func TestClient_SubmitResponse(t *testing.T) {
	c := application.NewClient()
	formID := setupTestForm(t, c)

	testResponse := form.Response{
		Values: map[string]any{"field1": "value1"},
	}

	err := c.SubmitResponse(t.Context(), formID, testResponse)
	require.NoError(t, err)

	responses, err := c.ListResponses(t.Context(), formID)
	require.NoError(t, err)
	require.Len(t, responses, 1)
	assert.Equal(t, formID, responses[0].FormID)
	assert.Equal(t, "value1", responses[0].Values["field1"])
}

func TestClient_GetResponse(t *testing.T) {
	c := application.NewClient()
	formID := setupTestForm(t, c)
	responseID := setupTestResponse(t, c, formID)

	r, err := c.GetResponse(t.Context(), responseID)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, formID, r.FormID)
	assert.Equal(t, "value1", r.Values["field1"])
}

func TestClient_ListResponses(t *testing.T) {
	c := application.NewClient()
	formID := setupTestForm(t, c)
	setupTestResponse(t, c, formID)

	responses, err := c.ListResponses(t.Context(), formID)
	require.NoError(t, err)
	require.Len(t, responses, 1)
	assert.Equal(t, formID, responses[0].FormID)
	assert.Equal(t, "value1", responses[0].Values["field1"])
}

func TestClient_DeleteResponse(t *testing.T) {
	c := application.NewClient()
	formID := setupTestForm(t, c)
	responseID := setupTestResponse(t, c, formID)

	err := c.DeleteResponse(t.Context(), responseID)
	require.NoError(t, err)

	responses, err := c.ListResponses(t.Context(), formID)
	require.NoError(t, err)
	require.Empty(t, responses)
}
