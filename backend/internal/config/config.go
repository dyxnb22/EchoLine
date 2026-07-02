package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	AppEnv        string
	HTTPAddr      string
	DatabaseURL   string
	RedisAddr     string
	KafkaBrokers  string
	JWTSecret     string
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3Bucket      string
	OpenSearchURL string
	WebhookURL    string
	AdminUserIDs  string
	SentryDSN     string
	GraphiQL      bool
}

// Load reads configuration from environment variables and validates required fields.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:        envOrDefault("APP_ENV", "development"),
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL:   strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisAddr:     strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		KafkaBrokers:  strings.TrimSpace(os.Getenv("KAFKA_BROKERS")),
		JWTSecret:     strings.TrimSpace(os.Getenv("JWT_SECRET")),
		S3Endpoint:    strings.TrimSpace(os.Getenv("S3_ENDPOINT")),
		S3AccessKey:   strings.TrimSpace(os.Getenv("S3_ACCESS_KEY")),
		S3SecretKey:   strings.TrimSpace(os.Getenv("S3_SECRET_KEY")),
		S3Bucket:      envOrDefault("S3_BUCKET", "echoline"),
		OpenSearchURL: strings.TrimSpace(os.Getenv("OPENSEARCH_URL")),
		WebhookURL:    strings.TrimSpace(os.Getenv("WEBHOOK_URL")),
		AdminUserIDs:  strings.TrimSpace(os.Getenv("ADMIN_USER_IDS")),
		SentryDSN:     strings.TrimSpace(os.Getenv("SENTRY_DSN")),
		GraphiQL:      strings.EqualFold(os.Getenv("GRAPHIQL"), "true"),
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
