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
	"github.com/notifyx/worker-push/config"
	"github.com/notifyx/worker-push/internal/provider"
	"github.com/notifyx/worker-push/internal/worker"
	workerlib "github.com/notifyx/workerx/worker"
)

func main() {
	configPath := os.Getenv("NOTIFYX_WORKER_PUSH_CONFIG")
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

	// Initialize Push provider
	var pushProvider provider.Provider
	switch cfg.Push.Provider.Type {
	case "firebase":
		fbProvider, err := provider.NewFirebaseProvider(ctx, provider.FirebaseConfig{
			ProjectID:  cfg.Push.Provider.Firebase.ProjectID,
			Credential: cfg.Push.Provider.Firebase.Credential,
		})
		if err != nil {
			logger.Error("failed to create Firebase provider", slog.String("error", err.Error()))
			os.Exit(1)
		}
		pushProvider = fbProvider
		logger.Info("using Firebase push provider", slog.String("projectId", cfg.Push.Provider.Firebase.ProjectID))
	case "apns":
		apnsProvider, err := provider.NewAPNSProvider(ctx, provider.APNSConfig{
			KeyID:      cfg.Push.Provider.APNS.KeyID,
			TeamID:     cfg.Push.Provider.APNS.TeamID,
			BundleID:   cfg.Push.Provider.APNS.BundleID,
			KeyPath:    cfg.Push.Provider.APNS.KeyPath,
			Production: cfg.Push.Provider.APNS.Production,
		})
		if err != nil {
			logger.Error("failed to create APNS provider", slog.String("error", err.Error()))
			os.Exit(1)
		}
		pushProvider = apnsProvider
		logger.Info("using APNS push provider", slog.String("bundleId", cfg.Push.Provider.APNS.BundleID))
	case "mock":
		pushProvider = provider.NewMockPushProvider()
		logger.Info("using mock push provider")
	default:
		logger.Error("unsupported push provider type", slog.String("type", cfg.Push.Provider.Type))
		os.Exit(1)
	}

	// Create base worker
	baseWorker := workerlib.NewBaseWorker(workerlib.BaseWorkerOptions{
		Reader:        reader,
		DLQ:           dlqWriter,
		TemplateStore: stores.Templates,
		ResultHandler: resultHandler,
		Logger:        logger,
		Channel:       domain.ChannelPush,
	})

	// Create Push worker
	pushWorker := worker.NewPushWorker(baseWorker, pushProvider)

	defer func() {
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = pushWorker.Close()
		if resultHandler != nil {
			_ = resultHandler.Close(timeout)
		}
	}()

	logger.Info("Push worker started",
		slog.String("topic", cfg.Base.Kafka.Topic),
		slog.String("groupId", cfg.Base.Kafka.GroupID),
	)

	if err := pushWorker.Run(ctx); err != nil {
		logger.Error("Push worker stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Push worker exited cleanly")
}

