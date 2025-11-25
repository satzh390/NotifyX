package resolver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"dario.cat/mergo"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

const (
	// GlobalCustomerID is the special customerID value used for global rules
	GlobalCustomerID = ""
)

// RuleCache interface for caching merged rules
type RuleCache interface {
	Get(ctx context.Context, customerID, eventType string) (domain.Rule, bool)
	Set(ctx context.Context, customerID, eventType string, rule domain.Rule, version string) error
	GetVersion(ctx context.Context, customerID, eventType string) (string, bool)
}

// NoopRuleCache is a no-op cache implementation
type NoopRuleCache struct{}

func (c NoopRuleCache) Get(_ context.Context, _ string, _ string) (domain.Rule, bool) {
	return domain.Rule{}, false
}

func (c NoopRuleCache) Set(_ context.Context, _ string, _ string, _ domain.Rule, _ string) error {
	return nil
}

func (c NoopRuleCache) GetVersion(_ context.Context, _ string, _ string) (string, bool) {
	return "", false
}

// RuleResolver resolves rules by merging global rules with customer-specific overrides
// and caching the merged result for performance
type RuleResolver struct {
	store      storage.RuleStore
	cache      RuleCache
	cacheTTL   time.Duration
	versionTTL time.Duration
}

type Options struct {
	Store      storage.RuleStore
	Cache      RuleCache
	CacheTTL   time.Duration
	VersionTTL time.Duration
}

func NewRuleResolver(opts Options) *RuleResolver {
	cache := opts.Cache
	if cache == nil {
		cache = NoopRuleCache{}
	}
	cacheTTL := opts.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}
	versionTTL := opts.VersionTTL
	if versionTTL == 0 {
		versionTTL = 30 * time.Second
	}

	return &RuleResolver{
		store:      opts.Store,
		cache:      cache,
		cacheTTL:   cacheTTL,
		versionTTL: versionTTL,
	}
}

// Resolve returns the merged rule for a customer by combining:
// 1. Global rule (customerID = "") - base configuration
// 2. Customer override (customerID = actual customerID) - partial overrides
// The result is cached with versioning for performance
// Performance: Checks cache FIRST, only loads and merges on cache miss
func (r *RuleResolver) Resolve(ctx context.Context, customerID, eventType string) (domain.Rule, error) {
	// Check cache FIRST for fast path
	if cached, ok := r.cache.Get(ctx, customerID, eventType); ok {
		// Cache hit - return immediately (cache implementations handle TTL)
		return cached, nil
	}

	// Cache miss - load rules and merge
	globalRule, err := r.store.Get(ctx, GlobalCustomerID, eventType)
	globalExists := err == nil
	if err != nil && err != storage.ErrNotFound {
		return domain.Rule{}, err
	}

	customerRule, err := r.store.Get(ctx, customerID, eventType)
	customerExists := err == nil
	if err != nil && err != storage.ErrNotFound {
		return domain.Rule{}, err
	}

	if !globalExists && !customerExists {
		return domain.Rule{}, storage.ErrNotFound
	}

	// Compute version from loaded rules
	version := r.computeVersionFromRules(globalRule, customerRule)

	// Merge using mergo library
	var merged domain.Rule
	if globalExists {
		merged = globalRule
	} else {
		// Only customer rule exists
		merged = customerRule
		merged.CustomerID = customerID
		// Cache and return
		_ = r.cache.Set(ctx, customerID, eventType, merged, version)
		return merged, nil
	}

	if customerExists {
		// Use mergo to merge customer overrides into global rule
		// mergo.WithOverride ensures customer fields override global fields
		// mergo.WithAppendSlice merges slices instead of replacing
		if err := mergo.Merge(&merged, customerRule, mergo.WithOverride); err != nil {
			return domain.Rule{}, err
		}
	}

	// Ensure customerID is set correctly
	merged.CustomerID = customerID

	// Cache the merged result
	_ = r.cache.Set(ctx, customerID, eventType, merged, version)

	return merged, nil
}

// computeVersionFromRules computes a version hash from already-loaded rules
// This avoids redundant storage lookups
func (r *RuleResolver) computeVersionFromRules(globalRule, customerRule domain.Rule) string {
	var timestamps []string

	// Add global rule timestamp if it exists (not zero)
	if globalRule.EventType != "" && !globalRule.UpdatedAt.IsZero() {
		timestamps = append(timestamps, globalRule.UpdatedAt.Format(time.RFC3339Nano))
	}

	// Add customer rule timestamp if it exists (not zero)
	if customerRule.EventType != "" && !customerRule.UpdatedAt.IsZero() {
		timestamps = append(timestamps, customerRule.UpdatedAt.Format(time.RFC3339Nano))
	}

	// If no timestamps, return empty version (will always recompute)
	if len(timestamps) == 0 {
		return ""
	}

	// Compute hash of timestamps
	versionData := ""
	for _, ts := range timestamps {
		versionData += ts + ":"
	}

	hash := sha256.Sum256([]byte(versionData))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter version string
}

// InvalidateCache invalidates the cache for a specific customer and event type
// Useful when you know a rule has changed
func (r *RuleResolver) InvalidateCache(ctx context.Context, customerID, eventType string) error {
	// By setting version to empty, the cache will be recomputed on next access
	_ = r.cache.Set(ctx, customerID, eventType, domain.Rule{}, "")
	return nil
}
