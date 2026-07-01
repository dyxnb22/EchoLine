package tests

import (
	"os"
	"testing"
)

func TestIntegrationSkippedWithoutEnv(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" || os.Getenv("DATABASE_URL") == "" {
		t.Skip("integration tests require RUN_INTEGRATION=1 and DATABASE_URL")
	}
}

func TestIntegrationHealthPlaceholder(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" || os.Getenv("DATABASE_URL") == "" {
		t.Skip("integration tests require RUN_INTEGRATION=1 and DATABASE_URL")
	}
	// Full register/login/send flow runs via scripts/smoke-api-full.sh when server is up.
	t.Log("integration placeholder: use make smoke-full with running API")
}
