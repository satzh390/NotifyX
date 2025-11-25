package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	mongoStore "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/core/resolver"
	"github.com/notifyx/processor/config"
	"github.com/notifyx/processor/internal/cache"
	"github.com/notifyx/processor/internal/fanout"
	"github.com/notifyx/processor/internal/filter"
	"github.com/notifyx/processor/internal/pipeline"
	"github.com/notifyx/processor/internal/recipients"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to processor config file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	stores, cleanup, err := mongoStore.NewStoreSet(ctx, mongoStore.Options{
		URI:              cfg.Storage.Mongo.URI,
		Database:         cfg.Storage.Mongo.Database,
		CollectionPrefix: cfg.Storage.Mongo.CollectionPrefix,
	})
	if err != nil {
		logger.Error("failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		_ = cleanup(context.Background())
	}()

	subCache := cache.SubscriberCache(cache.NoopSubscriberCache{})
	var redisClient *redis.Client
	var ruleCache resolver.RuleCache

	if cfg.Cache.Redis.Enabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Cache.Redis.Addr,
			DB:       cfg.Cache.Redis.DB,
			Password: cfg.Cache.Redis.Pass,
		})
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("redis disabled due to ping error, using in-memory cache", slog.String("error", err.Error()))
			// Fallback to in-memory LRU cache
			if memCache, err := resolver.NewMemoryRuleCache(5000, cfg.Cache.Redis.TTL); err == nil {
				ruleCache = memCache
				logger.Info("in-memory LRU cache enabled for rules (capacity: 5000)")
			} else {
				ruleCache = resolver.NoopRuleCache{}
				logger.Warn("failed to create in-memory cache, using no-op cache", slog.String("error", err.Error()))
			}
		} else {
			subCache = cache.NewRedisSubscriberCache(redisClient, cfg.Cache.Redis.TTL)
			ruleCache = resolver.NewRedisRuleCache(redisClient, cfg.Cache.Redis.TTL, cfg.Cache.Redis.TTL/6)
			logger.Info("redis cache enabled for subscribers and rules")
		}
		defer func() {
			if redisClient != nil {
				_ = redisClient.Close()
			}
		}()
	} else {
		// No Redis - use in-memory LRU cache
		ttl := cfg.Cache.Redis.TTL
		if ttl <= 0 {
			ttl = 5 * time.Minute // Default TTL if not configured
		}
		if memCache, err := resolver.NewMemoryRuleCache(5000, ttl); err == nil {
			ruleCache = memCache
			logger.Info("in-memory LRU cache enabled for rules (capacity: 5000)")
		} else {
			ruleCache = resolver.NoopRuleCache{}
			logger.Warn("failed to create in-memory cache, using no-op cache", slog.String("error", err.Error()))
		}
	}

	reader := kafka.NewReader(cfg.ReaderConfig())
	defer reader.Close()

	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Kafka.Brokers...),
		Topic:    cfg.Kafka.DLQTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer dlqWriter.Close()

	topics := make(map[domain.ChannelType]string, len(cfg.Worker.Topics))
	for ch, topic := range cfg.Worker.Topics {
		topics[domain.ChannelType(ch)] = topic
	}

	publisher := fanout.NewKafkaPublisher(cfg.Kafka.Brokers, topics)
	defer func() {
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = publisher.Close(timeout)
	}()

	ruleResolver := resolver.NewRuleResolver(resolver.Options{
		Store:      stores.Rules,
		Cache:      ruleCache,
		CacheTTL:   cfg.Cache.Redis.TTL,
		VersionTTL: cfg.Cache.Redis.TTL / 6, // Version TTL is 1/6th of cache TTL for faster invalidation
	})

	proc := pipeline.NewProcessor(pipeline.Options{
		Reader:       reader,
		DLQ:          dlqWriter,
		Resolver:     recipients.NewResolver(stores, subCache),
		RuleResolver: ruleResolver,
		Filter:       filter.NewPreferencesFilter(),
		Fanout:       publisher,
		Stores:       stores,
		Logger:       logger,
	})

	logger.Info("processor started",
		slog.String("topic", cfg.Kafka.InputTopic),
		slog.String("groupId", cfg.Kafka.GroupID),
	)

	if err := proc.Run(ctx); err != nil {
		logger.Error("processor stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("processor exited cleanly")
}
