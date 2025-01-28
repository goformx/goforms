package contact_test

import (
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/contact"
)

func TestValidateSubmission(t *testing.T) {
	tests := []struct {
		name    string
		sub     *contact.Submission
		wantErr bool
	}{
		{
			name: "valid submission",
			sub: &contact.Submission{
				Email:   "test@example.com",
				Name:    "Test User",
				Message: "Test message",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			sub: &contact.Submission{
				Name:    "Test User",
				Message: "Test message",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			sub: &contact.Submission{
				Email:   "invalid-email",
				Name:    "Test User",
				Message: "Test message",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			sub: &contact.Submission{
				Email:   "test@example.com",
				Message: "Test message",
			},
			wantErr: true,
		},
		{
			name: "missing message",
			sub: &contact.Submission{
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: true,
		},
		{
			name:    "empty submission",
			sub:     &contact.Submission{},
			wantErr: true,
		},
		{
			name:    "nil submission",
			sub:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := contact.ValidateSubmission(tt.sub)
			if tt.wantErr && err == nil {
				t.Error("ValidateSubmission() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateSubmission() error = %v, want nil", err)
			}
		})
	}
}
