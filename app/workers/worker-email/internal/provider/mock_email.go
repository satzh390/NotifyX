package provider

import (
	"context"
)

// MockEmailProvider is a mock provider for testing
type MockEmailProvider struct{}

func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{}
}

func (m *MockEmailProvider) Send(ctx context.Context, to, subject, body string, metadata map[string]string) error {
	// Mock provider - no-op, just returns success
	return nil
}

func (m *MockEmailProvider) Name() string {
	return "mock-email"
}

