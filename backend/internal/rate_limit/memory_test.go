package rate_limit

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMemoryLimiterAllowsWithinWindow(t *testing.T) {
	l := NewMemoryLimiter()
	ctx := context.Background()
	ok, err := l.Allow(ctx, "k", 2, time.Minute)
	if err != nil || !ok {
		t.Fatalf("first allow = %v err %v", ok, err)
	}
	ok, err = l.Allow(ctx, "k", 2, time.Minute)
	if err != nil || !ok {
		t.Fatalf("second allow = %v err %v", ok, err)
	}
	ok, err = l.Allow(ctx, "k", 2, time.Minute)
	if err != nil || ok {
		t.Fatalf("third allow = %v err %v, want blocked", ok, err)
	}
}

func TestIPKeyUsesRemoteAddrByDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.5:12345"
	if got := IPKey(req); got != "203.0.113.5" {
		t.Fatalf("IPKey = %q, want 203.0.113.5", got)
	}
}
