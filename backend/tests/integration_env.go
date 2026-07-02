package tests

import (
	"os"
	"testing"
)

const integrationJWTSecret = "integration-test-secret-do-not-use-in-production"

func ensureIntegrationEnv(t *testing.T) {
	t.Helper()
	if os.Getenv("JWT_SECRET") == "" {
		t.Setenv("JWT_SECRET", integrationJWTSecret)
	}
}
