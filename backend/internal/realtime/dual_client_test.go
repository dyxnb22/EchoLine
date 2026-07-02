package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// TestDualClientWSReceive verifies two clients on the same user receive broadcasts.
func TestDualClientWSReceive(t *testing.T) {
	userID := uuid.New()
	hub := NewHub()
	connA := &Connection{UserID: userID, DeviceID: "a", send: make(chan []byte, 256)}
	connB := &Connection{UserID: userID, DeviceID: "b", send: make(chan []byte, 256)}
	hub.Register(userID, "a", connA)
	hub.Register(userID, "b", connB)

	payload, _ := json.Marshal(map[string]any{"type": "message.created", "payload": map[string]any{"seq": 1}})
	hub.PushToUser(context.Background(), userID, payload)

	select {
	case msg := <-connA.send:
		if !strings.Contains(string(msg), "message.created") {
			t.Fatalf("connA unexpected: %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("connA timeout")
	}
	select {
	case msg := <-connB.send:
		if !strings.Contains(string(msg), "message.created") {
			t.Fatalf("connB unexpected: %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("connB timeout")
	}
}

// TestWSAuthRejectsMissingToken ensures unauthenticated upgrades fail.
func TestWSAuthRejectsMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	w := httptest.NewRecorder()
	s := &Server{auth: nil, hub: NewHub()}
	s.HandleWS(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	_ = websocket.IsUnexpectedCloseError
}
