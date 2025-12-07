package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// APNSProvider implements push notifications via Apple Push Notification Service (APNS)
type APNSProvider struct {
	client   *apns2.Client
	bundleID string
}

type APNSConfig struct {
	KeyID      string // APNS Key ID (for token-based auth)
	TeamID     string // Apple Developer Team ID
	BundleID   string // App Bundle ID
	KeyPath    string // Path to APNS key file (.p8)
	Production bool   // Use production APNS gateway (false = sandbox)
}

func NewAPNSProvider(ctx context.Context, cfg APNSConfig) (*APNSProvider, error) {
	if cfg.KeyPath == "" {
		return nil, fmt.Errorf("apns: key path is required")
	}
	if cfg.KeyID == "" || cfg.TeamID == "" {
		return nil, fmt.Errorf("apns: keyId and teamId are required")
	}
	if cfg.BundleID == "" {
		return nil, fmt.Errorf("apns: bundleId is required")
	}

	// Load the auth key from file
	authKey, err := token.AuthKeyFromFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("apns: load auth key: %w", err)
	}

	// Create token
	apnsToken := &token.Token{
		AuthKey: authKey,
		KeyID:   cfg.KeyID,
		TeamID:  cfg.TeamID,
	}

	// Create APNS client with token authentication
	var client *apns2.Client
	if cfg.Production {
		client = apns2.NewTokenClient(apnsToken).Production()
	} else {
		client = apns2.NewTokenClient(apnsToken).Development()
	}

	return &APNSProvider{
		client:   client,
		bundleID: cfg.BundleID,
	}, nil
}

func (p *APNSProvider) Send(ctx context.Context, pushToken, title, body string, metadata map[string]string) error {
	// Build the APNS payload
	pld := payload.NewPayload().
		AlertTitle(title).
		AlertBody(body).
		Sound("default").
		Badge(1)

	// Add custom data from metadata
	for k, v := range metadata {
		if k != "" && v != "" {
			pld.Custom(k, v)
		}
	}

	// Create the notification
	notification := &apns2.Notification{
		DeviceToken: pushToken,
		Topic:       p.bundleID,
		Payload:     pld,
		Expiration:  time.Now().Add(24 * time.Hour),
		Priority:    apns2.PriorityHigh,
	}

	// Send the notification
	res, err := p.client.PushWithContext(ctx, notification)
	if err != nil {
		return fmt.Errorf("apns: push notification: %w", err)
	}

	if !res.Sent() {
		return fmt.Errorf("apns: notification not sent: %s (reason: %s)", res.Reason, res.ApnsID)
	}

	return nil
}

func (p *APNSProvider) Name() string {
	return "apns"
}
