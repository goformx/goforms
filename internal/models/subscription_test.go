package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptionValidation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "Valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Invalid email - no @",
			email:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "Invalid email - no domain",
			email:   "test@",
			wantErr: true,
		},
		{
			name:    "Invalid email - spaces",
			email:   "test @example.com",
			wantErr: true,
		},
		{
			name:    "Empty email",
			email:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{Email: tt.email}
			err := s.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
