package config

import (
	"veemon/pkg/logger"
)

func NewLogger(cfg *Config) (*logger.Logger, error) {
	return logger.New(logger.Config{
		Level:       cfg.LogLevel,
		Format:      cfg.LogFormat,
		Environment: cfg.Environment,
		ServiceName: cfg.ServiceName,
	})
}
