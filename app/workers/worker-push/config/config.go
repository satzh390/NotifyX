package config

import (
	"github.com/notifyx/workerx/config"
)

type Config struct {
	Base config.Config
}

func Load(path string) (Config, error) {
	base, err := config.Load(path, "NOTIFYX_WORKER_PUSH")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Base: base,
	}, nil
}

