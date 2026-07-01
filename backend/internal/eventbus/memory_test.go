package eventbus

import (
	"context"
	"testing"
)

func TestMemoryPublisherPublish(t *testing.T) {
	pub := NewMemoryPublisher(4)
	if err := pub.Publish(context.Background(), "message.created", Event{Type: "message.created"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	select {
	case evt := <-pub.C():
		if evt.Type != "message.created" {
			t.Fatalf("type = %q", evt.Type)
		}
	default:
		t.Fatal("expected event on channel")
	}
}
