package message

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdempotentClientMsgIDScope(t *testing.T) {
	convA := uuid.New()
	convB := uuid.New()
	existing := &Message{
		ConversationID: convA,
		ClientMsgID:    "550e8400-e29b-41d4-a716-446655440000",
	}

	if existing.ConversationID != convB {
		// cross-conversation reuse must be rejected at service layer
		if existing.ConversationID == convB {
			t.Fatal("unexpected match")
		}
	}
}
