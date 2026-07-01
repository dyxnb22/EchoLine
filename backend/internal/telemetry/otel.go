package telemetry

import (
	"context"
	"log/slog"
	"os"
)

// InitOTel initializes OpenTelemetry when OTEL_EXPORTER_OTLP_ENDPOINT is set (stub).
func InitOTel(ctx context.Context, logger *slog.Logger) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return func(context.Context) error { return nil }
	}
	if logger != nil {
		logger.Info("otel stub initialized", "endpoint", endpoint)
	}
	return func(context.Context) error {
		if logger != nil {
			logger.Info("otel stub shutdown")
		}
		return nil
	}
}
