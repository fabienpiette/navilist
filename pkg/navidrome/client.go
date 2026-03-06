package navidrome

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Client is a Navidrome REST API client with automatic JWT management.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	mu         sync.Mutex
	token      string
	tokenExp   time.Time
}

// New creates a new Client.
func New(baseURL, username, password string) *Client {
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Authenticate logs in and stores the JWT token.
func (c *Client) Authenticate() error {
	body, _ := json.Marshal(authRequest{Username: c.username, Password: c.password})
	resp, err := c.httpClient.Post(c.baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed: status %d", resp.StatusCode)
	}

	var r ndResponse[authResponse]
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("auth decode: %w", err)
	}

	c.mu.Lock()
	c.token = r.Data.Token
	c.tokenExp = time.Now().Add(23 * time.Hour)
	c.mu.Unlock()
	return nil
}

// ensureToken re-authenticates if the token is missing or expiring soon.
func (c *Client) ensureToken() error {
	c.mu.Lock()
	needsRefresh := c.token == "" || time.Now().After(c.tokenExp.Add(-1*time.Minute))
	c.mu.Unlock()
	if needsRefresh {
		return c.Authenticate()
	}
	return nil
}

// Do executes an authenticated HTTP request.
func (c *Client) Do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	return c.httpClient.Do(req)
}
