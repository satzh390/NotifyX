package provider

import (
	"context"
)

// MockSMSProvider is a mock provider for testing
type MockSMSProvider struct{}

func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{}
}

func (m *MockSMSProvider) Send(ctx context.Context, phone, message string, metadata map[string]string) error {
	// Mock provider - no-op, just returns success
	return nil
}

func (m *MockSMSProvider) Name() string {
	return "mock-sms"
}

