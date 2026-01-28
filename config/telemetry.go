package config

import (
	"context"

	"go-grst-boilerplate/pkg/telemetry"
)

func NewTelemetry(ctx context.Context, cfg *Config) (*telemetry.Telemetry, error) {
	return telemetry.New(ctx, telemetry.Config{
		ServiceName:  cfg.OTelServiceName,
		Environment:  cfg.Environment,
		Endpoint:     cfg.OTelEndpoint,
		ExporterType: cfg.OTelExporterType,
		Enabled:      cfg.OTelEnabled,
	})
}
