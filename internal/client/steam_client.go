// Package client provides the Steam Web API HTTP client used by all step implementations.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const DefaultSteamAPIBase = "https://api.steampowered.com"

// SteamClient wraps the Steam Web API with a configurable base URL so
// tests can substitute an httptest.Server.
type SteamClient struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new SteamClient. If baseURL is empty, the default Steam API base is used.
func New(baseURL string) *SteamClient {
	if baseURL == "" {
		baseURL = DefaultSteamAPIBase
	}
	return &SteamClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Get performs an HTTP GET to the Steam API.
func (c *SteamClient) Get(path string, params url.Values) (map[string]any, error) {
	u := c.baseURL + path + "?" + params.Encode()
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("steam API GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("steam API read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam API %s: status %d: %s", path, resp.StatusCode, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("steam API %s: parse JSON: %w", path, err)
	}
	return result, nil
}

// Post performs an HTTP POST to the Steam API.
func (c *SteamClient) Post(path string, params url.Values) (map[string]any, error) {
	resp, err := c.httpClient.PostForm(c.baseURL+path, params)
	if err != nil {
		return nil, fmt.Errorf("steam API POST %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("steam API read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam API %s: status %d: %s", path, resp.StatusCode, body)
	}
	if len(body) == 0 {
		return map[string]any{}, nil
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("steam API %s: parse JSON: %w", path, err)
	}
	return result, nil
}
