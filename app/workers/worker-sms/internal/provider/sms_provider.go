package provider

import (
	"context"
)

// Provider defines the interface for SMS providers
type Provider interface {
	Send(ctx context.Context, phone, message string, metadata map[string]string) error
	Name() string
}

