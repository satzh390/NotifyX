package config

import (
	"fmt"
	"strings"

	"github.com/notifyx/workerx/config"
	"github.com/spf13/viper"
)

type Config struct {
	Base config.Config
	SMS  struct {
		Provider struct {
			Type string `mapstructure:"type"` // "sns" or other providers
			SNS  struct {
				Region    string `mapstructure:"region"`
				AccessKey string `mapstructure:"accessKey"`
				SecretKey string `mapstructure:"secretKey"`
			} `mapstructure:"sns"`
		} `mapstructure:"provider"`
	} `mapstructure:"sms"`
}

func Load(path string) (Config, error) {
	base, err := config.Load(path, "NOTIFYX_WORKER_SMS")
	if err != nil {
		return Config{}, err
	}

	// Load SMS-specific config
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("NOTIFYX_WORKER_SMS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var smsConfig struct {
		SMS struct {
			Provider struct {
				Type string `mapstructure:"type"`
				SNS  struct {
					Region    string `mapstructure:"region"`
					AccessKey string `mapstructure:"accessKey"`
					SecretKey string `mapstructure:"secretKey"`
				} `mapstructure:"sns"`
			} `mapstructure:"provider"`
		} `mapstructure:"sms"`
	}
	if err := v.Unmarshal(&smsConfig); err != nil {
		return Config{}, err
	}

	return Config{
		Base: base,
		SMS:  smsConfig.SMS,
	}, nil
}
