package publicapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents an API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// CheckHealth performs a health check against the API
func (c *Client) CheckHealth() (*HealthResponse, error) {
	url := fmt.Sprintf("%s/health", c.BaseURL)
	
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &health, nil
}

// GetVersion returns the API version (placeholder functionality)
func (c *Client) GetVersion() string {
	return "1.0.0"
} 