package config

import (
	"fmt"
	"strings"

	"github.com/notifyx/workerx/config"
	"github.com/spf13/viper"
)

type Config struct {
	Base config.Config
	Push struct {
		Provider struct {
			Type string `mapstructure:"type"` // "firebase", "apns", or other providers
			Firebase struct {
				ProjectID string `mapstructure:"projectId"`
				Credential string `mapstructure:"credential"` // Path to service account JSON
			} `mapstructure:"firebase"`
			APNS struct {
				KeyID      string `mapstructure:"keyId"`
				TeamID     string `mapstructure:"teamId"`
				BundleID   string `mapstructure:"bundleId"`
				KeyPath    string `mapstructure:"keyPath"`
				Production bool   `mapstructure:"production"`
			} `mapstructure:"apns"`
		} `mapstructure:"provider"`
	} `mapstructure:"push"`
}

func Load(path string) (Config, error) {
	base, err := config.Load(path, "NOTIFYX_WORKER_PUSH")
	if err != nil {
		return Config{}, err
	}

	// Load Push-specific config
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("NOTIFYX_WORKER_PUSH")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var pushConfig struct {
		Push struct {
			Provider struct {
				Type string `mapstructure:"type"`
				Firebase struct {
					ProjectID  string `mapstructure:"projectId"`
					Credential string `mapstructure:"credential"`
				} `mapstructure:"firebase"`
				APNS struct {
					KeyID      string `mapstructure:"keyId"`
					TeamID     string `mapstructure:"teamId"`
					BundleID   string `mapstructure:"bundleId"`
					KeyPath    string `mapstructure:"keyPath"`
					Production bool   `mapstructure:"production"`
				} `mapstructure:"apns"`
			} `mapstructure:"provider"`
		} `mapstructure:"push"`
	}
	if err := v.Unmarshal(&pushConfig); err != nil {
		return Config{}, err
	}

	return Config{
		Base: base,
		Push: pushConfig.Push,
	}, nil
}

