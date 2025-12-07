package provider

import (
	"context"
)

// MockWebhookProvider is a mock provider for testing
type MockWebhookProvider struct{}

func NewMockWebhookProvider() *MockWebhookProvider {
	return &MockWebhookProvider{}
}

func (m *MockWebhookProvider) Send(ctx context.Context, url string, payload map[string]interface{}, metadata map[string]string) error {
	// Mock provider - no-op, just returns success
	return nil
}

func (m *MockWebhookProvider) Name() string {
	return "mock-webhook"
}

