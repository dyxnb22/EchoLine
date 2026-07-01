package rate_limit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubLimiter struct {
	allow bool
}

func (s stubLimiter) Allow(_ context.Context, _ string, _ int64, _ time.Duration) (bool, error) {
	return s.allow, nil
}

func TestMiddlewareBlocks(t *testing.T) {
	called := false
	h := Middleware(stubLimiter{allow: false}, "test", 1, time.Minute, IPKey)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d", rec.Code)
	}
	if called {
		t.Fatal("handler should not run")
	}
}

func TestMiddlewareAllows(t *testing.T) {
	called := false
	h := Middleware(stubLimiter{allow: true}, "test", 1, time.Minute, IPKey)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("handler should run")
	}
}
