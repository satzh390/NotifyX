package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSMTPProvider_Name(t *testing.T) {
	cfg := SMTPConfig{
		Host: "smtp.example.com",
		Port: "587",
		From: "test@example.com",
	}
	
	provider, err := NewSMTPProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create SMTP provider: %v", err)
	}
	
	if provider.Name() != "smtp" {
		t.Errorf("Expected provider name 'smtp', got '%s'", provider.Name())
	}
}

func TestSMTPProvider_NewSMTPProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  SMTPConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SMTPConfig{
				Host: "smtp.example.com",
				Port: "587",
				From: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: SMTPConfig{
				Host: "",
				Port: "587",
				From: "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing from",
			config: SMTPConfig{
				Host: "smtp.example.com",
				Port: "587",
				From: "",
			},
			wantErr: true,
		},
		{
			name: "default port",
			config: SMTPConfig{
				Host: "smtp.example.com",
				Port: "",
				From: "test@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewSMTPProvider(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if provider != nil {
					t.Error("Expected nil provider on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if provider == nil {
					t.Error("Expected provider but got nil")
				}
			}
		})
	}
}

func TestSMTPProvider_Send(t *testing.T) {
	// This is a placeholder test - actual SMTP testing would require a test SMTP server
	// In a real scenario, you'd use a mock SMTP server or integration test
	t.Run("send email", func(t *testing.T) {
		t.Skip("SMTP provider send test requires a test SMTP server")
	})

	t.Run("send with metadata headers", func(t *testing.T) {
		cfg := SMTPConfig{
			Host:     "smtp.example.com",
			Port:     "587",
			From:     "test@example.com",
			Username: "user",
			Password: "pass",
		}

		provider, err := NewSMTPProvider(cfg)
		require.NoError(t, err)

		// This would test sending with metadata
		// But requires actual SMTP server
		_ = provider
		t.Skip("SMTP provider send test requires a test SMTP server")
	})
}

