package worker

import (
	"context"
	"testing"
)

func TestMessageCreatedHandlerIdempotent(t *testing.T) {
	h := NewMessageCreatedHandler(nil, nil)
	payload := []byte(`{"id":"00000000-0000-0000-0000-000000000001","conversation_id":"00000000-0000-0000-0000-000000000002","sender_id":"00000000-0000-0000-0000-000000000003","seq":1,"type":"text","body":"hi","created_at":"2026-07-01T00:00:00Z"}`)
	if err := h.Handle(context.Background(), payload); err != nil {
		t.Fatalf("first handle: %v", err)
	}
	if err := h.Handle(context.Background(), payload); err != nil {
		t.Fatalf("second handle: %v", err)
	}
}
