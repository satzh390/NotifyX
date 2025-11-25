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
	ruleCacheKeyPrefix = "notifyx:rule:"
)

// RedisRuleCache implements RuleCache using Redis
type RedisRuleCache struct {
	client   *redis.Client
	cacheTTL time.Duration
}

// NewRedisRuleCache creates a new Redis-based rule cache
func NewRedisRuleCache(client *redis.Client, cacheTTL time.Duration) *RedisRuleCache {
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}
	return &RedisRuleCache{
		client:   client,
		cacheTTL: cacheTTL,
	}
}

func (redisCache *RedisRuleCache) Get(ctx context.Context, customerID, eventType string) (domain.Rule, bool) {
	key := ruleCacheKey(customerID, eventType)
	payload, err := redisCache.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) || err != nil || len(payload) == 0 {
		return domain.Rule{}, false
	}

	var rule domain.Rule
	if err := json.Unmarshal(payload, &rule); err != nil {
		return domain.Rule{}, false
	}

	return rule, true
}

func (redisCache *RedisRuleCache) Set(ctx context.Context, customerID, eventType string, rule domain.Rule) error {
	payload, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("redis cache: marshal rule: %w", err)
	}

	key := ruleCacheKey(customerID, eventType)

	// Store rule with cache TTL
	if err := redisCache.client.Set(ctx, key, payload, redisCache.cacheTTL).Err(); err != nil {
		return fmt.Errorf("redis cache: set rule: %w", err)
	}

	return nil
}

func (redisCache *RedisRuleCache) Delete(ctx context.Context, customerID, eventType string) error {
	key := ruleCacheKey(customerID, eventType)
	return redisCache.client.Del(ctx, key).Err()
}

func ruleCacheKey(customerID, eventType string) string {
	return ruleCacheKeyPrefix + customerID + ":" + eventType
}
