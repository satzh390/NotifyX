package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/notifyx/core/domain"
	"github.com/redis/go-redis/v9"
)

type SubscriberCache interface {
	Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, bool)
	Set(ctx context.Context, subscriber domain.Subscriber) error
}

type NoopSubscriberCache struct{}

func (noopCache NoopSubscriberCache) Get(_ context.Context, _ string, _ string) (domain.Subscriber, bool) {
	return domain.Subscriber{}, false
}

func (noopCache NoopSubscriberCache) Set(_ context.Context, _ domain.Subscriber) error {
	return nil
}

type RedisSubscriberCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisSubscriberCache(client *redis.Client, ttl time.Duration) *RedisSubscriberCache {
	return &RedisSubscriberCache{client: client, ttl: ttl}
}

func cacheKey(orgID, subscriberID string) string {
	return "notifyx:subscriber:" + orgID + ":" + subscriberID
}

func (subscriberCache *RedisSubscriberCache) Get(ctx context.Context, orgID, subscriberID string) (domain.Subscriber, bool) {
	payload, err := subscriberCache.client.Get(ctx, cacheKey(orgID, subscriberID)).Bytes()
	if errors.Is(err, redis.Nil) || err != nil || len(payload) == 0 {
		return domain.Subscriber{}, false
	}
	var sub domain.Subscriber
	if err := json.Unmarshal(payload, &sub); err != nil {
		return domain.Subscriber{}, false
	}
	return sub, true
}

func (subscriberCache *RedisSubscriberCache) Set(ctx context.Context, sub domain.Subscriber) error {
	payload, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	return subscriberCache.client.Set(ctx, cacheKey(sub.OrgID, sub.ID), payload, subscriberCache.ttl).Err()
}
