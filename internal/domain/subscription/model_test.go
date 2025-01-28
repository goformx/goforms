package subscription

import (
	"testing"
)

func TestSubscription_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sub     *Subscription
		wantErr bool
	}{
		{
			name: "valid subscription",
			sub: &Subscription{
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			sub: &Subscription{
				Name: "Test User",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			sub: &Subscription{
				Email: "invalid-email",
				Name:  "Test User",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			sub: &Subscription{
				Email: "test@example.com",
			},
			wantErr: true,
		},
		{
			name:    "empty subscription",
			sub:     &Subscription{},
			wantErr: true,
		},
		{
			name:    "nil subscription",
			sub:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
		})
	}
}
