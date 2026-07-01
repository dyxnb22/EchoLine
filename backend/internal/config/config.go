package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string
	RedisAddr   string
	JWTSecret   string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:      envOrDefault("APP_ENV", "development"),
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisAddr:   strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		JWTSecret:   strings.TrimSpace(os.Getenv("JWT_SECRET")),
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
