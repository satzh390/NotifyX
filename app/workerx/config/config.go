package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

type Config struct {
	Kafka struct {
		Brokers        []string      `mapstructure:"brokers"`
		GroupID        string        `mapstructure:"groupId"`
		Topic          string        `mapstructure:"topic"`
		DLQTopic       string        `mapstructure:"dlqTopic"`
		MinBytes       int           `mapstructure:"minBytes"`
		MaxBytes       int           `mapstructure:"maxBytes"`
		MaxWait        time.Duration `mapstructure:"maxWait"`
		CommitInterval time.Duration `mapstructure:"commitInterval"`
	} `mapstructure:"kafka"`
	Storage struct {
		Mongo struct {
			URI              string `mapstructure:"uri"`
			Database         string `mapstructure:"database"`
			CollectionPrefix string `mapstructure:"collectionPrefix"`
		} `mapstructure:"mongo"`
	} `mapstructure:"storage"`
	Delivery struct {
		Mode string `mapstructure:"mode"` // "mongo", "broker", "both", or "none"
		Mongo struct {
			Enabled bool `mapstructure:"enabled"`
		} `mapstructure:"mongo"`
		Broker struct {
			Enabled  bool   `mapstructure:"enabled"`
			TaskTopic string `mapstructure:"taskTopic"`
			LogTopic  string `mapstructure:"logTopic"`
		} `mapstructure:"broker"`
	} `mapstructure:"delivery"`
}

func Load(path string, envPrefix string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	// Set defaults
	v.SetDefault("kafka.minBytes", 1e4)
	v.SetDefault("kafka.maxBytes", 1e6)
	v.SetDefault("kafka.maxWait", time.Second)
	v.SetDefault("kafka.commitInterval", time.Second*5)
	v.SetDefault("delivery.mode", "mongo")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("config: unmarshal: %w", err)
	}

	// Validate required fields
	if len(cfg.Kafka.Brokers) == 0 || cfg.Kafka.Topic == "" || cfg.Kafka.GroupID == "" {
		return Config{}, errors.New("config: kafka brokers, groupId and topic are required")
	}
	if cfg.Storage.Mongo.URI == "" || cfg.Storage.Mongo.Database == "" {
		return Config{}, errors.New("config: storage.mongo uri and database are required")
	}
	if cfg.Kafka.DLQTopic == "" {
		return Config{}, errors.New("config: kafka.dlqTopic is required")
	}

	return cfg, nil
}

func (config Config) ReaderConfig() kafka.ReaderConfig {
	return kafka.ReaderConfig{
		Brokers:        config.Kafka.Brokers,
		GroupID:        config.Kafka.GroupID,
		Topic:          config.Kafka.Topic,
		MinBytes:       config.Kafka.MinBytes,
		MaxBytes:       config.Kafka.MaxBytes,
		MaxWait:        config.Kafka.MaxWait,
		CommitInterval: config.Kafka.CommitInterval,
	}
}

