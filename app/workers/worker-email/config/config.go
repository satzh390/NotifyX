package config

import (
	"fmt"
	"strings"

	"github.com/notifyx/workerx/config"
	"github.com/spf13/viper"
)

type Config struct {
	Base  config.Config
	Email struct {
		Provider struct {
			Type string `mapstructure:"type"` // "smtp" or other providers
			SMTP struct {
				Host     string `mapstructure:"host"`
				Port     string `mapstructure:"port"`
				Username string `mapstructure:"username"`
				Password string `mapstructure:"password"`
				From     string `mapstructure:"from"`
			} `mapstructure:"smtp"`
		} `mapstructure:"provider"`
	} `mapstructure:"email"`
}

func Load(path string) (Config, error) {
	base, err := config.Load(path, "NOTIFYX_WORKER_EMAIL")
	if err != nil {
		return Config{}, err
	}

	// Load Email-specific config
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("NOTIFYX_WORKER_EMAIL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var emailConfig struct {
		Email struct {
			Provider struct {
				Type string `mapstructure:"type"`
				SMTP struct {
					Host     string `mapstructure:"host"`
					Port     string `mapstructure:"port"`
					Username string `mapstructure:"username"`
					Password string `mapstructure:"password"`
					From     string `mapstructure:"from"`
				} `mapstructure:"smtp"`
			} `mapstructure:"provider"`
		} `mapstructure:"email"`
	}
	if err := v.Unmarshal(&emailConfig); err != nil {
		return Config{}, err
	}

	return Config{
		Base:  base,
		Email: emailConfig.Email,
	}, nil
}

