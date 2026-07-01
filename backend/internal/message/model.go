package message

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("message not found")
	ErrDuplicateClientID = errors.New("duplicate client message id")
)

// Type identifies message payload kind.
type Type string

const (
	TypeText   Type = "text"
	TypeImage  Type = "image"
	TypeFile   Type = "file"
	TypeSystem Type = "system"
)

// Message is a persisted chat message.
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	ClientMsgID    string
	Seq            int64
	Type           Type
	Body           string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
