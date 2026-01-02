package provider

import (
	"context"
)

// MockPushProvider is a mock provider for testing
type MockPushProvider struct{}

func NewMockPushProvider() *MockPushProvider {
	return &MockPushProvider{}
}

func (m *MockPushProvider) Send(ctx context.Context, pushToken, title, body string, metadata map[string]string) error {
	// Mock provider - no-op, just returns success
	return nil
}

func (m *MockPushProvider) Name() string {
	return "mock-push"
}

