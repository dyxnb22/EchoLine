package realtime

import (
	"encoding/json"
	"testing"
)

func TestMarshalUnmarshalEnvelope(t *testing.T) {
	raw, err := marshalEnvelope("message.created", "req-1", MessageCreatedPayload{
		ID:             "m1",
		ConversationID: "c1",
		SenderID:       "u1",
		Seq:            1,
		Type:           "text",
		Body:           "hello",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	env, err := parseEnvelope(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if env.Type != "message.created" || env.RequestID != "req-1" {
		t.Fatalf("unexpected envelope: %+v", env)
	}

	var payload MessageCreatedPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Body != "hello" {
		t.Fatalf("body = %q", payload.Body)
	}
}

func TestMarshalErrorEnvelope(t *testing.T) {
	raw, err := marshalError("req-2", "invalid_request", "bad payload")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	env, err := parseEnvelope(raw)
	if err != nil || env.Type != "error" {
		t.Fatalf("parse error envelope: %v", err)
	}
}

func TestMessageEditedPayloadIncludesIDAlias(t *testing.T) {
	raw, err := marshalEnvelope("message.edited", "", MessageEditedPayload{
		ID:             "m1",
		MessageID:      "m1",
		ConversationID: "c1",
		Body:           "updated",
		UpdatedAt:      "2026-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var doc struct {
		Payload struct {
			ID        string `json:"id"`
			MessageID string `json:"message_id"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc.Payload.ID != "m1" || doc.Payload.MessageID != "m1" {
		t.Fatalf("payload ids = %q %q, want m1 m1", doc.Payload.ID, doc.Payload.MessageID)
	}
}

func TestMessageRecalledPayloadIncludesIDAlias(t *testing.T) {
	raw, err := marshalEnvelope("message.recalled", "", MessageRecalledPayload{
		ID:             "m2",
		MessageID:      "m2",
		ConversationID: "c1",
		UpdatedAt:      "2026-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var doc struct {
		Payload struct {
			ID        string `json:"id"`
			MessageID string `json:"message_id"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc.Payload.ID != "m2" || doc.Payload.MessageID != "m2" {
		t.Fatalf("payload ids = %q %q, want m2 m2", doc.Payload.ID, doc.Payload.MessageID)
	}
}
