package provider

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FirebaseProvider implements push notifications via Firebase Cloud Messaging (FCM)
type FirebaseProvider struct {
	app    *firebase.App
	client *messaging.Client
}

type FirebaseConfig struct {
	ProjectID  string // GCP Project ID
	Credential string // Path to service account JSON file (optional, can use default credentials)
}

func NewFirebaseProvider(ctx context.Context, cfg FirebaseConfig) (*FirebaseProvider, error) {
	opts := []option.ClientOption{}

	// If credential path is provided, use it; otherwise use default credentials
	if cfg.Credential != "" {
		if _, err := os.Stat(cfg.Credential); err != nil {
			return nil, fmt.Errorf("firebase: credential file not found: %w", err)
		}
		opts = append(opts, option.WithCredentialsFile(cfg.Credential))
	}
	// If no credential path provided, Firebase SDK will use default credentials:
	// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable
	// 2. Application Default Credentials (ADC) in GCP environments
	// This allows the SDK to automatically find credentials

	config := &firebase.Config{
		ProjectID: cfg.ProjectID,
	}

	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase: initialize app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: get messaging client: %w", err)
	}

	return &FirebaseProvider{
		app:    app,
		client: client,
	}, nil
}

func (p *FirebaseProvider) Send(ctx context.Context, pushToken, title, body string, metadata map[string]string) error {
	// Build the FCM message
	message := &messaging.Message{
		Token: pushToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: make(map[string]string),
	}

	// Add metadata as data fields (FCM data payload must be string key-value pairs)
	for k, v := range metadata {
		if k != "" && v != "" {
			message.Data[k] = v
		}
	}

	// Send the message
	_, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("firebase: send message: %w", err)
	}

	return nil
}

func (p *FirebaseProvider) Name() string {
	return "firebase"
}
