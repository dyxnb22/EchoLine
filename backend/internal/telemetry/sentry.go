package telemetry

import (
	"log/slog"
	"os"
)

// InitSentry initializes Sentry when SENTRY_DSN is set (stub — logs only).
func InitSentry(logger *slog.Logger) {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return
	}
	if logger != nil {
		logger.Info("sentry stub initialized", "dsn_set", true)
	}
}
