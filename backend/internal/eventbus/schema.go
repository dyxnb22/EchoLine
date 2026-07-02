package eventbus

import (
	"encoding/json"
	"time"
)

const (
	TopicMessageCreated = "message.created"
	TopicMessageEdited  = "message.edited"
	TopicMessageRecalled = "message.recalled"
)

// MessageCreatedEvent is published after a message is persisted.
type MessageCreatedEvent struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Seq            int64     `json:"seq"`
	Type           string    `json:"type"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

// EncodeMessageCreated serializes a message.created event.
func EncodeMessageCreated(evt MessageCreatedEvent) ([]byte, error) {
	return json.Marshal(evt)
}

// DecodeMessageCreated parses a message.created event payload.
func DecodeMessageCreated(data []byte) (MessageCreatedEvent, error) {
	var evt MessageCreatedEvent
	err := json.Unmarshal(data, &evt)
	return evt, err
}
