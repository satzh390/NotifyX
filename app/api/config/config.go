package config

import (
	"errors"
	"fmt"
	"strings"

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
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("NOTIFYX_API")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
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
