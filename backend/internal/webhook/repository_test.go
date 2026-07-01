package webhook

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPersistingDispatcherNoOpWhenDisabled(t *testing.T) {
	p := NewPersistingDispatcher(NewDispatcher(""), nil)
	p.DispatchMessageCreated(context.Background(), "m", "c", "u", "hi", time.Now())
}

func TestRetryWorkerNilSafe(t *testing.T) {
	w := NewRetryWorker(nil, nil)
	n, err := w.RunOnce(context.Background())
	if err != nil || n != 0 {
		t.Fatalf("got %d %v", n, err)
	}
}

func TestDeliveryStruct(t *testing.T) {
	d := Delivery{ID: uuid.New(), EventType: "message.created", Status: "pending"}
	if d.EventType != "message.created" {
		t.Fatal("unexpected type")
	}
}
