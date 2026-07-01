package realtime

import (
	"testing"

	"github.com/google/uuid"
)

func TestHubRegisterUnregister(t *testing.T) {
	hub := NewHub()
	userID := uuid.New()
	conn := &Connection{UserID: userID, DeviceID: "d1"}

	hub.Register(userID, "d1", conn)
	if hub.ConnectionCount() != 1 {
		t.Fatalf("count = %d, want 1", hub.ConnectionCount())
	}

	hub.Unregister(userID, "d1")
	if hub.ConnectionCount() != 0 {
		t.Fatalf("count = %d, want 0", hub.ConnectionCount())
	}
}
