package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func New() *Client {
	baseURL := os.Getenv("REPLICATED_SDK_URL")
	if baseURL == "" {
		baseURL = "http://statuspage-sdk:3000"
	}
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// --- License ---

type LicenseInfo struct {
	LicenseID    string            `json:"licenseID"`
	CustomerName string            `json:"customerName"`
	LicenseType  string            `json:"licenseType"`
	Entitlements map[string]any    `json:"entitlements"`
	ExpiresAt    string            `json:"expiresAt,omitempty"`
}

type LicenseField struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Value     any    `json:"value"`
	ValueType string `json:"valueType"`
}

func (c *Client) GetLicenseInfo() (*LicenseInfo, error) {
	resp, err := c.get("/api/v1/license/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info LicenseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode license info: %w", err)
	}
	return &info, nil
}

func (c *Client) GetLicenseField(name string) (*LicenseField, error) {
	resp, err := c.get("/api/v1/license/fields/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var field LicenseField
	if err := json.NewDecoder(resp.Body).Decode(&field); err != nil {
		return nil, fmt.Errorf("decode license field: %w", err)
	}
	return &field, nil
}

// --- Updates ---

type Update struct {
	VersionLabel string `json:"versionLabel"`
	CreatedAt    string `json:"createdAt"`
	ReleaseNotes string `json:"releaseNotes"`
}

func (c *Client) GetUpdates() ([]Update, error) {
	resp, err := c.get("/api/v1/app/updates")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var updates []Update
	if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
		return nil, fmt.Errorf("decode updates: %w", err)
	}
	return updates, nil
}

// --- Custom Metrics ---

func (c *Client) SendMetrics(metrics map[string]any) error {
	body, err := json.Marshal(map[string]any{"data": metrics})
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/v1/app/custom-metrics", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send metrics failed (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// --- Instance Tags ---

func (c *Client) SetInstanceTags(tags map[string]string) error {
	body, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"force": true,
			"tags":  tags,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/v1/app/instance-tags", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("set tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("set tags failed: %d", resp.StatusCode)
	}
	return nil
}

// --- Health ---

func (c *Client) Healthy() bool {
	resp, err := c.get("/api/v1/app/info")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *Client) get(path string) (*http.Response, error) {
	resp, err := c.client.Get(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("sdk request %s: %w", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("sdk %s returned %d: %s", path, resp.StatusCode, string(body))
	}
	return resp, nil
}

// StartMetricsReporter periodically sends app metrics to the SDK
func (c *Client) StartMetricsReporter(getMetrics func() map[string]any) {
	go func() {
		// Wait briefly for SDK to be ready on startup
		time.Sleep(10 * time.Second)
		for {
			metrics := getMetrics()
			if err := c.SendMetrics(metrics); err != nil {
				log.Printf("failed to send metrics: %v", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}()
}
