package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with the AI emulator service.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// EmulationRequest is sent to the AI emulator.
type EmulationRequest struct {
	Service         string         `json:"service"`
	Protocol        string         `json:"protocol,omitempty"`
	Context         map[string]any `json:"context,omitempty"`
	AttackerProfile map[string]any `json:"attacker_profile,omitempty"`
	Temperature    float64        `json:"temperature,omitempty"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
}

// EmulationResponse is returned by the AI emulator.
type EmulationResponse struct {
	Response   string  `json:"response"`
	Service    string  `json:"service"`
	Model      string  `json:"model"`
	Cached     bool    `json:"cached"`
	LatencyMs  float64 `json:"latency_ms"`
	Intent     string  `json:"intent"`
	Confidence float64 `json:"confidence"`
}

// HealthResponse from the AI emulator.
type HealthResponse struct {
	Status          string   `json:"status"`
	Model           string   `json:"model"`
	ModelAvailable  bool     `json:"model_available"`
	AvailableModels []string `json:"available_models"`
}

// NewClient creates a new AI emulator client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generate sends a request to the AI emulator and returns the dynamic response.
func (c *Client) Generate(ctx context.Context, req EmulationRequest) (*EmulationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/ai-response", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ai emulator request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ai emulator returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result EmulationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Health checks the AI emulator service status.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/ai/health", nil)
	if err != nil {
		return nil, fmt.Errorf("create health request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return &HealthResponse{Status: "unhealthy"}, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode health response: %w", err)
	}

	return &result, nil
}