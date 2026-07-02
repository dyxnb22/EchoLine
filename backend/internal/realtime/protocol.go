package realtime

import (
	"encoding/json"
	"fmt"
)

// Envelope is the standard WebSocket message wrapper.
type Envelope struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// ErrorPayload is returned for WS errors.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PingPayload is a client heartbeat.
type PingPayload struct {
	TS int64 `json:"ts"`
}

// PongPayload is a server heartbeat response.
type PongPayload struct {
	TS int64 `json:"ts"`
}

// MessageSendPayload creates a message over WebSocket.
type MessageSendPayload struct {
	ConversationID      string `json:"conversation_id"`
	ClientMsgID         string `json:"client_msg_id"`
	Type                string `json:"type"`
	Body                string `json:"body"`
	AttachmentObjectKey string `json:"attachment_object_key,omitempty"`
}

// MessageCreatedPayload is pushed when a message is persisted.
type MessageCreatedPayload struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	ClientMsgID    string `json:"client_msg_id"`
	Seq            int64  `json:"seq"`
	Type           string `json:"type"`
	Body           string `json:"body"`
	CreatedAt      string `json:"created_at"`
}

// MessageAckPayload acknowledges delivery/read state.
type MessageAckPayload struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Seq            int64  `json:"seq"`
	Status         string `json:"status"`
}

// TypingPayload is sent by a client to indicate typing activity.
type TypingPayload struct {
	ConversationID string `json:"conversation_id"`
}

// TypingIndicatorPayload is broadcast to conversation members.
type TypingIndicatorPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

// TypingStopPayload is sent by a client to signal end of typing.
type TypingStopPayload struct {
	ConversationID string `json:"conversation_id"`
}

// MessageEditedPayload is broadcast when a message body is updated.
type MessageEditedPayload struct {
	ID             string `json:"id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Body           string `json:"body"`
	UpdatedAt      string `json:"updated_at"`
}

// MessageRecalledPayload is broadcast when a message is recalled.
type MessageRecalledPayload struct {
	ID             string `json:"id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	UpdatedAt      string `json:"updated_at"`
}

func marshalEnvelope(typ, requestID string, payload any) ([]byte, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	return json.Marshal(Envelope{Type: typ, RequestID: requestID, Payload: raw})
}

func marshalError(requestID, code, message string) ([]byte, error) {
	payload, err := json.Marshal(ErrorPayload{Code: code, Message: message})
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: "error", RequestID: requestID, Payload: payload})
}

func parseEnvelope(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, fmt.Errorf("invalid envelope: %w", err)
	}
	if env.Type == "" {
		return Envelope{}, fmt.Errorf("missing type")
	}
	return env, nil
}
