package webhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDispatchNoOpWhenDisabled(t *testing.T) {
	d := NewDispatcher("")
	err := d.Dispatch(context.Background(), "test", map[string]string{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDispatchMessageCreated(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := NewDispatcher(srv.URL)
	d.DispatchMessageCreated(context.Background(), "m1", "c1", "u1", "hello", time.Now())
	if hits.Load() != 1 {
		t.Fatalf("expected 1 webhook hit, got %d", hits.Load())
	}
}
