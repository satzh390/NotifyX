package provider

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

// ProviderManager manages multiple push providers (one per app) with caching
type ProviderManager struct {
	providers map[string]Provider // Cache: appId -> Provider
	mu        sync.RWMutex
	appStore  storage.AppConfigStore
	customerStore storage.CustomerStore
	logger    *slog.Logger
}

// NewProviderManager creates a new provider manager
func NewProviderManager(appStore storage.AppConfigStore, customerStore storage.CustomerStore, logger *slog.Logger) *ProviderManager {
	return &ProviderManager{
		providers:     make(map[string]Provider),
		appStore:      appStore,
		customerStore: customerStore,
		logger:        logger,
	}
}

// GetProvider gets or creates a provider for the given appId
// It looks up the Customer to get orgID, then loads AppConfig
func (pm *ProviderManager) GetProvider(ctx context.Context, customerID, appID string) (Provider, error) {
	if appID == "" {
		return nil, fmt.Errorf("provider manager: appId is required")
	}

	// Check cache first
	pm.mu.RLock()
	if provider, ok := pm.providers[appID]; ok {
		pm.mu.RUnlock()
		return provider, nil
	}
	pm.mu.RUnlock()

	// Cache miss - need to create provider
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Double-check after acquiring write lock
	if provider, ok := pm.providers[appID]; ok {
		return provider, nil
	}

	// Get Customer to find orgID
	customer, err := pm.customerStore.Get(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("provider manager: get customer: %w", err)
	}

	// Load AppConfig
	appConfig, err := pm.appStore.Get(ctx, customer.OrgID, appID)
	if err != nil {
		return nil, fmt.Errorf("provider manager: get app config: %w", err)
	}

	// Create provider based on AppConfig
	provider, err := pm.createProvider(ctx, appConfig)
	if err != nil {
		return nil, fmt.Errorf("provider manager: create provider: %w", err)
	}

	// Cache the provider
	pm.providers[appID] = provider

	pm.logger.Info("provider manager: created and cached provider",
		slog.String("appId", appID),
		slog.String("orgId", customer.OrgID),
		slog.String("provider", string(appConfig.Provider)),
	)

	return provider, nil
}

// createProvider creates a provider instance based on AppConfig
func (pm *ProviderManager) createProvider(ctx context.Context, appConfig domain.AppConfig) (Provider, error) {
	switch appConfig.Provider {
	case domain.PushProviderAPNS:
		if appConfig.APNS == nil {
			return nil, fmt.Errorf("provider manager: APNS config is required for apns provider")
		}
		return NewAPNSProvider(ctx, APNSConfig{
			KeyID:      appConfig.APNS.KeyID,
			TeamID:     appConfig.APNS.TeamID,
			BundleID:   appConfig.APNS.BundleID,
			KeyPath:    appConfig.APNS.KeyPath,
			Production: appConfig.APNS.Production,
		})

	case domain.PushProviderFirebase:
		if appConfig.Firebase == nil {
			return nil, fmt.Errorf("provider manager: Firebase config is required for firebase provider")
		}
		return NewFirebaseProvider(ctx, FirebaseConfig{
			ProjectID:  appConfig.Firebase.ProjectID,
			Credential: appConfig.Firebase.Credential,
		})

	case domain.PushProviderMock:
		return NewMockPushProvider(), nil

	default:
		return nil, fmt.Errorf("provider manager: unsupported provider type: %s", appConfig.Provider)
	}
}

// Close closes all cached providers (if they implement io.Closer)
func (pm *ProviderManager) Close() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Clear cache (providers themselves don't need explicit cleanup in current implementation)
	pm.providers = make(map[string]Provider)
	return nil
}

// ClearCache clears the provider cache (useful for testing or when app configs change)
func (pm *ProviderManager) ClearCache() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.providers = make(map[string]Provider)
}

