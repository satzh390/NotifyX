package resolver

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"github.com/notifyx/core/domain"
)

const (
	// DefaultMemoryCacheSize is the default maximum number of rules in memory cache
	DefaultMemoryCacheSize = 5000
)

// MemoryRuleCache implements RuleCache using an in-memory LRU cache
type MemoryRuleCache struct {
	cache        *lru.Cache[string, cachedRule]
	versionCache *lru.Cache[string, cachedVersion]
	ttl          time.Duration
	mu           sync.RWMutex
}

type cachedRule struct {
	rule    domain.Rule
	expires time.Time
}

type cachedVersion struct {
	version string
	expires time.Time
}

// NewMemoryRuleCache creates a new in-memory LRU cache for rules
// size: maximum number of rules to cache (default: 5000)
// ttl: time-to-live for cached rules
func NewMemoryRuleCache(size int, ttl time.Duration) (*MemoryRuleCache, error) {
	if size <= 0 {
		size = DefaultMemoryCacheSize
	}
	
	ruleCache, err := lru.New[string, cachedRule](size)
	if err != nil {
		return nil, err
	}

	versionCache, err := lru.New[string, cachedVersion](size)
	if err != nil {
		return nil, err
	}

	if ttl <= 0 {
		ttl = 5 * time.Minute // Default TTL
	}

	cache := &MemoryRuleCache{
		cache:        ruleCache,
		versionCache: versionCache,
		ttl:          ttl,
	}

	// Start cleanup goroutine to remove expired entries
	go cache.cleanup(ttl / 4) // Cleanup every 1/4 of TTL

	return cache, nil
}

func (c *MemoryRuleCache) Get(ctx context.Context, customerID, eventType string) (domain.Rule, bool) {
	key := cacheKey(customerID, eventType)
	
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.cache.Get(key)
	if !ok {
		return domain.Rule{}, false
	}

	// Check if expired
	if time.Now().After(entry.expires) {
		c.cache.Remove(key)
		c.versionCache.Remove(key)
		return domain.Rule{}, false
	}

	return entry.rule, true
}

func (c *MemoryRuleCache) Set(ctx context.Context, customerID, eventType string, rule domain.Rule, version string) error {
	key := cacheKey(customerID, eventType)
	
	c.mu.Lock()
	defer c.mu.Unlock()

	if rule.EventType == "" {
		// Empty rule means invalidate cache
		c.cache.Remove(key)
		c.versionCache.Remove(key)
		return nil
	}

	entry := cachedRule{
		rule:    rule,
		expires: time.Now().Add(c.ttl),
	}

	c.cache.Add(key, entry)

	if version != "" {
		versionEntry := cachedVersion{
			version: version,
			expires: time.Now().Add(c.ttl),
		}
		c.versionCache.Add(key, versionEntry)
	}

	return nil
}

func (c *MemoryRuleCache) GetVersion(ctx context.Context, customerID, eventType string) (string, bool) {
	key := cacheKey(customerID, eventType)
	
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.versionCache.Get(key)
	if !ok {
		return "", false
	}

	// Check if expired
	if time.Now().After(entry.expires) {
		c.versionCache.Remove(key)
		return "", false
	}

	return entry.version, true
}

// cleanup periodically removes expired entries
func (c *MemoryRuleCache) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		// Clean expired rules
		keys := c.cache.Keys()
		for _, key := range keys {
			if entry, ok := c.cache.Peek(key); ok {
				if now.After(entry.expires) {
					c.cache.Remove(key)
				}
			}
		}

		// Clean expired versions
		versionKeys := c.versionCache.Keys()
		for _, key := range versionKeys {
			if entry, ok := c.versionCache.Peek(key); ok {
				if now.After(entry.expires) {
					c.versionCache.Remove(key)
				}
			}
		}

		c.mu.Unlock()
	}
}

func cacheKey(customerID, eventType string) string {
	return customerID + ":" + eventType
}

