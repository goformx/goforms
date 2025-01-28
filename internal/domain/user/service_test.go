package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/user"
	mock_logging "github.com/jonesrussell/goforms/test/mocks/logging"
	mock_store "github.com/jonesrussell/goforms/test/mocks/store/user"
)

func TestSignUp(t *testing.T) {
	tests := []struct {
		name      string
		signup    *user.Signup
		setupMock func(*mock_store.UserStore)
		wantErr   string
	}{
		{
			name: "successful signup",
			signup: &user.Signup{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(s *mock_store.UserStore) {},
			wantErr:   "",
		},
		{
			name: "email already exists",
			signup: &user.Signup{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(s *mock_store.UserStore) {
				existingUser := &user.User{
					ID:    1,
					Email: "existing@example.com",
				}
				s.Create(existingUser)
			},
			wantErr: "failed to create user: email already exists",
		},
		{
			name: "store error",
			signup: &user.Signup{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock: func(s *mock_store.UserStore) {
				s.SetError("create", errors.New("store error"))
			},
			wantErr: "failed to create user: store error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := mock_store.NewUserStore()
			logger := mock_logging.NewMockLogger()
			tt.setupMock(store)

			service := user.NewService(store, logger)
			u, err := service.SignUp(context.Background(), tt.signup)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if u == nil {
				t.Error("expected user to be created, got nil")
				return
			}

			if u.Email != tt.signup.Email {
				t.Errorf("expected email %s, got %s", tt.signup.Email, u.Email)
			}
			if u.FirstName != tt.signup.FirstName {
				t.Errorf("expected first name %s, got %s", tt.signup.FirstName, u.FirstName)
			}
			if u.LastName != tt.signup.LastName {
				t.Errorf("expected last name %s, got %s", tt.signup.LastName, u.LastName)
			}
			if u.Role != "user" {
				t.Errorf("expected role 'user', got %s", u.Role)
			}
			if !u.Active {
				t.Error("expected user to be active")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name      string
		login     *user.Login
		setupMock func(*mock_store.UserStore)
		wantErr   string
	}{
		{
			name: "successful login",
			login: &user.Login{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(s *mock_store.UserStore) {
				u := &user.User{
					ID:    1,
					Email: "test@example.com",
					Role:  "user",
				}
				u.SetPassword("password123")
				s.Create(u)
			},
			wantErr: "",
		},
		{
			name: "user not found",
			login: &user.Login{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMock: func(s *mock_store.UserStore) {},
			wantErr:   "failed to login: invalid credentials",
		},
		{
			name: "invalid password",
			login: &user.Login{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(s *mock_store.UserStore) {
				u := &user.User{
					ID:    1,
					Email: "test@example.com",
					Role:  "user",
				}
				u.SetPassword("password123")
				s.Create(u)
			},
			wantErr: "failed to login: invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := mock_store.NewUserStore()
			logger := mock_logging.NewMockLogger()
			tt.setupMock(store)

			service := user.NewService(store, logger)
			tokens, err := service.Login(context.Background(), tt.login)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tokens == nil {
				t.Error("expected tokens to be returned, got nil")
			}
		})
	}
}

func TestLogout(t *testing.T) {
	store := mock_store.NewUserStore()
	logger := mock_logging.NewMockLogger()
	service := user.NewService(store, logger)

	// Create a valid test user with proper password hash
	u := &user.User{
		ID:       1,
		Email:    "test@example.com",
		Role:     "user",
		Active:   true,
	}
	if err := u.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}
	store.Create(u)

	login := &user.Login{
		Email:    "test@example.com",
		Password: "password123",
	}
	tokens, err := service.Login(context.Background(), login)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr string
	}{
		{
			name:    "successful logout",
			token:   tokens.AccessToken,
			wantErr: "",
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: "failed to logout: invalid token",
		},
		{
			name:    "already blacklisted token",
			token:   tokens.AccessToken,
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Logout(context.Background(), tt.token)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify token is blacklisted
			if !service.IsTokenBlacklisted(tt.token) {
				t.Error("expected token to be blacklisted")
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupMock func(*mock_store.UserStore)
		wantErr   string
	}{
		{
			name:   "successful get",
			userID: 1,
			setupMock: func(s *mock_store.UserStore) {
				u := &user.User{
					ID:    1,
					Email: "test@example.com",
				}
				s.Create(u)
			},
			wantErr: "",
		},
		{
			name:      "user not found",
			userID:    999,
			setupMock: func(s *mock_store.UserStore) {},
			wantErr:   "failed to get user: user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := mock_store.NewUserStore()
			logger := mock_logging.NewMockLogger()
			tt.setupMock(store)

			service := user.NewService(store, logger)
			u, err := service.GetUserByID(context.Background(), tt.userID)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if u == nil {
				t.Error("expected user to be returned, got nil")
				return
			}

			if u.ID != tt.userID {
				t.Errorf("expected user ID %d, got %d", tt.userID, u.ID)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mock_store.UserStore)
		wantErr   string
	}{
		{
			name: "successful list",
			setupMock: func(s *mock_store.UserStore) {
				u1 := &user.User{ID: 1, Email: "test1@example.com"}
				u2 := &user.User{ID: 2, Email: "test2@example.com"}
				s.Create(u1)
				s.Create(u2)
			},
			wantErr: "",
		},
		{
			name: "store error",
			setupMock: func(s *mock_store.UserStore) {
				s.SetError("list", errors.New("store error"))
			},
			wantErr: "failed to list users: store error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := mock_store.NewUserStore()
			logger := mock_logging.NewMockLogger()
			tt.setupMock(store)

			service := user.NewService(store, logger)
			users, err := service.ListUsers(context.Background())

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(users) == 0 {
				t.Error("expected users to be returned, got empty list")
			}
		})
	}
}
