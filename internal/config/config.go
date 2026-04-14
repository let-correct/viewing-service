package config

import "github.com/caarlos0/env/v11"

type ViewingWorkerConfig struct {
	LogLevel     string `env:"LOG_LEVEL"           envDefault:"INFO"`
	EventBusName string `env:"EVENT_BUS_NAME,required"`
}

func LoadViewingWorker() (ViewingWorkerConfig, error) {
	var cfg ViewingWorkerConfig
	return cfg, env.Parse(&cfg)
}
