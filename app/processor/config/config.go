package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/notifyx/httpx"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

type Config struct {
	Kafka struct {
		Brokers        []string      `mapstructure:"brokers"`
		GroupID        string        `mapstructure:"groupId"`
		InputTopic     string        `mapstructure:"inputTopic"`
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
	Worker struct {
		Topics map[string]string `mapstructure:"topics"`
	} `mapstructure:"worker"`
	Cache struct {
		Redis struct {
			Enabled bool          `mapstructure:"enabled"`
			Addr    string        `mapstructure:"addr"`
			DB      int           `mapstructure:"db"`
			Pass    string        `mapstructure:"pass"`
			TTL     time.Duration `mapstructure:"ttl"`
		} `mapstructure:"redis"`
	} `mapstructure:"cache"`
}

func Load(path string) (Config, error) {
	// Read config file and expand environment variables
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("config: read file: %w", err)
	}

	// Expand environment variables in config (supports ${VAR} and ${VAR:-default})
	expandedConfig := httpx.ExpandEnvWithDefaults(string(configBytes))

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("NOTIFYX_PROCESSOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	v.SetDefault("kafka.minBytes", 1e4)
	v.SetDefault("kafka.maxBytes", 1e6)
	v.SetDefault("kafka.maxWait", time.Second)
	v.SetDefault("kafka.commitInterval", time.Second*5)
	v.SetDefault("cache.redis.ttl", time.Minute*5)

	// Read from expanded config string
	if err := v.ReadConfig(strings.NewReader(expandedConfig)); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("config: unmarshal: %w", err)
	}

	if len(cfg.Kafka.Brokers) == 0 || cfg.Kafka.InputTopic == "" || cfg.Kafka.GroupID == "" {
		return Config{}, errors.New("config: kafka brokers, groupId and inputTopic are required")
	}
	if cfg.Storage.Mongo.URI == "" || cfg.Storage.Mongo.Database == "" {
		return Config{}, errors.New("config: storage.mongo uri and database are required")
	}
	if len(cfg.Worker.Topics) == 0 {
		return Config{}, errors.New("config: worker.topics must declare at least one channel topic")
	}

	// Validate DLQ requirement
	if cfg.Kafka.DLQTopic == "" {
		return Config{}, errors.New("config: kafka.dlqTopic is required")
	}

	// Validate worker topics by channel name to avoid typos
	for channel, topic := range cfg.Worker.Topics {
		if channel == "" || topic == "" {
			return Config{}, fmt.Errorf("config: worker topics cannot have empty channel/topic")
		}
	}

	return cfg, nil
}

func (config Config) ReaderConfig() kafka.ReaderConfig {
	return kafka.ReaderConfig{
		Brokers:        config.Kafka.Brokers,
		GroupID:        config.Kafka.GroupID,
		Topic:          config.Kafka.InputTopic,
		MinBytes:       config.Kafka.MinBytes,
		MaxBytes:       config.Kafka.MaxBytes,
		MaxWait:        config.Kafka.MaxWait,
		CommitInterval: config.Kafka.CommitInterval,
	}
}
