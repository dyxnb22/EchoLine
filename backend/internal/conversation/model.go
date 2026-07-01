package conversation

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound        = errors.New("conversation not found")
	ErrNotMember       = errors.New("not a conversation member")
	ErrInvalidType     = errors.New("invalid conversation type")
	ErrDuplicateDirect = errors.New("direct conversation already exists")
)

// Type identifies conversation kind.
type Type string

const (
	TypeDirect  Type = "direct"
	TypeGroup   Type = "group"
	TypeChannel Type = "channel"
)

// Role identifies member permissions.
type Role string

const (
	RoleOwner      Role = "owner"
	RoleAdmin      Role = "admin"
	RoleMember     Role = "member"
	RoleSubscriber Role = "subscriber"
)

// Conversation is a chat container.
type Conversation struct {
	ID            uuid.UUID
	Type          Type
	Title         string
	LatestSeq     int64
	LastMessageID *uuid.UUID
	CreatedBy     *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Member links a user to a conversation.
type Member struct {
	ConversationID uuid.UUID
	UserID         uuid.UUID
	Role           Role
	LastReadSeq    int64
	JoinedAt       time.Time
}
