package provider

import (
	"context"
)

// Provider defines the interface for webhook providers
type Provider interface {
	Send(ctx context.Context, url string, payload map[string]interface{}, metadata map[string]string) error
	Name() string
}
