package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"

	mongoStore "github.com/notifyx/core/adapters/mongo"
	"github.com/notifyx/core/domain"
	"github.com/notifyx/workerx/delivery"
	"github.com/notifyx/worker-webhook/config"
	"github.com/notifyx/worker-webhook/internal/provider"
	"github.com/notifyx/worker-webhook/internal/worker"
	workerlib "github.com/notifyx/workerx/worker"
)

func main() {
	configPath := os.Getenv("NOTIFYX_WORKER_WEBHOOK_CONFIG")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize storage
	stores, cleanup, err := mongoStore.NewStoreSet(ctx, mongoStore.Options{
		URI:              cfg.Base.Storage.Mongo.URI,
		Database:         cfg.Base.Storage.Mongo.Database,
		CollectionPrefix: cfg.Base.Storage.Mongo.CollectionPrefix,
	})
	if err != nil {
		logger.Error("failed to initialize storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		_ = cleanup(context.Background())
	}()

	// Initialize Kafka reader
	reader := kafka.NewReader(cfg.Base.ReaderConfig())
	defer reader.Close()

	// Initialize DLQ writer
	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Base.Kafka.Brokers...),
		Topic:    cfg.Base.Kafka.DLQTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer dlqWriter.Close()

	// Initialize result handler
	var resultHandler delivery.ResultHandler
	var handlers []delivery.ResultHandler

	if cfg.Base.Delivery.Mode == workerlib.DeliveryModeMongo || cfg.Base.Delivery.Mode == workerlib.DeliveryModeBoth {
		if cfg.Base.Delivery.Mongo.Enabled {
			mongoHandler := delivery.NewMongoResultHandler(stores.DeliveryTasks, stores.DeliveryLogs)
			handlers = append(handlers, mongoHandler)
			logger.Info("delivery: MongoDB handler enabled")
		}
	}

	if cfg.Base.Delivery.Mode == workerlib.DeliveryModeBroker || cfg.Base.Delivery.Mode == workerlib.DeliveryModeBoth {
		if cfg.Base.Delivery.Broker.Enabled {
			brokerHandler := delivery.NewBrokerResultHandler(delivery.BrokerConfig{
				Brokers:   cfg.Base.Kafka.Brokers,
				TaskTopic: cfg.Base.Delivery.Broker.TaskTopic,
				LogTopic:  cfg.Base.Delivery.Broker.LogTopic,
			})
			handlers = append(handlers, brokerHandler)
			logger.Info("delivery: Broker handler enabled")
		}
	}

	if len(handlers) == 0 {
		logger.Info("delivery: no handlers enabled, results will not be stored")
		resultHandler = nil
	} else if len(handlers) == 1 {
		resultHandler = handlers[0]
	} else {
		resultHandler = delivery.NewCompositeResultHandler(handlers...)
	}

	// Initialize Webhook provider
	var webhookProvider provider.Provider
	switch cfg.Webhook.Provider.Type {
	case "http":
		webhookProvider = provider.NewHTTPProvider(provider.HTTPConfig{
			Timeout: cfg.Webhook.Provider.HTTP.Timeout,
		})
	case "mock":
		webhookProvider = provider.NewMockWebhookProvider()
		logger.Info("using mock webhook provider")
	default:
		logger.Error("unsupported webhook provider type", slog.String("type", cfg.Webhook.Provider.Type))
		os.Exit(1)
	}

	// Create base worker
	baseWorker := workerlib.NewBaseWorker(workerlib.BaseWorkerOptions{
		Reader:        reader,
		DLQ:           dlqWriter,
		TemplateStore: stores.Templates,
		ResultHandler: resultHandler,
		Logger:        logger,
		Channel:       domain.ChannelWebhook,
	})

	// Create Webhook worker
	webhookWorker := worker.NewWebhookWorker(baseWorker, webhookProvider)

	defer func() {
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = webhookWorker.Close()
		if resultHandler != nil {
			_ = resultHandler.Close(timeout)
		}
	}()

	logger.Info("Webhook worker started",
		slog.String("topic", cfg.Base.Kafka.Topic),
		slog.String("groupId", cfg.Base.Kafka.GroupID),
	)

	if err := webhookWorker.Run(ctx); err != nil {
		logger.Error("Webhook worker stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Webhook worker exited cleanly")
}

