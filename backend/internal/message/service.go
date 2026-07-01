package message

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/echoline/echoline/backend/internal/conversation"
)

// Broadcaster pushes realtime events to online users.
type Broadcaster interface {
	BroadcastMessageCreated(ctx context.Context, convID uuid.UUID, msg *Message, excludeSender bool, senderID uuid.UUID) error
}

// Service coordinates message persistence and realtime fanout.
type Service struct {
	repo          *Repository
	conversations *conversation.Repository
	broadcaster   Broadcaster
}

// SetBroadcaster attaches a realtime broadcaster after construction.
func (s *Service) SetBroadcaster(b Broadcaster) {
	s.broadcaster = b
}

// NewService creates a message service.
func NewService(repo *Repository, conversations *conversation.Repository, broadcaster Broadcaster) *Service {
	return &Service{
		repo:          repo,
		conversations: conversations,
		broadcaster:   broadcaster,
	}
}

// Send persists a message and notifies online members.
func (s *Service) Send(ctx context.Context, convID, senderID uuid.UUID, clientMsgID string, msgType Type, body string) (*Message, error) {
	member, err := s.conversations.IsMember(ctx, convID, senderID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !member {
		return nil, conversation.ErrNotMember
	}

	msg, err := s.repo.Create(ctx, convID, senderID, clientMsgID, msgType, body)
	if err != nil {
		return nil, err
	}

	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastMessageCreated(ctx, convID, msg, true, senderID)
	}
	return msg, nil
}

// ListSince returns messages with seq greater than afterSeq.
func (s *Service) ListSince(ctx context.Context, convID uuid.UUID, afterSeq int64, limit int) ([]Message, error) {
	return s.repo.ListSince(ctx, convID, afterSeq, limit)
}

// List returns paginated history.
func (s *Service) List(ctx context.Context, convID uuid.UUID, beforeSeq *int64, limit int) ([]Message, error) {
	return s.repo.List(ctx, convID, beforeSeq, limit)
}

// ToCreatedPayload converts a message for WS/REST responses.
func ToCreatedPayload(msg *Message) map[string]any {
	return map[string]any{
		"id":              msg.ID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"client_msg_id":   msg.ClientMsgID,
		"seq":             msg.Seq,
		"type":            msg.Type,
		"body":            msg.Body,
		"status":          msg.Status,
		"created_at":      msg.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      msg.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
