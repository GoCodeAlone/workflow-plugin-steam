package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultSteamAPIBase = "https://api.steampowered.com"

// steamClient wraps the Steam Web API with a configurable base URL so
// tests can substitute an httptest.Server.
type steamClient struct {
	baseURL    string
	httpClient *http.Client
}

func newSteamClient(baseURL string) *steamClient {
	if baseURL == "" {
		baseURL = defaultSteamAPIBase
	}
	return &steamClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *steamClient) get(path string, params url.Values) (map[string]any, error) {
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

func (c *steamClient) post(path string, params url.Values) (map[string]any, error) {
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
