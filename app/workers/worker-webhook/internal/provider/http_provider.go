package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPProvider implements webhook delivery via HTTP POST requests
type HTTPProvider struct {
	client *http.Client
}

type HTTPConfig struct {
	Timeout time.Duration // Request timeout, defaults to 30s
}

func NewHTTPProvider(cfg HTTPConfig) *HTTPProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPProvider{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *HTTPProvider) Send(ctx context.Context, url string, payload map[string]interface{}, metadata map[string]string) error {
	// Add metadata to payload if provided
	finalPayload := payload
	if len(metadata) > 0 {
		// Create a copy to avoid mutating the original
		finalPayload = make(map[string]interface{})
		for k, v := range payload {
			finalPayload[k] = v
		}
		finalPayload["metadata"] = metadata
	}

	body, err := json.Marshal(finalPayload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add metadata as headers if needed (optional)
	for k, v := range metadata {
		if k != "" && v != "" {
			req.Header.Set("X-"+k, v)
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: received status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (p *HTTPProvider) Name() string {
	return "http"
}
