package subscription_test

import (
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

func TestSubscription_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sub     *subscription.Subscription
		wantErr bool
	}{
		{
			name: "valid subscription",
			sub: &subscription.Subscription{
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			sub: &subscription.Subscription{
				Name: "Test User",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			sub: &subscription.Subscription{
				Email: "invalid-email",
				Name:  "Test User",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			sub: &subscription.Subscription{
				Email: "test@example.com",
			},
			wantErr: true,
		},
		{
			name:    "empty subscription",
			sub:     &subscription.Subscription{},
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
