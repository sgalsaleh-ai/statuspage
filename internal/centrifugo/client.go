package centrifugo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Client struct {
	apiURL    string // internal API endpoint (port 9000 in k8s, 8000 in docker-compose)
	publicURL string // client WebSocket endpoint (port 8000)
	apiKey    string
	client    *http.Client
}

func New() *Client {
	apiURL := os.Getenv("CENTRIFUGO_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8000"
	}
	publicURL := os.Getenv("CENTRIFUGO_PUBLIC_URL")
	if publicURL == "" {
		publicURL = apiURL
	}
	apiKey := os.Getenv("CENTRIFUGO_API_KEY")
	if apiKey == "" {
		apiKey = "statuspage-api-key"
	}
	return &Client{
		apiURL:    apiURL,
		publicURL: publicURL,
		apiKey:    apiKey,
		client:    &http.Client{},
	}
}

type publishRequest struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}

type publishParams struct {
	Channel string `json:"channel"`
	Data    any    `json:"data"`
}

func (c *Client) Publish(channel string, data any) error {
	payload := publishRequest{
		Method: "publish",
		Params: publishParams{
			Channel: channel,
			Data:    data,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal publish request: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL+"/api", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "apikey "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to publish to centrifugo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("centrifugo returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Healthy() bool {
	// Use the API info method to check health
	body := []byte(`{"method":"info"}`)
	req, err := http.NewRequest("POST", c.apiURL+"/api", bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "apikey "+c.apiKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// PublicURL returns the URL clients should connect to for WebSocket
func (c *Client) PublicURL() string {
	return c.publicURL
}
