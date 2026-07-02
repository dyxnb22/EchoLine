package outbox

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventFields(t *testing.T) {
	e := Event{
		ID:        uuid.New(),
		Topic:     "message.created",
		Payload:   []byte(`{"id":"m1"}`),
		Attempts:  0,
		CreatedAt: time.Now().UTC(),
	}
	if e.Topic == "" {
		t.Fatal("expected topic")
	}
}
