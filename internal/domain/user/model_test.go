package user_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/domain/user"
)

func TestUser_SetPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "testpassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt handles empty strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user.User{}
			err := u.SetPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.SetPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && u.HashedPassword == "" {
				t.Error("User.SetPassword() did not set HashedPassword")
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	tests := []struct {
		name         string
		password     string
		checkAgainst string
		want         bool
	}{
		{
			name:         "correct password",
			password:     "testpassword123",
			checkAgainst: "testpassword123",
			want:         true,
		},
		{
			name:         "incorrect password",
			password:     "testpassword123",
			checkAgainst: "wrongpassword",
			want:         false,
		},
		{
			name:         "empty password check",
			password:     "testpassword123",
			checkAgainst: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user.User{}
			_ = u.SetPassword(tt.password)
			if got := u.CheckPassword(tt.checkAgainst); got != tt.want {
				t.Errorf("User.CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Fields(t *testing.T) {
	now := time.Now()
	u := &user.User{
		ID:             1,
		Email:          "test@example.com",
		FirstName:      "John",
		LastName:       "Doe",
		Role:           "user",
		Active:         true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if u.ID != 1 {
		t.Errorf("User.ID = %v, want %v", u.ID, 1)
	}
	if u.Email != "test@example.com" {
		t.Errorf("User.Email = %v, want %v", u.Email, "test@example.com")
	}
	if u.FirstName != "John" {
		t.Errorf("User.FirstName = %v, want %v", u.FirstName, "John")
	}
	if u.LastName != "Doe" {
		t.Errorf("User.LastName = %v, want %v", u.LastName, "Doe")
	}
	if u.Role != "user" {
		t.Errorf("User.Role = %v, want %v", u.Role, "user")
	}
	if !u.Active {
		t.Error("User.Active = false, want true")
	}
	if !u.CreatedAt.Equal(now) {
		t.Errorf("User.CreatedAt = %v, want %v", u.CreatedAt, now)
	}
	if !u.UpdatedAt.Equal(now) {
		t.Errorf("User.UpdatedAt = %v, want %v", u.UpdatedAt, now)
	}
}
