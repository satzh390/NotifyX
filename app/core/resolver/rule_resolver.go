package resolver

import (
	"context"

	"dario.cat/mergo"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/storage"
)

const (
	// GlobalCustomerID is the special customerID value used for global rules
	GlobalCustomerID = ""
)

// RuleCache interface for caching individual DB rules
type RuleCache interface {
	Get(ctx context.Context, customerID, eventType string) (domain.Rule, bool)
	Set(ctx context.Context, customerID, eventType string, rule domain.Rule) error
	Delete(ctx context.Context, customerID, eventType string) error
}

// NoopRuleCache is a no-op cache implementation
type NoopRuleCache struct{}

func (c NoopRuleCache) Get(_ context.Context, _ string, _ string) (domain.Rule, bool) {
	return domain.Rule{}, false
}

func (c NoopRuleCache) Set(_ context.Context, _ string, _ string, _ domain.Rule) error {
	return nil
}

func (c NoopRuleCache) Delete(_ context.Context, _ string, _ string) error {
	return nil
}

// RuleResolver resolves rules by merging global rules with customer-specific overrides
// Individual rules are cached, but merged results are computed on each request
type RuleResolver struct {
	store storage.RuleStore
	cache RuleCache
}

type Options struct {
	Store storage.RuleStore
	Cache RuleCache
}

func NewRuleResolver(opts Options) *RuleResolver {
	cache := opts.Cache
	if cache == nil {
		cache = NoopRuleCache{}
	}

	return &RuleResolver{
		store: opts.Store,
		cache: cache,
	}
}

// Resolve returns the merged rule for a customer by combining:
// 1. Global rule (customerID = "") - base configuration
// 2. Customer override (customerID = actual customerID) - partial overrides
// Individual rules are cached, but merged result is computed on each request
func (resolver *RuleResolver) Resolve(ctx context.Context, customerID, eventType string) (domain.Rule, error) {
	// Load rules (with caching)
	globalRule, err := resolver.loadRule(ctx, GlobalCustomerID, eventType)
	if err != nil && err != storage.ErrNotFound {
		return domain.Rule{}, err
	}

	customerRule, err := resolver.loadRule(ctx, customerID, eventType)
	if err != nil && err != storage.ErrNotFound {
		return domain.Rule{}, err
	}

	globalExists := globalRule.EventType != ""
	customerExists := customerRule.EventType != ""

	if !globalExists && !customerExists {
		return domain.Rule{}, storage.ErrNotFound
	}

	// Merge rules
	var merged domain.Rule
	if globalExists {
		merged = globalRule
	} else {
		// Only customer rule exists
		merged = customerRule
		merged.CustomerID = customerID
		return merged, nil
	}

	if customerExists {
		// Use mergo to merge customer overrides into global rule
		// mergo.WithOverride ensures customer fields override global fields
		if err := mergo.Merge(&merged, customerRule, mergo.WithOverride); err != nil {
			return domain.Rule{}, err
		}
	}

	// Ensure customerID is set correctly
	merged.CustomerID = customerID

	return merged, nil
}

// loadRule loads a rule from cache or database, caching it if found
func (resolver *RuleResolver) loadRule(ctx context.Context, customerID, eventType string) (domain.Rule, error) {
	// Check cache first
	if cached, ok := resolver.cache.Get(ctx, customerID, eventType); ok {
		return cached, nil
	}

	// Cache miss - load from database
	rule, err := resolver.store.Get(ctx, customerID, eventType)
	if err != nil {
		return domain.Rule{}, err
	}

	// Cache the rule
	_ = resolver.cache.Set(ctx, customerID, eventType, rule)

	return rule, nil
}

// InvalidateCache invalidates the cache for a specific customer and event type
// Useful when you know a rule has changed
func (resolver *RuleResolver) InvalidateCache(ctx context.Context, customerID, eventType string) error {
	return resolver.cache.Delete(ctx, customerID, eventType)
}
