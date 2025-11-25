package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/redis/go-redis/v9"
)

const (
	ruleCacheKeyPrefix    = "notifyx:rule:"
	ruleVersionKeyPrefix  = "notifyx:rule:version:"
)

// RedisRuleCache implements RuleCache using Redis
type RedisRuleCache struct {
	client    *redis.Client
	cacheTTL  time.Duration
	versionTTL time.Duration
}

// NewRedisRuleCache creates a new Redis-based rule cache
func NewRedisRuleCache(client *redis.Client, cacheTTL, versionTTL time.Duration) *RedisRuleCache {
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}
	if versionTTL == 0 {
		versionTTL = 30 * time.Second
	}
	return &RedisRuleCache{
		client:     client,
		cacheTTL:   cacheTTL,
		versionTTL: versionTTL,
	}
}

func (c *RedisRuleCache) Get(ctx context.Context, customerID, eventType string) (domain.Rule, bool) {
	key := ruleCacheKey(customerID, eventType)
	payload, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) || err != nil || len(payload) == 0 {
		return domain.Rule{}, false
	}

	var rule domain.Rule
	if err := json.Unmarshal(payload, &rule); err != nil {
		return domain.Rule{}, false
	}

	return rule, true
}

func (c *RedisRuleCache) Set(ctx context.Context, customerID, eventType string, rule domain.Rule, version string) error {
	if rule.EventType == "" {
		// Empty rule - invalidate cache
		cacheKey := ruleCacheKey(customerID, eventType)
		versionKey := ruleVersionKey(customerID, eventType)
		_ = c.client.Del(ctx, cacheKey)
		_ = c.client.Del(ctx, versionKey)
		return nil
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("redis cache: marshal rule: %w", err)
	}

	cacheKey := ruleCacheKey(customerID, eventType)
	versionKey := ruleVersionKey(customerID, eventType)

	// Store rule with cache TTL
	if err := c.client.Set(ctx, cacheKey, payload, c.cacheTTL).Err(); err != nil {
		return fmt.Errorf("redis cache: set rule: %w", err)
	}

	// Store version separately with longer TTL (if version provided)
	if version != "" {
		if err := c.client.Set(ctx, versionKey, version, c.versionTTL).Err(); err != nil {
			return fmt.Errorf("redis cache: set version: %w", err)
		}
	}

	return nil
}

func (c *RedisRuleCache) GetVersion(ctx context.Context, customerID, eventType string) (string, bool) {
	key := ruleVersionKey(customerID, eventType)
	version, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) || err != nil {
		return "", false
	}
	return version, true
}

func ruleCacheKey(customerID, eventType string) string {
	return ruleCacheKeyPrefix + customerID + ":" + eventType
}

func ruleVersionKey(customerID, eventType string) string {
	return ruleVersionKeyPrefix + customerID + ":" + eventType
}

