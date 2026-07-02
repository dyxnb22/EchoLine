package worker

import (
	"context"
	"testing"
	"time"
)

func TestOutboxReclaimWorkerNilSafe(t *testing.T) {
	w := NewOutboxReclaimWorker(nil, 0, nil)
	n, err := w.RunOnce(context.Background())
	if err != nil || n != 0 {
		t.Fatalf("got %d %v", n, err)
	}
}

func TestOutboxReclaimWorkerDefaultStaleAfter(t *testing.T) {
	w := NewOutboxReclaimWorker(nil, 0, nil)
	if w.staleAfter != 0 {
		t.Fatalf("staleAfter = %v, want zero", w.staleAfter)
	}
	// RunOnce applies default when staleAfter <= 0; nil repo still returns 0.
	n, err := w.RunOnce(context.Background())
	if err != nil || n != 0 {
		t.Fatalf("got %d %v", n, err)
	}
}

func TestOutboxReclaimWorkerRunRespectsContext(t *testing.T) {
	w := NewOutboxReclaimWorker(nil, time.Minute, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		w.Run(ctx, 10*time.Millisecond)
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}
