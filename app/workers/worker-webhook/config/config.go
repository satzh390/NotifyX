package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/notifyx/workerx/config"
	"github.com/spf13/viper"
)

type Config struct {
	Base    config.Config
	Webhook struct {
		Provider struct {
			Type string `mapstructure:"type"` // "http" or "mock" for testing
			HTTP struct {
				Timeout time.Duration `mapstructure:"timeout"` // Request timeout
			} `mapstructure:"http"`
		} `mapstructure:"provider"`
	} `mapstructure:"webhook"`
}

func Load(path string) (Config, error) {
	base, err := config.Load(path, "NOTIFYX_WORKER_WEBHOOK")
	if err != nil {
		return Config{}, err
	}

	// Load Webhook-specific config
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("NOTIFYX_WORKER_WEBHOOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var webhookConfig struct {
		Webhook struct {
			Provider struct {
				Type string `mapstructure:"type"`
				HTTP struct {
					Timeout time.Duration `mapstructure:"timeout"`
				} `mapstructure:"http"`
			} `mapstructure:"provider"`
		} `mapstructure:"webhook"`
	}
	if err := v.Unmarshal(&webhookConfig); err != nil {
		return Config{}, err
	}

	return Config{
		Base:    base,
		Webhook: webhookConfig.Webhook,
	}, nil
}

