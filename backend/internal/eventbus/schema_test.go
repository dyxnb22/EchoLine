package eventbus

import (
	"testing"
	"time"
)

func TestMessageCreatedRoundTrip(t *testing.T) {
	evt := MessageCreatedEvent{
		ID:             "m1",
		ConversationID: "c1",
		SenderID:       "u1",
		Seq:            1,
		Type:           "text",
		Body:           "hello",
		CreatedAt:      time.Unix(1, 0).UTC(),
	}
	raw, err := EncodeMessageCreated(evt)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := DecodeMessageCreated(raw)
	if err != nil || decoded.Body != "hello" {
		t.Fatalf("decode: %v %+v", err, decoded)
	}
}
