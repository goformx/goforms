package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// defaultTimeout is the default HTTP client timeout in seconds
	defaultTimeout = 30
)

// Client represents an HTTP client for interacting with the API
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     logging.Logger
}

// NewClient creates a new API client
func NewClient(baseURL string, logger logging.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout * time.Second,
		},
		logger: logger,
	}
}

// Auth API

// SignUp creates a new user account
func (c *Client) SignUp(ctx context.Context, signup *user.Signup) (*user.User, error) {
	url := fmt.Sprintf("%s/api/v1/auth/signup", c.baseURL)
	resp, err := c.doRequest(ctx, http.MethodPost, url, signup)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var newUser user.User
	if decodeErr := json.NewDecoder(resp.Body).Decode(&newUser); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}
	return &newUser, nil
}

// Login authenticates a user and returns JWT tokens
func (c *Client) Login(ctx context.Context, login *user.Login) (*user.TokenPair, error) {
	url := fmt.Sprintf("%s/api/v1/auth/login", c.baseURL)
	resp, err := c.doRequest(ctx, http.MethodPost, url, login)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenPair user.TokenPair
	if decodeErr := json.NewDecoder(resp.Body).Decode(&tokenPair); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}
	return &tokenPair, nil
}

// Logout invalidates the user's tokens
func (c *Client) Logout(ctx context.Context, token string) error {
	url := fmt.Sprintf("%s/api/v1/auth/logout", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// Version API

// GetVersion retrieves the application version information
func (c *Client) GetVersion(ctx context.Context) (*VersionInfo, error) {
	url := fmt.Sprintf("%s/v1/version", c.baseURL)
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info VersionInfo
	if decodeErr := json.NewDecoder(resp.Body).Decode(&info); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}
	return &info, nil
}

// VersionInfo represents the application version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	GitCommit string `json:"gitCommit"`
	GoVersion string `json:"goVersion"`
}

// Helper methods

func (c *Client) doRequest(ctx context.Context, method, url string, body any) (*http.Response, error) {
	var reqBody []byte
	if body != nil {
		var marshalErr error
		reqBody, marshalErr = json.Marshal(body)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", marshalErr)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Body = http.NoBody
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(reqBody)), nil
		}
		req.Body, _ = req.GetBody()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}
