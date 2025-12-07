package provider

import (
	"context"
)

// Provider defines the interface for push notification providers
type Provider interface {
	Send(ctx context.Context, pushToken, title, body string, metadata map[string]string) error
	Name() string
}
