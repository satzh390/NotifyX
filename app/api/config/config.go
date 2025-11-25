package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/notifyx/httpx"
	"github.com/spf13/viper"
)

type Config struct {
	HTTP struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"http"`
	OAuth struct {
		Issuer    string   `mapstructure:"issuer"`
		JWKS      string   `mapstructure:"jwks"`
		Audiences []string `mapstructure:"audiences"`
	} `mapstructure:"oauth"`
	Storage struct {
		Mongo struct {
			URI      string `mapstructure:"uri"`
			Database string `mapstructure:"database"`
		} `mapstructure:"mongo"`
	} `mapstructure:"storage"`
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
	v.SetEnvPrefix("NOTIFYX_API")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	// Read from expanded config string
	if err := v.ReadConfig(strings.NewReader(expandedConfig)); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("config: unmarshal: %w", err)
	}

	if cfg.OAuth.Issuer == "" || cfg.OAuth.JWKS == "" {
		return Config{}, errors.New("config: oauth issuer and jwks are required")
	}
	if cfg.Storage.Mongo.URI == "" || cfg.Storage.Mongo.Database == "" {
		return Config{}, errors.New("config: storage.mongo uri and database are required")
	}

	return cfg, nil
}
