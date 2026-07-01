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
