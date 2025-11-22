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
	viperInstance := viper.New()
	viperInstance.SetConfigFile(path)
	viperInstance.SetConfigType("yaml")
	viperInstance.SetEnvPrefix("NOTIFYX_API")
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viperInstance.AutomaticEnv()

	if err := viperInstance.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	var config Config
	if err := viperInstance.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("config: unmarshal: %w", err)
	}

	if config.OAuth.Issuer == "" || config.OAuth.JWKS == "" {
		return Config{}, errors.New("config: oauth issuer and jwks are required")
	}
	if config.Storage.Mongo.URI == "" || config.Storage.Mongo.Database == "" {
		return Config{}, errors.New("config: storage.mongo uri and database are required")
	}

	return config, nil
}

