package realtime

import (
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCheckWSOriginAllowsLocalhostDev(t *testing.T) {
	t.Setenv("WS_ALLOWED_ORIGINS", "")
	wsAllowedOriginsOnce.Do(func() {})
	wsAllowedOrigins = nil
	wsAllowedOriginsOnce = sync.Once{}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	if !checkWSOrigin(req) {
		t.Fatal("expected localhost dev origin to be allowed")
	}
}

func TestCheckWSOriginRejectsUnknown(t *testing.T) {
	t.Setenv("WS_ALLOWED_ORIGINS", "https://app.example.com")
	wsAllowedOrigins = nil
	wsAllowedOriginsOnce = sync.Once{}

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	if checkWSOrigin(req) {
		t.Fatal("expected unknown origin to be rejected")
	}
}
