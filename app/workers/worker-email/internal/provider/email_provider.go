package provider

import (
	"context"
)

// Provider defines the interface for email providers
type Provider interface {
	Send(ctx context.Context, to, subject, body string, metadata map[string]string) error
	Name() string
}

